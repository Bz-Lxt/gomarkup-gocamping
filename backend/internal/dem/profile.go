package dem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"

	"gocamping/internal/geo"
)

type Profile struct {
	Provider    string    `json:"provider"`
	DistanceM   float64   `json:"distance_m"`
	AscentM     float64   `json:"ascent_m"`
	DescentM    float64   `json:"descent_m"`
	MinElev     float64   `json:"min_elev"`
	MaxElev     float64   `json:"max_elev"`
	AvgSlope    float64   `json:"avg_slope"`
	Samples     []Sample  `json:"samples"`
	AlongM      []float64 `json:"along_m"`
	Slopes      []float64 `json:"slopes"`
	GeometryHash string   `json:"geometry_hash"`
}

func GeometryHash(pts [][2]float64) string {
	h := sha256.New()
	for _, p := range pts {
		fmt.Fprintf(h, "%.6f,%.6f;", p[0], p[1])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func BuildProfile(ctx context.Context, p Provider, path [][2]float64) (*Profile, error) {
	if len(path) < 2 {
		return nil, fmt.Errorf("路线至少需要 2 个点")
	}
	rs := geo.Resample(path, 25, 2000)
	samples, err := p.Lookup(ctx, rs)
	if err != nil {
		return nil, err
	}
	prof := &Profile{
		Provider:     p.Name(),
		Samples:      samples,
		AlongM:       make([]float64, len(samples)),
		Slopes:       make([]float64, len(samples)),
		GeometryHash: GeometryHash(path),
		MinElev:      1e9,
		MaxElev:      -1e9,
	}
	var along float64
	for i, s := range samples {
		if s.Elevation < prof.MinElev {
			prof.MinElev = s.Elevation
		}
		if s.Elevation > prof.MaxElev {
			prof.MaxElev = s.Elevation
		}
		if i == 0 {
			continue
		}
		d := geo.HaversineM(samples[i-1].Lat, samples[i-1].Lon, s.Lat, s.Lon)
		along += d
		prof.AlongM[i] = along
		de := s.Elevation - samples[i-1].Elevation
		if de > 0.5 {
			prof.AscentM += de
		} else if de < -0.5 {
			prof.DescentM += -de
		}
		sl := geo.SlopeRatio(d, de)
		prof.Slopes[i] = sl
	}
	prof.DistanceM = along
	if along > 0 {
		prof.AvgSlope = math.Abs((prof.MaxElev - prof.MinElev) / along)
	}
	return prof, nil
}
