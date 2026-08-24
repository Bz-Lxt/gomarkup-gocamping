package geo

// PointInRing reports whether (lat,lon) is inside a closed or open ring
// of [lat,lon] vertices using the even-odd ray-casting algorithm.
func PointInRing(lat, lon float64, ring [][2]float64) bool {
	n := len(ring)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		yi, xi := ring[i][0], ring[i][1]
		yj, xj := ring[j][0], ring[j][1]
		intersect := ((yi > lat) != (yj > lat)) &&
			(lon < (xj-xi)*(lat-yi)/(yj-yi+1e-18)+xi)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

func PointInCircle(lat, lon, clat, clon, radiusM float64) bool {
	return HaversineM(lat, lon, clat, clon) <= radiusM
}

func RingBBox(ring [][2]float64) BBox {
	if len(ring) == 0 {
		return BBox{}
	}
	minLat, maxLat := ring[0][0], ring[0][0]
	minLon, maxLon := ring[0][1], ring[0][1]
	for _, p := range ring[1:] {
		if p[0] < minLat {
			minLat = p[0]
		}
		if p[0] > maxLat {
			maxLat = p[0]
		}
		if p[1] < minLon {
			minLon = p[1]
		}
		if p[1] > maxLon {
			maxLon = p[1]
		}
	}
	return BBox{MinLat: minLat, MinLon: minLon, MaxLat: maxLat, MaxLon: maxLon}
}

func CircleBBox(lat, lon, radiusM float64) BBox {
	north, _ := Destination(lat, lon, 0, radiusM)
	south, _ := Destination(lat, lon, 180, radiusM)
	_, east := Destination(lat, lon, 90, radiusM)
	_, west := Destination(lat, lon, 270, radiusM)
	return BBox{MinLat: south, MinLon: west, MaxLat: north, MaxLon: east}
}
