package geo

func PathDistance(pts [][2]float64) float64 {
	var d float64
	for i := 1; i < len(pts); i++ {
		d += HaversineM(pts[i-1][0], pts[i-1][1], pts[i][0], pts[i][1])
	}
	return d
}

// Resample walks a polyline and emits points every stepM metres.
func Resample(pts [][2]float64, stepM float64, maxN int) [][2]float64 {
	if len(pts) == 0 {
		return nil
	}
	if stepM <= 0 {
		stepM = 25
	}
	if maxN <= 0 {
		maxN = 2000
	}
	out := [][2]float64{pts[0]}
	remain := stepM
	for i := 1; i < len(pts) && len(out) < maxN; i++ {
		a, b := pts[i-1], pts[i]
		seg := HaversineM(a[0], a[1], b[0], b[1])
		if seg < 1e-6 {
			continue
		}
		br := BearingDeg(a[0], a[1], b[0], b[1])
		used := 0.0
		for remain <= seg-used && len(out) < maxN {
			used += remain
			lat, lon := Destination(a[0], a[1], br, used)
			out = append(out, [2]float64{lat, lon})
			remain = stepM
		}
		remain -= (seg - used)
	}
	last := pts[len(pts)-1]
	if out[len(out)-1] != last && len(out) < maxN {
		out = append(out, last)
	}
	return out
}

// ProjectOnto returns the closest point on the polyline and remaining distance to the end.
func ProjectOnto(lat, lon float64, path [][2]float64) (projLat, projLon, alongM, remainM float64) {
	if len(path) == 0 {
		return lat, lon, 0, 0
	}
	if len(path) == 1 {
		return path[0][0], path[0][1], 0, 0
	}
	bestD := 1e18
	bestAlong := 0.0
	bestLat, bestLon := path[0][0], path[0][1]
	prefix := 0.0
	total := PathDistance(path)
	for i := 1; i < len(path); i++ {
		a, b := path[i-1], path[i]
		plat, plon, t := closestOnSeg(lat, lon, a[0], a[1], b[0], b[1])
		d := HaversineM(lat, lon, plat, plon)
		seg := HaversineM(a[0], a[1], b[0], b[1])
		along := prefix + t*seg
		if d < bestD {
			bestD = d
			bestAlong = along
			bestLat, bestLon = plat, plon
		}
		prefix += seg
	}
	return bestLat, bestLon, bestAlong, total - bestAlong
}

func closestOnSeg(lat, lon, aLat, aLon, bLat, bLon float64) (float64, float64, float64) {
	// Local ENU approximation for projection parameter t.
	midLat := Deg2Rad((aLat + bLat) / 2)
	ax := Deg2Rad(aLon-lon) * mathCos(midLat) * EarthRadiusM
	ay := Deg2Rad(aLat-lat) * EarthRadiusM
	bx := Deg2Rad(bLon-lon) * mathCos(midLat) * EarthRadiusM
	by := Deg2Rad(bLat-lat) * EarthRadiusM
	// vectors from A: P is origin (0,0) in this frame? Use A as origin.
	px := -ax
	py := -ay
	vx := bx - ax
	vy := by - ay
	den := vx*vx + vy*vy
	if den < 1e-6 {
		return aLat, aLon, 0
	}
	t := (px*vx + py*vy) / den
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return aLat + t*(bLat-aLat), aLon + t*(bLon-aLon), t
}

func mathCos(x float64) float64 {
	return _cos(x)
}
