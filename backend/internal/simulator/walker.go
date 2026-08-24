package simulator

import (
	"math"
	"time"

	"gocamping/internal/geo"
	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

type Options struct {
	NoiseSigmaM float64
	OfflineGap  bool
	SpeedKmh    float64
}

// Walk interpolates along a path at a constant hiking speed, optionally injecting
// Gaussian noise and a mid-route offline gap (missing points).
func Walk(path [][2]float64, start time.Time, n int, opt Options) []model.RawPoint {
	if n < 2 {
		n = 40
	}
	if opt.SpeedKmh <= 0 {
		opt.SpeedKmh = 3.6
	}
	if len(path) < 2 {
		return []model.RawPoint{}
	}
	rs := geo.Resample(path, 20, 2000)
	if len(rs) < 2 {
		return []model.RawPoint{}
	}
	step := float64(len(rs)-1) / float64(n-1)
	out := make([]model.RawPoint, 0, n)
	t := start
	mps := opt.SpeedKmh / 3.6
	var prevLat, prevLon float64
	for i := 0; i < n; i++ {
		idx := int(math.Round(float64(i) * step))
		if idx >= len(rs) {
			idx = len(rs) - 1
		}
		lat, lon := rs[idx][0], rs[idx][1]
		if opt.NoiseSigmaM > 0 {
			ang := float64((i*47)%360) * math.Pi / 180
			off := opt.NoiseSigmaM * (0.4 + 0.6*math.Mod(float64(i*13), 1.7) - 0.3)
			lat, lon = geo.Destination(lat, lon, ang*180/math.Pi, math.Abs(off))
		}
		if i > 0 {
			d := geo.HaversineM(prevLat, prevLon, lat, lon)
			sec := d / mps
			if sec < 1 {
				sec = 1
			}
			if sec > 180 {
				sec = 180
			}
			t = t.Add(time.Duration(sec * float64(time.Second)))
		}
		if opt.OfflineGap && i > n/3 && i < n/3+6 {
			prevLat, prevLon = lat, lon
			continue
		}
		acc := 8.0
		spd := opt.SpeedKmh / 3.6
		if t.After(timeutil.Now().Add(-30 * time.Second)) {
			t = timeutil.Now().Add(-30 * time.Second)
		}
		out = append(out, model.RawPoint{
			Lat: lat, Lon: lon, Accuracy: &acc, Speed: &spd,
			RecordedAt: timeutil.FormatISO(t),
		})
		prevLat, prevLon = lat, lon
	}
	return out
}

func Spike(base model.RawPoint, jumpM float64) model.RawPoint {
	lat, lon := geo.Destination(base.Lat, base.Lon, 90, jumpM)
	base.Lat, base.Lon = lat, lon
	return base
}
