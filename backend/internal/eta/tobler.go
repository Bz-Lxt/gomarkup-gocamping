package eta

import "math"

// Tobler hiking function (simplified):
// V = 5 * exp(-3.5 * |slope + 0.05|)  km/h
func ToblerKmh(slope float64) float64 {
	v := 5 * math.Exp(-3.5*math.Abs(slope+0.05))
	if v < 0.4 {
		return 0.4
	}
	if v > 7 {
		return 7
	}
	return v
}

func CorrectSpeed(recentKmh, slope float64) float64 {
	base := recentKmh
	if base < 0.3 {
		base = 3.2
	}
	if base > 8 {
		base = 8
	}
	factor := ToblerKmh(slope) / ToblerKmh(0)
	out := base * factor
	if out < 0.4 {
		return 0.4
	}
	return out
}
