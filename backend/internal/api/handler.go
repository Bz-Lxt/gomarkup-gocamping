package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"gocamping/internal/config"
	"gocamping/internal/gpx"
	"gocamping/internal/httpx"
	"gocamping/internal/model"
	"gocamping/internal/replay"
	"gocamping/internal/service"
	"gocamping/internal/simulator"
	"gocamping/internal/timeutil"
)

type Handler struct {
	Cfg config.Config
	App *service.App
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Password, Nickname string }
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, err := service.Register(r.Context(), h.App.Users, in.Username, in.Password, in.Nickname)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	tok, err := httpx.SignJWT(h.Cfg.JWTSecret, u.ID, u.Role, h.Cfg.JWTExpire)
	if err != nil {
		httpx.Fail(w, r, httpx.Internal("签发失败"))
		return
	}
	httpx.JSON(w, r, 200, map[string]any{"token": tok, "user": u.Public()})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Username, Password string }
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, err := h.App.Users.ByUsername(r.Context(), in.Username)
	if err != nil || u == nil || !service.CheckPassword(u.PasswordHash, in.Password) {
		httpx.Fail(w, r, httpx.Unauthorized("用户名或密码错误"))
		return
	}
	tok, err := httpx.SignJWT(h.Cfg.JWTSecret, u.ID, u.Role, h.Cfg.JWTExpire)
	if err != nil {
		httpx.Fail(w, r, httpx.Internal("签发失败"))
		return
	}
	httpx.JSON(w, r, 200, map[string]any{"token": tok, "user": u.Public()})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, err := h.App.Users.ByID(r.Context(), httpx.UserID(r))
	if err != nil || u == nil {
		httpx.Fail(w, r, httpx.NotFound("用户不存在"))
		return
	}
	httpx.JSON(w, r, 200, u.Public())
}

func (h *Handler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	mine := r.URL.Query().Get("scope")
	owner := int64(0)
	vis := ""
	if mine != "public" {
		owner = httpx.UserID(r)
	} else {
		vis = "public"
	}
	list, err := h.App.Routes.List(r.Context(), owner, vis)
	if err != nil {
		httpx.Fail(w, r, httpx.Internal("查询失败"))
		return
	}
	httpx.JSON(w, r, 200, list)
}

func (h *Handler) SaveRoute(w http.ResponseWriter, r *http.Request) {
	var in model.RouteBook
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	in.ID = 0
	rb, err := h.App.SaveRoute(r.Context(), httpx.UserID(r), &in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, rb)
}

func (h *Handler) UpdateRoute(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in model.RouteBook
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	in.ID = id
	rb, err := h.App.SaveRoute(r.Context(), httpx.UserID(r), &in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, rb)
}

func (h *Handler) GetRoute(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	rb, err := h.App.Routes.Get(r.Context(), id)
	if err != nil || rb == nil {
		httpx.Fail(w, r, httpx.NotFound("路书不存在"))
		return
	}
	httpx.JSON(w, r, 200, rb)
}

func (h *Handler) Elevation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	p, err := h.App.Elevation(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, p)
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	rb, err := h.App.Routes.Get(r.Context(), id)
	if err != nil || rb == nil {
		httpx.Fail(w, r, httpx.NotFound("路书不存在"))
		return
	}
	fmt := r.URL.Query().Get("format")
	if fmt == "geojson" {
		feats := make([]map[string]any, 0, len(rb.Waypoints))
		for _, wp := range rb.Waypoints {
			feats = append(feats, map[string]any{
				"type": "Feature",
				"geometry": map[string]any{"type": "Point", "coordinates": []float64{wp.Lon, wp.Lat}},
				"properties": map[string]any{"type": wp.Type, "note": wp.Note},
			})
		}
		httpx.JSON(w, r, 200, map[string]any{"type": "FeatureCollection", "features": feats})
		return
	}
	httpx.JSON(w, r, 200, map[string]string{"gpx": service.ExportGPX(rb)})
}

func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.Fail(w, r, httpx.BadRequest("读取失败"))
		return
	}
	var wps []model.Waypoint
	if len(body) > 0 && body[0] == '<' {
		wps, err = gpx.Parse(body)
	} else {
		wps, err = service.ImportGeoJSON(body)
	}
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, map[string]any{"waypoints": wps})
}

func (h *Handler) ListTeams(w http.ResponseWriter, r *http.Request) {
	list, err := h.App.Teams.ListByUser(r.Context(), httpx.UserID(r))
	if err != nil {
		httpx.Fail(w, r, httpx.Internal("查询失败"))
		return
	}
	httpx.JSON(w, r, 200, list)
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name    string `json:"name"`
		RouteID *int64 `json:"route_id"`
	}
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	t, err := h.App.CreateTeam(r.Context(), httpx.UserID(r), in.Name, in.RouteID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, t)
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	t, err := h.App.Teams.Get(r.Context(), id)
	if err != nil || t == nil {
		httpx.Fail(w, r, httpx.NotFound("队伍不存在"))
		return
	}
	httpx.JSON(w, r, 200, t)
}

func (h *Handler) JoinTeam(w http.ResponseWriter, r *http.Request) {
	var in struct{ Code string `json:"code"` }
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	t, err := h.App.JoinTeam(r.Context(), httpx.UserID(r), in.Code)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, t)
}

func (h *Handler) Kick(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in struct{ UserID int64 `json:"user_id"` }
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.App.Kick(r.Context(), httpx.UserID(r), id, in.UserID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, map[string]string{"status": "kicked"})
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var in struct{ TeamID int64 `json:"team_id"` }
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	tr, err := h.App.CreateTrip(r.Context(), httpx.UserID(r), in.TeamID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, tr)
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	tr, _, err := h.App.TripView(r.Context(), httpx.UserID(r), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, tr)
}

func (h *Handler) ListTrips(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	list, err := h.App.Trips.ListByTeam(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, list)
}

func (h *Handler) StartTrip(w http.ResponseWriter, r *http.Request)  { h.transit(w, r, model.TripActive) }
func (h *Handler) PauseTrip(w http.ResponseWriter, r *http.Request)  { h.transit(w, r, model.TripPaused) }
func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request) { h.transit(w, r, model.TripFinished) }

func (h *Handler) transit(w http.ResponseWriter, r *http.Request, to string) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	tr, err := h.App.Transit(r.Context(), httpx.UserID(r), id, to)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, tr)
}

func (h *Handler) PostPosition(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in struct {
		Lat, Lon float64
		Elevation *float64 `json:"elevation"`
	}
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.App.LivePosition(r.Context(), httpx.UserID(r), id, in.Lat, in.Lon, in.Elevation); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, map[string]string{"status": "ok"})
}

func (h *Handler) BatchTracks(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in struct {
		Points []model.RawPoint `json:"points"`
	}
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	res, err := h.App.MergeBatch(r.Context(), httpx.UserID(r), id, in.Points)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, res)
}

func (h *Handler) GetTracks(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	mid, _ := strconv.ParseInt(r.URL.Query().Get("member_id"), 10, 64)
	pts, err := h.App.Tracks.List(r.Context(), id, mid)
	if err != nil {
		httpx.Fail(w, r, httpx.Internal("查询失败"))
		return
	}
	httpx.JSON(w, r, 200, pts)
}

func (h *Handler) ETA(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	out, err := h.App.ComputeETA(r.Context(), httpx.UserID(r), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, out)
}

func (h *Handler) SOS(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var in struct {
		Lat, Lon float64
		Reason   string
	}
	if err := decode(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	e, err := h.App.RaiseSOS(r.Context(), httpx.UserID(r), id, "manual", in.Lat, in.Lon, in.Reason)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, e)
}

func (h *Handler) ListSOS(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	list, err := h.App.SOS.List(r.Context(), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, list)
}

func (h *Handler) ResolveSOS(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := h.App.SOS.Resolve(r.Context(), id); err != nil {
		httpx.Fail(w, r, httpx.Internal("处置失败"))
		return
	}
	httpx.JSON(w, r, 200, map[string]string{"status": "resolved"})
}

func (h *Handler) Risk(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	rep, err := h.App.RiskNow(r.Context(), httpx.UserID(r), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, rep)
}

func (h *Handler) Backtrack(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	out, err := h.App.Backtrack(r.Context(), httpx.UserID(r), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, out)
}

func (h *Handler) Replay(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if _, _, err := h.App.TripView(r.Context(), httpx.UserID(r), id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	pts, err := h.App.Tracks.List(r.Context(), id, 0)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, replay.Build(pts, 10*time.Second))
}

func (h *Handler) Simulate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	tr, team, err := h.App.TripView(r.Context(), httpx.UserID(r), id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var path [][2]float64
	if team.RouteID != nil {
		rb, _ := h.App.Routes.Get(r.Context(), *team.RouteID)
		if rb != nil {
			for _, wp := range rb.Waypoints {
				if wp.Type != "danger" {
					path = append(path, [2]float64{wp.Lat, wp.Lon})
				}
			}
		}
	}
	if len(path) < 2 {
		path = [][2]float64{{30.26, 119.72}, {30.28, 119.75}, {30.30, 119.78}}
	}
	pts := simulator.Walk(path, timeutil.Now().Add(-4*time.Hour), 48, simulator.Options{NoiseSigmaM: 12, SpeedKmh: 3.5})
	pts = simulator.InjectOutliers(pts, 8, 420)
	res, err := h.App.MergeBatch(r.Context(), httpx.UserID(r), tr.ID, pts)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, res)
}

func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	list, err := h.App.Users.List(r.Context(), 100, 0)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	pub := make([]model.PublicUser, 0, len(list))
	for _, u := range list {
		pub = append(pub, u.Public())
	}
	httpx.JSON(w, r, 200, pub)
}

func (h *Handler) AdminRoutes(w http.ResponseWriter, r *http.Request) {
	list, err := h.App.Routes.List(r.Context(), 0, "")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, r, 200, list)
}

func (h *Handler) AdminMetrics(w http.ResponseWriter, r *http.Request) {
	uc, _ := h.App.Users.Count(r.Context())
	httpx.JSON(w, r, 200, map[string]any{
		"users": uc, "ws_hint": "per-trip rooms", "tz": "Asia/Shanghai",
		"providers": map[string]string{"tile": h.Cfg.TileProvider, "dem": h.Cfg.DEMProvider, "gps": h.Cfg.GPSProvider, "notify": h.Cfg.NotifyProvider},
	})
}

func (h *Handler) WS(w http.ResponseWriter, r *http.Request) {
	tok := httpx.Bearer(r)
	if tok == "" {
		httpx.Fail(w, r, httpx.Unauthorized("请先登录"))
		return
	}
	c, err := httpx.ParseJWT(h.Cfg.JWTSecret, tok)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	tripID, _ := strconv.ParseInt(r.URL.Query().Get("trip_id"), 10, 64)
	if tripID == 0 {
		httpx.Fail(w, r, httpx.Validation("缺少 trip_id"))
		return
	}
	if _, _, err := h.App.TripView(r.Context(), c.UserID, tripID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	h.App.Hub.ServeWS(w, r, c.UserID, tripID)
}

func decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 4<<20))
	if err := dec.Decode(dst); err != nil {
		return httpx.Validation("JSON 无法解析或字段类型错误")
	}
	return nil
}

func pathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, httpx.Validation("非法 id")
	}
	return id, nil
}

