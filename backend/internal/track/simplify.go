package track

import "gocamping/internal/geo"

// DouglasPeucker simplifies a path. Tolerance is metres.
func DouglasPeucker(pts []ParsedPoint, tolM float64) []ParsedPoint {
	if len(pts) < 3 {
		return append([]ParsedPoint(nil), pts...)
	}
	if tolM <= 0 {
		tolM = 8
	}
	keep := make([]bool, len(pts))
	keep[0] = true
	keep[len(pts)-1] = true
	dp(pts, 0, len(pts)-1, tolM, keep)
	out := make([]ParsedPoint, 0, len(pts))
	for i, k := range keep {
		if k {
			out = append(out, pts[i])
		}
	}
	return out
}

func dp(pts []ParsedPoint, start, end int, tol float64, keep []bool) {
	if end <= start+1 {
		return
	}
	maxD := -1.0
	idx := start
	for i := start + 1; i < end; i++ {
		d := perpDistM(pts[i], pts[start], pts[end])
		if d > maxD {
			maxD = d
			idx = i
		}
	}
	if maxD > tol {
		keep[idx] = true
		dp(pts, start, idx, tol, keep)
		dp(pts, idx, end, tol, keep)
	}
}

func perpDistM(p, a, b ParsedPoint) float64 {
	seg := geo.HaversineM(a.Lat, a.Lon, b.Lat, b.Lon)
	if seg < 1e-3 {
		return geo.HaversineM(p.Lat, p.Lon, a.Lat, a.Lon)
	}
	// use cross-track approximation via destination projection
	path := [][2]float64{{a.Lat, a.Lon}, {b.Lat, b.Lon}}
	plat, plon, _, _ := geo.ProjectOnto(p.Lat, p.Lon, path)
	return geo.HaversineM(p.Lat, p.Lon, plat, plon)
}
