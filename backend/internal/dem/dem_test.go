package dem

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyntheticDeterministic(t *testing.T) {
	a := ElevationAt(30.15, 118.86)
	b := ElevationAt(30.15, 118.86)
	require.Equal(t, a, b)
	require.Greater(t, a, 0.0)
}

func TestBuildProfile(t *testing.T) {
	path := [][2]float64{{30.124, 118.852}, {30.1315, 118.861}, {30.1488, 118.8812}}
	p, err := BuildProfile(context.Background(), NewSynthetic(), path)
	require.NoError(t, err)
	require.Greater(t, p.DistanceM, 1000.0)
	require.Equal(t, "synthetic", p.Provider)
	require.Equal(t, GeometryHash(path), p.GeometryHash)
	require.NotEmpty(t, p.Samples)
}

func TestHashChangesWithGeometry(t *testing.T) {
	a := GeometryHash([][2]float64{{1, 2}, {3, 4}})
	b := GeometryHash([][2]float64{{1, 2}, {3, 4.001}})
	require.NotEqual(t, a, b)
}
