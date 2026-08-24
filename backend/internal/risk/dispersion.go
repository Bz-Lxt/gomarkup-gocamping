package risk

import "gocamping/internal/geo"

type Disp struct {
	Index    float64
	Mean     float64
	Std      float64
	Farthest float64
	Span     float64
}

func Dispersion(members [][2]float64) Disp {
	if len(members) == 0 {
		return Disp{}
	}
	clat, clon := geo.Centroid(members)
	mean, std, far := geo.StdDevDistances(members, clat, clon)
	// span: max pairwise as proxy of front-back stretch
	span := 0.0
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			d := geo.HaversineM(members[i][0], members[i][1], members[j][0], members[j][1])
			if d > span {
				span = d
			}
		}
	}
	// normalize: 800m std → 1.0, 1200m farthest → 1.0, 1500m span → 1.0
	idx := 0.45*clamp01(std/800) + 0.35*clamp01(far/1200) + 0.20*clamp01(span/1500)
	return Disp{Index: idx, Mean: mean, Std: std, Farthest: far, Span: span}
}

func AutoSOSCandidates(members [][2]float64, ids []int64, thresholdM float64) []int64 {
	if len(members) == 0 || len(members) != len(ids) {
		return nil
	}
	if thresholdM <= 0 {
		thresholdM = 800
	}
	clat, clon := geo.Centroid(members)
	var out []int64
	for i, p := range members {
		if geo.HaversineM(clat, clon, p[0], p[1]) >= thresholdM {
			out = append(out, ids[i])
		}
	}
	return out
}
