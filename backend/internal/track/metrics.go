package track

import (
	"gocamping/internal/geo"
	"gocamping/internal/model"
)

type Metrics struct {
	DistanceM     float64
	MovingSeconds float64
	AvgSpeedKmh   float64
	AscentM       float64
	DescentM      float64
}

func ComputeMetrics(pts []ParsedPoint) Metrics {
	var m Metrics
	if len(pts) < 2 {
		return m
	}
	for i := 1; i < len(pts); i++ {
		dt := pts[i].RecordedAt.Sub(pts[i-1].RecordedAt).Seconds()
		if dt <= 0 {
			continue
		}
		d := geo.HaversineM(pts[i-1].Lat, pts[i-1].Lon, pts[i].Lat, pts[i].Lon)
		if d < 2 { // still
			continue
		}
		m.DistanceM += d
		m.MovingSeconds += dt
		if pts[i].Elevation != nil && pts[i-1].Elevation != nil {
			de := *pts[i].Elevation - *pts[i-1].Elevation
			if de > 0.8 {
				m.AscentM += de
			} else if de < -0.8 {
				m.DescentM += -de
			}
		}
	}
	if m.MovingSeconds > 0 {
		m.AvgSpeedKmh = (m.DistanceM / m.MovingSeconds) * 3.6
	}
	return m
}

func RecentSpeedKmh(pts []ParsedPoint, windowSec float64) float64 {
	if len(pts) < 2 {
		return 0
	}
	end := pts[len(pts)-1].RecordedAt
	var dist, dur float64
	for i := len(pts) - 1; i > 0; i-- {
		if end.Sub(pts[i-1].RecordedAt).Seconds() > windowSec {
			break
		}
		dt := pts[i].RecordedAt.Sub(pts[i-1].RecordedAt).Seconds()
		if dt <= 0 {
			continue
		}
		d := geo.HaversineM(pts[i-1].Lat, pts[i-1].Lon, pts[i].Lat, pts[i].Lon)
		if d < 2 {
			continue
		}
		dist += d
		dur += dt
	}
	if dur <= 0 {
		return 0
	}
	return (dist / dur) * 3.6
}

func ToModelPoints(tripID, memberID int64, src string, pts []ParsedPoint) []model.TrackPoint {
	out := make([]model.TrackPoint, 0, len(pts))
	for _, p := range pts {
		out = append(out, model.TrackPoint{
			TripID:      tripID,
			MemberID:    memberID,
			Lat:         p.Lat,
			Lon:         p.Lon,
			Elevation:   p.Elevation,
			Accuracy:    p.Accuracy,
			Speed:       p.Speed,
			RecordedAt:  p.RecordedAt,
			Source:      src,
			IsNoise:     false,
			Fingerprint: p.Fingerprint,
		})
	}
	return out
}

func FromModel(pts []model.TrackPoint) []ParsedPoint {
	out := make([]ParsedPoint, 0, len(pts))
	for _, p := range pts {
		out = append(out, ParsedPoint{
			Lat:         p.Lat,
			Lon:         p.Lon,
			Elevation:   p.Elevation,
			Accuracy:    p.Accuracy,
			Speed:       p.Speed,
			RecordedAt:  p.RecordedAt,
			Fingerprint: p.Fingerprint,
		})
	}
	return out
}
