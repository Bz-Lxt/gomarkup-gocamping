package geo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNeighborhoodInspect(t *testing.T) {
	g := NewGrid()
	g.Upsert(LiveFix{MemberID: 1, Lat: 30.1500, Lon: 118.8600})
	g.Upsert(LiveFix{MemberID: 2, Lat: 30.1504, Lon: 118.8603})
	n := g.Inspect(30.1500, 118.8600, 800)
	require.Equal(t, 2, n.Count)
	require.NotEmpty(t, n.Tokens)
	require.True(t, SameBucket(30.1500, 118.8600, 30.1501, 118.8601))
	require.NotEmpty(t, FineCell(30.15, 118.86))
}

func TestCircleAndRingBBox(t *testing.T) {
	b := CircleBBox(30.15, 118.86, 500)
	require.True(t, b.Contains(30.15, 118.86))
	require.False(t, b.Empty())
	ring := [][2]float64{{30.1, 118.8}, {30.2, 118.8}, {30.2, 118.9}, {30.1, 118.9}}
	rb := RingBBox(ring)
	require.True(t, rb.Contains(30.15, 118.85))
}
