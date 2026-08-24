package eta

import (
	"math"
	"time"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

func Estimate(remainM, recentKmh, slope float64, now time.Time) model.ETAResult {
	spd := CorrectSpeed(recentKmh, slope)
	hours := (remainM / 1000) / spd
	if hours < 0 {
		hours = 0
	}
	arr := timeutil.ToBeijing(now).Add(time.Duration(hours * float64(time.Hour)))
	conf := 0.55
	if recentKmh > 0.8 && remainM < 20000 {
		conf = 0.75
	}
	if recentKmh > 2 && remainM < 8000 {
		conf = 0.85
	}
	spread := 0.25 + (1-conf)*0.5
	lowH := hours * (1 - spread)
	highH := hours * (1 + spread)
	if lowH < 0 {
		lowH = 0
	}
	return model.ETAResult{
		RemainingM:    math.Round(remainM*10) / 10,
		AvgSpeedKmh:   math.Round(recentKmh*100) / 100,
		CorrectedKmh:  math.Round(spd*100) / 100,
		ETA:           timeutil.FormatDisplay(arr),
		ETALow:        timeutil.FormatDisplay(timeutil.ToBeijing(now).Add(time.Duration(lowH * float64(time.Hour)))),
		ETAHigh:       timeutil.FormatDisplay(timeutil.ToBeijing(now).Add(time.Duration(highH * float64(time.Hour)))),
		Confidence:    math.Round(conf*100) / 100,
		MovingSeconds: hours * 3600,
	}
}
