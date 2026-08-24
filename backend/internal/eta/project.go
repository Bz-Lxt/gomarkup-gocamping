package eta

import "gocamping/internal/geo"

type Projection struct {
	Lat      float64
	Lon      float64
	AlongM   float64
	RemainM  float64
	AvgSlope float64
}

func Project(lat, lon float64, path [][2]float64, elev []float64) Projection {
	plat, plon, along, remain := geo.ProjectOnto(lat, lon, path)
	p := Projection{Lat: plat, Lon: plon, AlongM: along, RemainM: remain}
	if len(elev) >= 2 && len(elev) == len(path) && remain > 0 {
		// average remaining slope using resampled elev
		total := geo.PathDistance(path)
		if total > 0 {
			startFrac := along / total
			i0 := int(startFrac * float64(len(elev)-1))
			if i0 < 0 {
				i0 = 0
			}
			if i0 >= len(elev)-1 {
				i0 = len(elev) - 2
			}
			de := elev[len(elev)-1] - elev[i0]
			p.AvgSlope = geo.SlopeRatio(remain, de)
		}
	}
	return p
}
