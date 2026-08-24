package replay

import (
	"sort"
	"time"

	"gocamping/internal/model"
)

type Frame struct {
	At     time.Time          `json:"at"`
	AtDisp string             `json:"at_disp"`
	Fixes  []model.LivePosition `json:"fixes"`
}

// Build constructs a time-aligned replay from all members' points.
func Build(pts []model.TrackPoint, step time.Duration) []Frame {
	if step <= 0 {
		step = 5 * time.Second
	}
	if len(pts) == 0 {
		return []Frame{}
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i].RecordedAt.Before(pts[j].RecordedAt) })
	start, end := pts[0].RecordedAt, pts[len(pts)-1].RecordedAt
	latest := map[int64]model.TrackPoint{}
	idx := 0
	var frames []Frame
	for t := start; !t.After(end); t = t.Add(step) {
		for idx < len(pts) && !pts[idx].RecordedAt.After(t) {
			latest[pts[idx].MemberID] = pts[idx]
			idx++
		}
		fixes := make([]model.LivePosition, 0, len(latest))
		for _, p := range latest {
			e := 0.0
			if p.Elevation != nil {
				e = *p.Elevation
			}
			fixes = append(fixes, model.LivePosition{
				MemberID: p.MemberID, Lat: p.Lat, Lon: p.Lon, Elevation: e, RecordedAt: p.RecordedAt,
			})
		}
		frames = append(frames, Frame{At: t, AtDisp: t.Format("2006-01-02 15:04:05"), Fixes: fixes})
		if len(frames) > 2000 {
			break
		}
	}
	return frames
}
