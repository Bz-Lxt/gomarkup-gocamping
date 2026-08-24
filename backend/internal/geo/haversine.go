package geo

import "math"

const EarthRadiusM = 6371000.0

func Deg2Rad(d float64) float64 { return d * math.Pi / 180 }
func Rad2Deg(r float64) float64 { return r * 180 / math.Pi }

func HaversineM(lat1, lon1, lat2, lon2 float64) float64 {
	p1 := Deg2Rad(lat1)
	p2 := Deg2Rad(lat2)
	dp := Deg2Rad(lat2 - lat1)
	dl := Deg2Rad(lon2 - lon1)
	a := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(math.Max(0, 1-a)))
	return EarthRadiusM * c
}

func BearingDeg(lat1, lon1, lat2, lon2 float64) float64 {
	y := math.Sin(Deg2Rad(lon2-lon1)) * math.Cos(Deg2Rad(lat2))
	x := math.Cos(Deg2Rad(lat1))*math.Sin(Deg2Rad(lat2)) - math.Sin(Deg2Rad(lat1))*math.Cos(Deg2Rad(lat2))*math.Cos(Deg2Rad(lon2-lon1))
	br := Rad2Deg(math.Atan2(y, x))
	if br < 0 {
		br += 360
	}
	return br
}

func Destination(lat, lon, bearingDeg, distM float64) (float64, float64) {
	d := distM / EarthRadiusM
	br := Deg2Rad(bearingDeg)
	p1 := Deg2Rad(lat)
	l1 := Deg2Rad(lon)
	p2 := math.Asin(math.Sin(p1)*math.Cos(d) + math.Cos(p1)*math.Sin(d)*math.Cos(br))
	l2 := l1 + math.Atan2(math.Sin(br)*math.Sin(d)*math.Cos(p1), math.Cos(d)-math.Sin(p1)*math.Sin(p2))
	return Rad2Deg(p2), Rad2Deg(l2)
}

func ValidateLatLon(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180 && !(lat == 0 && lon == 0)
}

func Quantize(v float64, places int) float64 {
	p := math.Pow10(places)
	return math.Round(v*p) / p
}
