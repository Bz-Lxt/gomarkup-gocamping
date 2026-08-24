package track

// Kalman1D is a constant-velocity Kalman filter used independently on lat and lon.
type Kalman1D struct {
	x    float64 // position
	v    float64 // velocity (units / second)
	p00  float64
	p01  float64
	p10  float64
	p11  float64
	qPos float64
	qVel float64
	r    float64
	init bool
}

func NewKalman(r float64) *Kalman1D {
	if r <= 0 {
		r = 8e-7 // ~3m in degrees at mid lat
	}
	return &Kalman1D{p00: 1, p11: 1, qPos: 2e-8, qVel: 4e-8, r: r}
}

func (k *Kalman1D) Update(z float64, dt float64) float64 {
	if !k.init {
		k.x = z
		k.init = true
		return z
	}
	if dt <= 0 {
		dt = 1
	}
	// predict
	k.x = k.x + k.v*dt
	k.p00 = k.p00 + (k.p01+k.p10)*dt + k.p11*dt*dt + k.qPos
	k.p01 = k.p01 + k.p11*dt
	k.p10 = k.p10 + k.p11*dt
	k.p11 = k.p11 + k.qVel
	// update
	y := z - k.x
	s := k.p00 + k.r
	k0 := k.p00 / s
	k1 := k.p10 / s
	k.x = k.x + k0*y
	k.v = k.v + k1*y
	p00, p01, p10, p11 := k.p00, k.p01, k.p10, k.p11
	k.p00 = (1-k0)*p00
	k.p01 = (1-k0)*p01
	k.p10 = p10 - k1*p00
	k.p11 = p11 - k1*p01
	return k.x
}

func Smooth(pts []ParsedPoint) []ParsedPoint {
	if len(pts) < 3 {
		return append([]ParsedPoint(nil), pts...)
	}
	klat := NewKalman(6e-7)
	klon := NewKalman(8e-7)
	out := make([]ParsedPoint, len(pts))
	for i, p := range pts {
		dt := 1.0
		if i > 0 {
			dt = p.RecordedAt.Sub(pts[i-1].RecordedAt).Seconds()
			if dt > 30 {
				dt = 30
			}
		}
		np := p
		np.Lat = klat.Update(p.Lat, dt)
		np.Lon = klon.Update(p.Lon, dt)
		out[i] = np
	}
	// moving-median fallback on remaining jitter (window 3)
	if len(out) >= 3 {
		med := append([]ParsedPoint(nil), out...)
		for i := 1; i < len(out)-1; i++ {
			med[i].Lat = median3(out[i-1].Lat, out[i].Lat, out[i+1].Lat)
			med[i].Lon = median3(out[i-1].Lon, out[i].Lon, out[i+1].Lon)
		}
		return med
	}
	return out
}

func median3(a, b, c float64) float64 {
	if a > b {
		a, b = b, a
	}
	if b > c {
		b, c = c, b
	}
	if a > b {
		a, b = b, a
	}
	return b
}

func RMSE(truth, est []ParsedPoint) float64 {
	n := len(truth)
	if n == 0 || n != len(est) {
		return 0
	}
	var s float64
	for i := 0; i < n; i++ {
		dlat := truth[i].Lat - est[i].Lat
		dlon := truth[i].Lon - est[i].Lon
		s += dlat*dlat + dlon*dlon
	}
	return s / float64(n)
}
