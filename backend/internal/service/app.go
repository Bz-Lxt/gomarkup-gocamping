package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gocamping/internal/dem"
	"gocamping/internal/eta"
	"gocamping/internal/geo"
	"gocamping/internal/httpx"
	"gocamping/internal/model"
	"gocamping/internal/notify"
	"gocamping/internal/repo"
	"gocamping/internal/risk"
	"gocamping/internal/timeutil"
	"gocamping/internal/track"
	"gocamping/internal/ws"

	"github.com/redis/go-redis/v9"
)

type App struct {
	Users   *repo.UserRepo
	Routes  *repo.RouteRepo
	Teams   *repo.TeamRepo
	Trips   *repo.TripRepo
	Tracks  *repo.TrackRepo
	SOS     *repo.SOSRepo
	Risks   *repo.RiskRepo
	DEM     dem.Provider
	Notify  notify.Provider
	Hub     *ws.Hub
	Grid    *geo.Grid
	Redis   *redis.Client
	PosHz   *ws.Throttle
	RiskHz  *risk.Throttle
}

func (a *App) SaveRoute(ctx context.Context, owner int64, in *model.RouteBook) (*model.RouteBook, error) {
	if strings.TrimSpace(in.Title) == "" {
		return nil, httpx.Validation("路书标题必填")
	}
	if in.Visibility == "" {
		in.Visibility = "private"
	}
	if in.Status == "" {
		in.Status = "active"
	}
	path := pathOf(in.Waypoints)
	in.DistanceM = geo.PathDistance(path)
	in.GeometryHash = dem.GeometryHash(path)
	in.OwnerID = owner
	if in.Version == 0 {
		in.Version = 1
	}
	if in.ID == 0 {
		if err := a.Routes.Create(ctx, in); err != nil {
			return nil, httpx.Internal("保存路书失败")
		}
	} else {
		old, err := a.Routes.Get(ctx, in.ID)
		if err != nil || old == nil {
			return nil, httpx.NotFound("路书不存在")
		}
		if old.OwnerID != owner {
			return nil, httpx.Forbidden("只能编辑自己的路书")
		}
		if err := a.Routes.Update(ctx, in); err != nil {
			return nil, httpx.Internal("更新路书失败")
		}
	}
	if err := a.Routes.ReplaceWaypoints(ctx, in.ID, in.Waypoints); err != nil {
		return nil, httpx.Internal("保存点位失败")
	}
	return a.Routes.Get(ctx, in.ID)
}

func (a *App) Elevation(ctx context.Context, routeID int64) (*dem.Profile, error) {
	rb, err := a.Routes.Get(ctx, routeID)
	if err != nil || rb == nil {
		return nil, httpx.NotFound("路书不存在")
	}
	path := pathOf(rb.Waypoints)
	hash := dem.GeometryHash(path)
	if raw, _, err := a.Routes.GetElevation(ctx, hash); err == nil && raw != nil {
		var p dem.Profile
		if json.Unmarshal(raw, &p) == nil {
			return &p, nil
		}
	}
	prof, err := dem.BuildProfile(ctx, a.DEM, path)
	if err != nil {
		return nil, httpx.Internal(err.Error())
	}
	blob, _ := json.Marshal(prof)
	_ = a.Routes.PutElevation(ctx, hash, a.DEM.Name(), blob)
	_ = a.Routes.Update(ctx, &model.RouteBook{
		ID: rb.ID, Title: rb.Title, Description: rb.Description, Visibility: rb.Visibility,
		DistanceM: prof.DistanceM, AscentM: prof.AscentM, GeometryHash: hash, Status: rb.Status,
	})
	return prof, nil
}

func (a *App) CreateTeam(ctx context.Context, leader int64, name string, routeID *int64) (*model.Team, error) {
	if strings.TrimSpace(name) == "" {
		return nil, httpx.Validation("队伍名称必填")
	}
	t := &model.Team{LeaderID: leader, RouteID: routeID, Name: name, InviteCode: invite6(), Status: "open"}
	if err := a.Teams.Create(ctx, t); err != nil {
		return nil, httpx.Internal("创建队伍失败")
	}
	if err := a.Teams.AddMember(ctx, t.ID, leader, "leader"); err != nil {
		return nil, httpx.Internal("加入队长失败")
	}
	return a.Teams.Get(ctx, t.ID)
}

func (a *App) JoinTeam(ctx context.Context, userID int64, code string) (*model.Team, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 6 {
		return nil, httpx.Validation("邀请码为 6 位")
	}
	t, err := a.Teams.ByInvite(ctx, code)
	if err != nil || t == nil {
		return nil, httpx.NotFound("邀请码无效")
	}
	if err := a.Teams.AddMember(ctx, t.ID, userID, "member"); err != nil {
		return nil, httpx.Internal("加入失败")
	}
	return a.Teams.Get(ctx, t.ID)
}

func (a *App) Kick(ctx context.Context, actor, teamID, target int64) error {
	t, err := a.Teams.Get(ctx, teamID)
	if err != nil || t == nil {
		return httpx.NotFound("队伍不存在")
	}
	if t.LeaderID != actor {
		return httpx.Forbidden("仅队长可踢人")
	}
	if target == actor {
		return httpx.Validation("队长不能踢自己")
	}
	return a.Teams.RemoveMember(ctx, teamID, target)
}

func (a *App) CreateTrip(ctx context.Context, userID, teamID int64) (*model.Trip, error) {
	t, err := a.mustMember(ctx, teamID, userID)
	if err != nil {
		return nil, err
	}
	tr := &model.Trip{TeamID: teamID, RouteID: t.RouteID, Status: model.TripDraft}
	if err := a.Trips.Create(ctx, tr); err != nil {
		return nil, httpx.Internal("创建行程失败")
	}
	return tr, nil
}

func (a *App) Transit(ctx context.Context, userID, tripID int64, to string) (*model.Trip, error) {
	tr, team, err := a.tripOf(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	if team.LeaderID != userID {
		return nil, httpx.Forbidden("仅队长可变更行程状态")
	}
	if !model.CanTransit(tr.Status, to) {
		return nil, httpx.BadState("非法状态跃迁 " + tr.Status + " → " + to)
	}
	if err := a.Trips.SetStatus(ctx, tripID, to); err != nil {
		return nil, httpx.Internal("更新行程失败")
	}
	return a.Trips.Get(ctx, tripID)
}

func (a *App) LivePosition(ctx context.Context, userID, tripID int64, lat, lon float64, elev *float64) error {
	if !geo.ValidateLatLon(lat, lon) {
		return httpx.Validation("经纬度越界")
	}
	tr, team, err := a.tripOf(ctx, userID, tripID)
	if err != nil {
		return err
	}
	if tr.Status != model.TripActive && tr.Status != model.TripPaused {
		return httpx.BadState("行程未开始")
	}
	now := timeutil.NowNaive()
	e := 0.0
	if elev != nil {
		e = *elev
	} else {
		e = dem.ElevationAt(lat, lon)
	}
	acc := 8.0
	pts := []model.TrackPoint{{
		TripID: tripID, MemberID: userID, Lat: lat, Lon: lon, Elevation: &e, Accuracy: &acc,
		RecordedAt: now, Source: "live", Fingerprint: track.Fingerprint(userID, lat, lon, now),
	}}
	if _, err := a.Tracks.InsertPoints(ctx, pts); err != nil {
		return httpx.Internal("写入位置失败")
	}
	a.Grid.Upsert(geo.LiveFix{MemberID: userID, Lat: lat, Lon: lon, Elev: e, UnixMs: now.UnixMilli()})
	_ = a.Teams.SetState(ctx, team.ID, userID, "online")
	if a.Redis != nil {
		_ = a.Redis.HSet(ctx, fmt.Sprintf("trip:%d:pos", tripID), fmt.Sprintf("%d", userID),
			fmt.Sprintf("%.6f,%.6f,%.1f,%d", lat, lon, e, now.Unix())).Err()
	}
	if a.PosHz.Allow(userID, time.Now()) {
		a.Hub.Broadcast(tripID, ws.TypePos, map[string]any{
			"member_id": userID, "lat": lat, "lon": lon, "elevation": e,
			"recorded_at": timeutil.FormatDisplay(now), "s2_cell": geo.CellToken(lat, lon, geo.BucketLevel),
		})
	}
	a.maybeRisk(ctx, tr, team)
	return nil
}

func (a *App) MergeBatch(ctx context.Context, userID, tripID int64, raw []model.RawPoint) (*model.MergeResult, error) {
	tr, _, err := a.tripOf(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	_ = tr
	hash := track.BatchHash(userID, tripID, raw)
	if old, _ := a.Tracks.GetBatch(ctx, tripID, userID, hash); old != nil {
		return &model.MergeResult{Accepted: old.Accepted, Rejected: old.Rejected, Idempotent: true, Segments: []model.TrackSegment{}}, nil
	}
	known, err := a.Tracks.Fingerprints(ctx, tripID, userID)
	if err != nil {
		return nil, httpx.Internal("读取指纹失败")
	}
	existModel, err := a.Tracks.List(ctx, tripID, userID)
	if err != nil {
		return nil, httpx.Internal("读取既有轨迹失败")
	}
	pipe := track.Pipeline{MemberID: userID, TripID: tripID}
	out, err := pipe.Run(raw, track.FromModel(existModel), known)
	if err != nil {
		return nil, err
	}
	smooth := out.Smoothed
	merged := out.Merged
	segs := out.Segments
	st := out.Stats
	dups := out.Dups
	_ = hash
	toWrite := track.ToModelPoints(tripID, userID, "offline_batch", smooth)
	n, err := a.Tracks.InsertPoints(ctx, toWrite)
	if err != nil {
		return nil, httpx.Internal("写入轨迹失败")
	}
	_ = a.Tracks.ReplaceSegments(ctx, tripID, userID, segs)
	b := &model.TrackBatch{TripID: tripID, MemberID: userID, BatchHash: hash, PointCount: len(raw), Accepted: n, Rejected: len(raw) - n}
	_ = a.Tracks.SaveBatch(ctx, b)
	met := track.ComputeMetrics(merged)
	if len(smooth) > 0 {
		last := smooth[len(smooth)-1]
		a.Grid.Upsert(geo.LiveFix{MemberID: userID, Lat: last.Lat, Lon: last.Lon, UnixMs: last.RecordedAt.UnixMilli()})
	}
	res := &model.MergeResult{
		Accepted: n, Rejected: st.Accuracy + st.Speed + st.Accel + st.Still, Duplicates: dups,
		DistanceM: met.DistanceM, MovingSeconds: met.MovingSeconds, AvgSpeedKmh: met.AvgSpeedKmh,
		AscentM: met.AscentM, Segments: segs,
	}
	a.Hub.Broadcast(tripID, ws.TypeTrack, res)
	return res, nil
}

func (a *App) ComputeETA(ctx context.Context, userID, tripID int64) (*model.ETAResult, error) {
	tr, team, err := a.tripOf(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	pts, err := a.Tracks.List(ctx, tripID, userID)
	if err != nil {
		return nil, httpx.Internal("读取轨迹失败")
	}
	parsed := track.FromModel(pts)
	spd := track.RecentSpeedKmh(parsed, 15*60)
	var path [][2]float64
	if team.RouteID != nil {
		rb, _ := a.Routes.Get(ctx, *team.RouteID)
		if rb != nil {
			path = pathOf(rb.Waypoints)
		}
	}
	lat, lon := 0.0, 0.0
	if len(parsed) > 0 {
		lat, lon = parsed[len(parsed)-1].Lat, parsed[len(parsed)-1].Lon
	} else if len(path) > 0 {
		lat, lon = path[0][0], path[0][1]
	}
	var elev []float64
	if len(path) >= 2 {
		if prof, err := dem.BuildProfile(ctx, a.DEM, path); err == nil {
			elev = make([]float64, len(path))
			for i := range path {
				elev[i] = dem.ElevationAt(path[i][0], path[i][1])
			}
			_ = prof
		}
	}
	pr := eta.Project(lat, lon, path, elev)
	out := eta.Estimate(pr.RemainM, spd, pr.AvgSlope, timeutil.Now())
	a.Hub.Broadcast(tr.ID, ws.TypeETA, out)
	return &out, nil
}

func (a *App) RaiseSOS(ctx context.Context, userID, tripID int64, typ string, lat, lon float64, reason string) (*model.SOSEvent, error) {
	if !geo.ValidateLatLon(lat, lon) {
		return nil, httpx.Validation("经纬度越界")
	}
	if typ == "" {
		typ = "manual"
	}
	if _, _, err := a.tripOf(ctx, userID, tripID); err != nil {
		return nil, err
	}
	e := &model.SOSEvent{TripID: tripID, MemberID: userID, Type: typ, Lat: lat, Lon: lon, Reason: reason, Status: "open"}
	if err := a.SOS.Create(ctx, e); err != nil {
		return nil, httpx.Internal("呼救失败")
	}
	u, _ := a.Users.ByID(ctx, userID)
	if u != nil {
		e.Nickname = u.Nickname
	}
	a.Hub.Broadcast(tripID, ws.TypeSOS, e)
	_ = a.Notify.Send(ctx, notify.Event{Channel: "sos", Kind: typ, Payload: map[string]any{"trip_id": tripID, "member_id": userID, "lat": lat, "lon": lon}})
	return e, nil
}

func (a *App) RiskNow(ctx context.Context, userID, tripID int64) (*model.RiskReport, error) {
	tr, team, err := a.tripOf(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	return a.evalRisk(ctx, tr, team)
}

func (a *App) Backtrack(ctx context.Context, userID, tripID int64) (map[string]any, error) {
	tr, team, err := a.tripOf(ctx, userID, tripID)
	if err != nil {
		return nil, err
	}
	pts, err := a.Tracks.List(ctx, tripID, userID)
	if err != nil {
		return nil, err
	}
	if len(pts) == 0 {
		return nil, httpx.NotFound("尚无轨迹可回溯")
	}
	var safe []model.Waypoint
	if team.RouteID != nil {
		rb, _ := a.Routes.Get(ctx, *team.RouteID)
		if rb != nil {
			for _, w := range rb.Waypoints {
				if w.Type == "camp" || w.Type == "water" {
					safe = append(safe, w)
				}
			}
		}
	}
	last := pts[len(pts)-1]
	bestI := 0
	bestD := 1e18
	target := "trailhead"
	tlat, tlon := pts[0].Lat, pts[0].Lon
	for i, p := range pts {
		for _, s := range safe {
			d := geo.HaversineM(p.Lat, p.Lon, s.Lat, s.Lon)
			if d < bestD {
				bestD = d
				bestI = i
				target = s.Type
				tlat, tlon = s.Lat, s.Lon
			}
		}
	}
	if len(safe) == 0 {
		bestI = 0
		tlat, tlon = pts[0].Lat, pts[0].Lon
	}
	rev := make([]map[string]float64, 0)
	var dist float64
	for i := len(pts) - 1; i >= bestI; i-- {
		rev = append(rev, map[string]float64{"lat": pts[i].Lat, "lon": pts[i].Lon})
		if i < len(pts)-1 {
			dist += geo.HaversineM(pts[i+1].Lat, pts[i+1].Lon, pts[i].Lat, pts[i].Lon)
		}
	}
	_ = tr
	return map[string]any{
		"from": map[string]float64{"lat": last.Lat, "lon": last.Lon},
		"to":   map[string]any{"lat": tlat, "lon": tlon, "kind": target},
		"path": rev, "distance_m": dist,
	}, nil
}

func (a *App) maybeRisk(ctx context.Context, tr *model.Trip, team *model.Team) {
	if !a.RiskHz.Allow(tr.ID, time.Now()) {
		return
	}
	go func() {
		bg := context.Background()
		rep, err := a.evalRisk(bg, tr, team)
		if err != nil || rep == nil {
			return
		}
		a.Hub.Broadcast(tr.ID, ws.TypeRisk, rep)
		ids := make([]int64, 0)
		pts := make([][2]float64, 0)
		for _, m := range team.Members {
			fixes := a.Grid.Members([]int64{m.UserID})
			if len(fixes) == 1 {
				ids = append(ids, m.UserID)
				pts = append(pts, [2]float64{fixes[0].Lat, fixes[0].Lon})
			}
		}
		for _, uid := range risk.AutoSOSCandidates(pts, ids, 800) {
			has, _ := a.SOS.HasOpenAuto(bg, tr.ID, uid)
			if has {
				continue
			}
			fx := a.Grid.Members([]int64{uid})
			if len(fx) == 0 {
				continue
			}
			_, _ = a.RaiseSOS(bg, uid, tr.ID, "auto", fx[0].Lat, fx[0].Lon, "队员距离质心超过 800 米")
		}
	}()
}

func (a *App) evalRisk(ctx context.Context, tr *model.Trip, team *model.Team) (*model.RiskReport, error) {
	ids := make([]int64, 0, len(team.Members))
	for _, m := range team.Members {
		ids = append(ids, m.UserID)
	}
	fixes := a.Grid.Members(ids)
	members := make([][2]float64, 0, len(fixes))
	for _, f := range fixes {
		members = append(members, [2]float64{f.Lat, f.Lon})
	}
	if len(members) == 0 {
		members = [][2]float64{{30.26, 119.72}}
	}
	clat, clon := geo.Centroid(members)
	var dangers, waters []model.Waypoint
	if team.RouteID != nil {
		rb, _ := a.Routes.Get(ctx, *team.RouteID)
		if rb != nil {
			for _, w := range rb.Waypoints {
				switch w.Type {
				case "danger":
					dangers = append(dangers, w)
				case "water":
					waters = append(waters, w)
				}
			}
		}
	}
	rep := risk.Evaluate(clat, clon, members, dangers, waters, timeutil.Now())
	rep.TripID = tr.ID
	_ = a.Risks.Save(ctx, &rep)
	return &rep, nil
}

func (a *App) TripView(ctx context.Context, userID, tripID int64) (*model.Trip, *model.Team, error) {
	return a.tripOf(ctx, userID, tripID)
}

func (a *App) tripOf(ctx context.Context, userID, tripID int64) (*model.Trip, *model.Team, error) {
	tr, err := a.Trips.Get(ctx, tripID)
	if err != nil || tr == nil {
		return nil, nil, httpx.NotFound("行程不存在")
	}
	team, err := a.mustMember(ctx, tr.TeamID, userID)
	if err != nil {
		return nil, nil, err
	}
	return tr, team, nil
}

func (a *App) mustMember(ctx context.Context, teamID, userID int64) (*model.Team, error) {
	ok, _, err := a.Teams.IsMember(ctx, teamID, userID)
	if err != nil {
		return nil, httpx.Internal("校验队员失败")
	}
	if !ok {
		return nil, httpx.Forbidden("不是该队伍成员")
	}
	return a.Teams.Get(ctx, teamID)
}

func pathOf(wps []model.Waypoint) [][2]float64 {
	out := make([][2]float64, 0, len(wps))
	for _, w := range wps {
		if w.Type == "danger" {
			continue
		}
		out = append(out, [2]float64{w.Lat, w.Lon})
	}
	return out
}

func invite6() string {
	var b [3]byte
	_, _ = rand.Read(b[:])
	s := strings.ToUpper(hex.EncodeToString(b[:]))
	if len(s) > 6 {
		return s[:6]
	}
	return s
}

func ExportGPX(rb *model.RouteBook) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?><gpx version="1.1" creator="gocamping"><trk><name>`)
	b.WriteString(rb.Title)
	b.WriteString(`</name><trkseg>`)
	for _, w := range rb.Waypoints {
		if w.Type == "danger" {
			continue
		}
		fmt.Fprintf(&b, `<trkpt lat="%.6f" lon="%.6f">`, w.Lat, w.Lon)
		if w.Elevation != nil {
			fmt.Fprintf(&b, `<ele>%.1f</ele>`, *w.Elevation)
		}
		b.WriteString(`</trkpt>`)
	}
	b.WriteString(`</trkseg></trk></gpx>`)
	return b.String()
}

func ImportGeoJSON(raw []byte) ([]model.Waypoint, error) {
	var gj struct {
		Type     string `json:"type"`
		Features []struct {
			Geometry struct {
				Type        string          `json:"type"`
				Coordinates json.RawMessage `json:"coordinates"`
			} `json:"geometry"`
			Properties struct {
				Type string `json:"type"`
				Note string `json:"note"`
			} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(raw, &gj); err != nil {
		return nil, httpx.Validation("GeoJSON 无法解析")
	}
	if gj.Type != "FeatureCollection" {
		return nil, httpx.Validation("需要 FeatureCollection")
	}
	var out []model.Waypoint
	for i, f := range gj.Features {
		typ := f.Properties.Type
		if typ == "" {
			typ = "waypoint"
		}
		switch f.Geometry.Type {
		case "Point":
			var c []float64
			if err := json.Unmarshal(f.Geometry.Coordinates, &c); err != nil || len(c) < 2 {
				return nil, httpx.Validation(fmt.Sprintf("第 %d 个点坐标非法", i+1))
			}
			lon, lat := c[0], c[1]
			if !geo.ValidateLatLon(lat, lon) {
				return nil, httpx.Validation(fmt.Sprintf("第 %d 个点经纬度越界", i+1))
			}
			out = append(out, model.Waypoint{Seq: i, Type: typ, Lat: lat, Lon: lon, Note: f.Properties.Note, RiskWeight: 1, Polygon: [][2]float64{}})
		default:
			return nil, httpx.Validation("暂仅支持 Point 要素")
		}
	}
	if len(out) == 0 {
		return nil, httpx.Validation("未找到有效点位")
	}
	return out, nil
}
