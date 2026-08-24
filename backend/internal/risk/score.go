package risk

import (
	"math"
	"time"

	"gocamping/internal/dem"
	"gocamping/internal/geo"
	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type Weights struct {
	Danger, Slope, Water, Sunset, Disp float64
}

func DefaultWeights() Weights {
	return Weights{Danger: 0.34, Slope: 0.18, Water: 0.16, Sunset: 0.12, Disp: 0.20}
}

func Evaluate(leaderLat, leaderLon float64, members [][2]float64, dangers []model.Waypoint, waters []model.Waypoint, now time.Time) model.RiskReport {
	tree := geo.NewRTree()
	for _, d := range dangers {
		box := geo.CircleBBox(d.Lat, d.Lon, 80)
		if d.RadiusM != nil && *d.RadiusM > 0 {
			box = geo.CircleBBox(d.Lat, d.Lon, *d.RadiusM)
		}
		if len(d.Polygon) >= 3 {
			box = geo.RingBBox(d.Polygon)
		}
		tree.Insert(geo.RTItem{ID: d.ID, Box: box, Data: d})
	}
	q := geo.CircleBBox(leaderLat, leaderLon, 1000)
	cands := tree.Search(q)
	hits := make([]model.RiskHit, 0)
	dangerScore := 0.0
	for _, c := range cands {
		d, _ := c.Data.(model.Waypoint)
		ok := false
		dist := geo.HaversineM(leaderLat, leaderLon, d.Lat, d.Lon)
		if len(d.Polygon) >= 3 {
			ok = geo.PointInRing(leaderLat, leaderLon, d.Polygon)
			// also count if within 1km of polygon bbox centroid
			if !ok && dist <= 1000 {
				ok = true
			}
		} else {
			r := 80.0
			if d.RadiusM != nil && *d.RadiusM > 0 {
				r = *d.RadiusM
			}
			ok = dist <= math.Max(r, 1000) && (dist <= r || dist <= 1000)
			if dist > 1000 {
				ok = false
			}
			if dist <= r {
				ok = true
			} else if dist <= 1000 {
				ok = true
			}
		}
		if !ok {
			continue
		}
		w := d.RiskWeight
		if w < 1 {
			w = 1
		}
		if w > 5 {
			w = 5
		}
		hits = append(hits, model.RiskHit{WaypointID: d.ID, Type: d.Type, Note: d.Note, Weight: w, DistanceM: math.Round(dist)})
		dangerScore += float64(w) / 5.0
	}
	if dangerScore > 1 {
		dangerScore = 1
	}

	// slope from synthetic DEM around leader
	e0 := dem.ElevationAt(leaderLat, leaderLon)
	eN := dem.ElevationAt(leaderLat+0.001, leaderLon)
	eE := dem.ElevationAt(leaderLat, leaderLon+0.001)
	dn := geo.HaversineM(leaderLat, leaderLon, leaderLat+0.001, leaderLon)
	de := geo.HaversineM(leaderLat, leaderLon, leaderLat, leaderLon+0.001)
	sl := math.Max(math.Abs(geo.SlopeRatio(dn, eN-e0)), math.Abs(geo.SlopeRatio(de, eE-e0)))
	slopeScore := clamp01(sl / 0.35)

	waterDist := 5000.0
	for _, w := range waters {
		d := geo.HaversineM(leaderLat, leaderLon, w.Lat, w.Lon)
		if d < waterDist {
			waterDist = d
		}
	}
	waterScore := clamp01(waterDist / 2500)

	sunLeft := timeutil.HoursUntilSunset(now, leaderLon, leaderLat)
	sunsetScore := 0.0
	if sunLeft < 2 {
		sunsetScore = clamp01((2 - sunLeft) / 2)
	}

	disp := Dispersion(members)
	wts := DefaultWeights()
	score := wts.Danger*dangerScore + wts.Slope*slopeScore + wts.Water*waterScore + wts.Sunset*sunsetScore + wts.Disp*disp.Index
	level := "low"
	switch {
	case score >= 0.75:
		level = "critical"
	case score >= 0.55:
		level = "high"
	case score >= 0.32:
		level = "medium"
	}
	if hits == nil {
		hits = []model.RiskHit{}
	}
	clat, clon := geo.Centroid(members)
	return model.RiskReport{
		Level:       level,
		Score:       math.Round(score*1000) / 1000,
		Dispersion:  math.Round(disp.Index*1000) / 1000,
		MaxSlope:    math.Round(sl*1000) / 1000,
		WaterDistM:  math.Round(waterDist),
		SunsetLeftH: math.Round(sunLeft*10) / 10,
		Hits:        hits,
		CentroidLat: clat,
		CentroidLon: clon,
		FarthestM:   math.Round(disp.Farthest),
		ComputedAt:  timeutil.ToBeijing(now),
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
