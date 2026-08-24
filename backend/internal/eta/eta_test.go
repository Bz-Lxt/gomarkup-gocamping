package eta

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gocamping/internal/timeutil"
)

func TestToblerBounds(t *testing.T) {
	require.Greater(t, ToblerKmh(0), 3.0)
	require.Less(t, ToblerKmh(0.3), ToblerKmh(0))
}

func TestEstimateAccuracy(t *testing.T) {
	// 3600m at 3.6km/h → 1 hour
	now := timeutil.Now()
	out := Estimate(3600, 3.6, 0, now)
	arr, err := time.ParseInLocation("2006-01-02 15:04:05", out.ETA, timeutil.Beijing)
	require.NoError(t, err)
	delta := arr.Sub(now.In(timeutil.Beijing)).Minutes()
	require.InDelta(t, 60, delta, 60*0.15)
	require.Greater(t, out.Confidence, 0.5)
}

func TestProjectRemain(t *testing.T) {
	path := [][2]float64{{30.0, 118.0}, {30.01, 118.01}, {30.02, 118.02}}
	p := Project(30.005, 118.005, path, nil)
	require.Greater(t, p.RemainM, 0.0)
	require.Greater(t, p.AlongM, 0.0)
}
