package risk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

func TestDispersionAndAutoSOS(t *testing.T) {
	pts := [][2]float64{{30.15, 118.86}, {30.1502, 118.8602}, {30.16, 118.87}}
	d := Dispersion(pts)
	require.Greater(t, d.Index, 0.0)
	cands := AutoSOSCandidates(pts, []int64{1, 2, 3}, 800)
	require.Contains(t, cands, int64(3))
}

func TestEvaluateLevels(t *testing.T) {
	r := 50.0
	dangers := []model.Waypoint{{ID: 1, Type: "danger", Lat: 30.15, Lon: 118.86, RiskWeight: 5, RadiusM: &r, Note: "滑坡"}}
	rep := Evaluate(30.1501, 118.8601, [][2]float64{{30.15, 118.86}, {30.151, 118.861}}, dangers, nil, timeutil.Now())
	require.NotEmpty(t, rep.Level)
	require.Greater(t, rep.Score, 0.0)
	require.NotEmpty(t, rep.Hits)
}

func TestThrottle(t *testing.T) {
	th := NewThrottle(time.Second)
	now := timeutil.Now()
	require.True(t, th.Allow(1, now))
	require.False(t, th.Allow(1, now))
}
