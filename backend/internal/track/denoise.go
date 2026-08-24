package track

import (
	"math"
	"time"

	"gocamping/internal/geo"
)

type FilterStats struct {
	Accuracy int
	Speed    int
	Accel    int
	Still    int
	Kept     int
}

func Denoise(pts []ParsedPoint) ([]ParsedPoint, FilterStats) {
	if len(pts) == 0 {
		return pts, FilterStats{}
	}
	tmp := make([]ParsedPoint, 0, len(pts))
	var st FilterStats
	for _, p := range pts {
		if p.Accuracy != nil && *p.Accuracy > MaxAccuracyM {
			st.Accuracy++
			continue
		}
		tmp = append(tmp, p)
	}
	if len(tmp) == 0 {
		return tmp, st
	}
	speedOK := make([]ParsedPoint, 0, len(tmp))
	speedOK = append(speedOK, tmp[0])
	for i := 1; i < len(tmp); i++ {
		dt := tmp[i].RecordedAt.Sub(tmp[i-1].RecordedAt).Seconds()
		if dt <= 0 {
			// same second: keep later one by replacing if farther? drop duplicate time
			st.Speed++
			continue
		}
		dist := geo.HaversineM(tmp[i-1].Lat, tmp[i-1].Lon, tmp[i].Lat, tmp[i].Lon)
		kmh := (dist / dt) * 3.6
		if kmh > MaxHikeKmh {
			st.Speed++
			continue
		}
		if len(speedOK) >= 2 {
			prev := speedOK[len(speedOK)-1]
			pprev := speedOK[len(speedOK)-2]
			dt1 := prev.RecordedAt.Sub(pprev.RecordedAt).Seconds()
			dt2 := tmp[i].RecordedAt.Sub(prev.RecordedAt).Seconds()
			if dt1 > 0 && dt2 > 0 {
				v1 := geo.HaversineM(pprev.Lat, pprev.Lon, prev.Lat, prev.Lon) / dt1
				v2 := geo.HaversineM(prev.Lat, prev.Lon, tmp[i].Lat, tmp[i].Lon) / dt2
				acc := math.Abs(v2-v1) / ((dt1 + dt2) / 2)
				if acc > MaxAccelMS2 {
					st.Accel++
					continue
				}
			}
		}
		speedOK = append(speedOK, tmp[i])
	}
	folded := foldStill(speedOK, &st)
	st.Kept = len(folded)
	return folded, st
}

func foldStill(pts []ParsedPoint, st *FilterStats) []ParsedPoint {
	if len(pts) < 3 {
		return pts
	}
	out := make([]ParsedPoint, 0, len(pts))
	i := 0
	for i < len(pts) {
		cluster := []ParsedPoint{pts[i]}
		j := i + 1
		for j < len(pts) {
			if geo.HaversineM(pts[i].Lat, pts[i].Lon, pts[j].Lat, pts[j].Lon) <= StillRadiusM &&
				pts[j].RecordedAt.Sub(pts[i].RecordedAt) <= time.Duration(StillWindowSec)*time.Second {
				cluster = append(cluster, pts[j])
				j++
				continue
			}
			break
		}
		if len(cluster) >= 3 && cluster[len(cluster)-1].RecordedAt.Sub(cluster[0].RecordedAt) < time.Duration(StillWindowSec)*time.Second {
			out = append(out, cluster[0])
			st.Still += len(cluster) - 1
			i = j
			continue
		}
		out = append(out, pts[i])
		i++
	}
	return out
}
