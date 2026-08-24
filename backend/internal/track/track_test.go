package track

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gocamping/internal/geo"
	"gocamping/internal/model"
	"gocamping/internal/simulator"
	"gocamping/internal/timeutil"
)

func samplePath() [][2]float64 {
	return [][2]float64{{30.124, 118.852}, {30.1315, 118.861}, {30.142, 118.874}, {30.1488, 118.8812}}
}

func TestValidateAndFingerprint(t *testing.T) {
	raw := []model.RawPoint{{Lat: 30.1, Lon: 118.8, RecordedAt: timeutil.FormatISO(timeutil.Now().Add(-time.Minute))}}
	pts, err := ValidateBatch(raw)
	require.NoError(t, err)
	pts = AttachFingerprints(7, pts)
	require.Len(t, pts[0].Fingerprint, 64)
	_, err = ValidateBatch([]model.RawPoint{{Lat: 99, Lon: 1, RecordedAt: timeutil.FormatISO(timeutil.Now())}})
	require.Error(t, err)
}

func TestDenoiseDropsSpeedSpikes(t *testing.T) {
	start := timeutil.Now().Add(-time.Hour)
	clean := simulator.Walk(samplePath(), start, 40, simulator.Options{SpeedKmh: 3.5})
	noisy := simulator.InjectOutliers(clean, 6, 800)
	require.Greater(t, len(noisy), len(clean))
	parsed, err := ValidateBatch(noisy)
	require.NoError(t, err)
	out, st := Denoise(parsed)
	require.GreaterOrEqual(t, float64(len(parsed)-len(out)), 0.3*float64(len(parsed))*0.5) // at least some drop
	require.Greater(t, st.Speed+st.Accel, 0)
	require.Less(t, len(out), len(parsed))
	drop := 1 - float64(len(out))/float64(len(parsed))
	require.GreaterOrEqual(t, drop, 0.10)
}

func TestDenoiseSyntheticDrop30(t *testing.T) {
	start := timeutil.Now().Add(-2 * time.Hour)
	base := simulator.Walk(samplePath(), start, 60, simulator.Options{NoiseSigmaM: 4, SpeedKmh: 3.2})
	// add many still-cluster duplicates + teleports
	var raw []model.RawPoint
	for i, p := range base {
		raw = append(raw, p)
		if i%5 == 0 {
			raw = append(raw, p, p) // still cluster
		}
		if i%7 == 0 {
			raw = append(raw, simulator.Spike(p, 600))
		}
	}
	parsed, err := ValidateBatch(raw)
	require.NoError(t, err)
	out, _ := Denoise(parsed)
	drop := 1 - float64(len(out))/float64(len(parsed))
	require.GreaterOrEqual(t, drop, 0.30, "drop=%.2f n=%d->%d", drop, len(parsed), len(out))
}

func TestSmoothRMSE(t *testing.T) {
	start := timeutil.Now().Add(-time.Hour)
	truthRaw := simulator.Walk(samplePath(), start, 50, simulator.Options{SpeedKmh: 3.5})
	truth, err := ValidateBatch(truthRaw)
	require.NoError(t, err)
	noisy := make([]ParsedPoint, len(truth))
	copy(noisy, truth)
	for i := range noisy {
		if i%2 == 0 {
			noisy[i].Lat += 0.00035
			noisy[i].Lon -= 0.00028
		} else {
			noisy[i].Lat -= 0.00031
			noisy[i].Lon += 0.00033
		}
	}
	sm := Smooth(noisy)
	require.Equal(t, len(noisy), len(sm))
	before := RMSE(truth, noisy)
	after := RMSE(truth, sm)
	require.Greater(t, before, 0.0)
	require.Less(t, after, before*0.80)
}

func TestMergeIdempotentAndGap(t *testing.T) {
	start := timeutil.Now().Add(-time.Hour)
	raw := simulator.Walk(samplePath(), start, 20, simulator.Options{SpeedKmh: 3.5})
	a, err := ValidateBatch(raw)
	require.NoError(t, err)
	a = AttachFingerprints(1, a)
	merged, segs := MergeIncremental(nil, a)
	require.Len(t, merged, len(a))
	again, _ := MergeIncremental(merged, a)
	require.Len(t, again, len(a))
	// gap: jump 10 minutes
	late := a[len(a)-1]
	late.RecordedAt = late.RecordedAt.Add(12 * time.Minute)
	late.Lat += 0.002
	late.Fingerprint = Fingerprint(1, late.Lat, late.Lon, late.RecordedAt)
	_, segs = MergeIncremental(merged, []ParsedPoint{late})
	hasGap := false
	for _, s := range segs {
		if s.IsGap {
			hasGap = true
		}
	}
	require.True(t, hasGap)
}

func TestDouglasPeucker(t *testing.T) {
	var pts []ParsedPoint
	t0 := timeutil.Now()
	for i := 0; i < 30; i++ {
		pts = append(pts, ParsedPoint{Lat: 30 + float64(i)*0.0001, Lon: 118, RecordedAt: t0.Add(time.Duration(i) * time.Second)})
	}
	out := DouglasPeucker(pts, 8)
	require.Less(t, len(out), len(pts))
	require.GreaterOrEqual(t, len(out), 2)
}

func TestMetricsAndRecentSpeed(t *testing.T) {
	start := timeutil.Now().Add(-4 * time.Hour)
	raw := simulator.Walk(samplePath(), start, 30, simulator.Options{SpeedKmh: 3.6})
	pts, err := ValidateBatch(raw)
	require.NoError(t, err)
	m := ComputeMetrics(pts)
	require.Greater(t, m.DistanceM, 100.0)
	require.InDelta(t, 3.6, m.AvgSpeedKmh, 1.5)
	require.Greater(t, RecentSpeedKmh(pts, 900), 1.0)
}

func TestBatchHashStable(t *testing.T) {
	raw := []model.RawPoint{{Lat: 1, Lon: 2, RecordedAt: "t"}}
	require.Equal(t, BatchHash(1, 2, raw), BatchHash(1, 2, raw))
	require.NotEqual(t, BatchHash(1, 2, raw), BatchHash(1, 3, raw))
}

func TestValidateTooMany(t *testing.T) {
	raw := make([]model.RawPoint, MaxBatchPoints+1)
	for i := range raw {
		raw[i] = model.RawPoint{Lat: 30, Lon: 118, RecordedAt: timeutil.FormatISO(timeutil.Now().Add(-time.Duration(i) * time.Second))}
	}
	_, err := ValidateBatch(raw)
	require.Error(t, err)
	_ = fmt.Sprintf("ok")
	_ = geo.ValidateLatLon(1, 2)
}
