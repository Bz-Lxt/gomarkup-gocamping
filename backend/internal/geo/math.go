package geo

import "math"

func _cos(x float64) float64 { return math.Cos(x) }

func SlopeRatio(distM, dElevM float64) float64 {
	if distM < 0.5 {
		return 0
	}
	return dElevM / distM
}

func Centroid(pts [][2]float64) (float64, float64) {
	if len(pts) == 0 {
		return 0, 0
	}
	var la, lo float64
	for _, p := range pts {
		la += p[0]
		lo += p[1]
	}
	n := float64(len(pts))
	return la / n, lo / n
}

func StdDevDistances(pts [][2]float64, clat, clon float64) (mean, std, farthest float64) {
	if len(pts) == 0 {
		return 0, 0, 0
	}
	ds := make([]float64, len(pts))
	var sum float64
	for i, p := range pts {
		ds[i] = HaversineM(clat, clon, p[0], p[1])
		sum += ds[i]
		if ds[i] > farthest {
			farthest = ds[i]
		}
	}
	mean = sum / float64(len(ds))
	var v float64
	for _, d := range ds {
		x := d - mean
		v += x * x
	}
	std = math.Sqrt(v / float64(len(ds)))
	return mean, std, farthest
}
