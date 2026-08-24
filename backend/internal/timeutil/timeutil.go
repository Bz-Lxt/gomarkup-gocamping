package timeutil

import "time"

// Beijing is GMT+8. All persisted and displayed timestamps use this zone.
var Beijing = time.FixedZone("CST", 8*3600)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func NowNaive() time.Time {
	return Now().Truncate(time.Microsecond)
}

func ToBeijing(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.In(Beijing)
}

func ParseISO(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return ToBeijing(t), nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, Beijing); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02T15:04:05", s, Beijing)
}

func FormatDisplay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return ToBeijing(t).Format("2006-01-02 15:04:05")
}

func FormatISO(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return ToBeijing(t).Format(time.RFC3339)
}

func CivilDate(t time.Time) (y int, m time.Month, d int) {
	return ToBeijing(t).Date()
}

func HoursUntilSunset(at time.Time, lon, lat float64) float64 {
	// Approximate solar sunset for mid-latitudes; no external weather API.
	t := ToBeijing(at)
	day := t.YearDay()
	decl := 23.44 * sind(360.0/365.0*(float64(day)-81))
	latRad := lat * 3.141592653589793 / 180
	declRad := decl * 3.141592653589793 / 180
	cosHA := -tandApprox(latRad) * tandApprox(declRad)
	if cosHA > 1 {
		return 0
	}
	if cosHA < -1 {
		return 16
	}
	ha := acosApprox(cosHA) * 180 / 3.141592653589793
	solarNoon := 12.0 - lon/15.0 + 8 // shift to GMT+8 civil clock roughly
	sunsetHour := solarNoon + ha/15.0
	nowHour := float64(t.Hour()) + float64(t.Minute())/60 + float64(t.Second())/3600
	left := sunsetHour - nowHour
	if left < 0 {
		return 0
	}
	if left > 16 {
		return 16
	}
	return left
}

func sind(deg float64) float64 {
	return sinApprox(deg * 3.141592653589793 / 180)
}

func tandApprox(rad float64) float64 {
	c := cosApprox(rad)
	if c == 0 {
		return 1e9
	}
	return sinApprox(rad) / c
}

func sinApprox(x float64) float64 {
	return _sin(x)
}

func cosApprox(x float64) float64 {
	return _cos(x)
}

func acosApprox(x float64) float64 {
	return _acos(x)
}
