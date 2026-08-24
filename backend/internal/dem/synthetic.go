package dem

import (
	"context"
	"math"
)

// Synthetic is a deterministic multi-octave value-noise DEM.
// Same (lat,lon) always yields the same elevation — mock legitimacy.
type Synthetic struct{}

func NewSynthetic() *Synthetic { return &Synthetic{} }

func (s *Synthetic) Name() string { return "synthetic" }

func (s *Synthetic) Lookup(_ context.Context, pts [][2]float64) ([]Sample, error) {
	out := make([]Sample, len(pts))
	for i, p := range pts {
		out[i] = Sample{Lat: p[0], Lon: p[1], Elevation: ElevationAt(p[0], p[1])}
	}
	return out, nil
}

func ElevationAt(lat, lon float64) float64 {
	// Centered around Zhejiang/Anhui hill country for a familiar hiking feel.
	base := 180 + 420*fbm(lat*18, lon*18, 5)
	ridge := 260 * math.Abs(math.Sin(lat*35)+math.Cos(lon*28))
	valley := 40 * math.Sin(lat*80) * math.Cos(lon*70)
	h := base + ridge + valley
	if h < 8 {
		h = 8
	}
	if h > 2200 {
		h = 2200
	}
	return math.Round(h*10) / 10
}

func fbm(x, y float64, oct int) float64 {
	var sum, amp, freq float64 = 0, 1, 1
	var norm float64
	for i := 0; i < oct; i++ {
		sum += amp * valueNoise(x*freq, y*freq)
		norm += amp
		amp *= 0.5
		freq *= 2.05
	}
	if norm == 0 {
		return 0
	}
	return sum / norm
}

func valueNoise(x, y float64) float64 {
	x0 := math.Floor(x)
	y0 := math.Floor(y)
	tx := fade(x - x0)
	ty := fade(y - y0)
	n00 := hash(int(x0), int(y0))
	n10 := hash(int(x0)+1, int(y0))
	n01 := hash(int(x0), int(y0)+1)
	n11 := hash(int(x0)+1, int(y0)+1)
	nx0 := lerp(n00, n10, tx)
	nx1 := lerp(n01, n11, tx)
	return lerp(nx0, nx1, ty)*2 - 1
}

func hash(x, y int) float64 {
	n := x*374761393 + y*668265263
	n = (n ^ (n >> 13)) * 1274126177
	n = n ^ (n >> 16)
	return float64(n&0xffff) / 65535.0
}

func fade(t float64) float64 { return t * t * (3 - 2*t) }
func lerp(a, b, t float64) float64 { return a + (b-a)*t }
