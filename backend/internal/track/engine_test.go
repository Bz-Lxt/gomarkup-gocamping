package track

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gocamping/internal/simulator"
	"gocamping/internal/timeutil"
)

func TestPipelineIdempotentHash(t *testing.T) {
	start := timeutil.Now().Add(-3 * time.Hour)
	raw := simulator.Walk([][2]float64{{30.12, 118.85}, {30.15, 118.88}}, start, 24, simulator.Options{SpeedKmh: 3.4, NoiseSigmaM: 6})
	raw = simulator.InjectOutliers(raw, 5, 500)
	p := Pipeline{MemberID: 9, TripID: 3}
	a, err := p.Run(raw, nil, EmptyKnown())
	require.NoError(t, err)
	b, err := p.Run(raw, a.Merged, fingerprintsOf(a.Merged))
	require.NoError(t, err)
	require.Equal(t, a.BatchHash, b.BatchHash)
	require.GreaterOrEqual(t, len(b.Merged), len(a.Merged))
	require.Greater(t, a.Metrics.DistanceM, 0.0)
}

func fingerprintsOf(pts []ParsedPoint) map[string]struct{} {
	m := map[string]struct{}{}
	for _, p := range pts {
		if p.Fingerprint != "" {
			m[p.Fingerprint] = struct{}{}
		}
	}
	return m
}
