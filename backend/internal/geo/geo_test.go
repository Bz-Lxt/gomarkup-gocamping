package geo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHaversineKnownDistance(t *testing.T) {
	// ~111.2 km per degree latitude
	d := HaversineM(30, 120, 31, 120)
	require.InDelta(t, 111200, d, 800)
}

func TestS2Cover1km(t *testing.T) {
	cells := CoverRadius(30.15, 118.86, 1000, BucketLevel, FineLevel, 20)
	require.NotEmpty(t, cells)
	require.LessOrEqual(t, len(cells), 20)
	tok := CellToken(30.15, 118.86, BucketLevel)
	require.NotEmpty(t, tok)
}

func TestGridNearby(t *testing.T) {
	g := NewGrid()
	g.Upsert(LiveFix{MemberID: 1, Lat: 30.1500, Lon: 118.8600})
	g.Upsert(LiveFix{MemberID: 2, Lat: 30.1520, Lon: 118.8620})
	g.Upsert(LiveFix{MemberID: 3, Lat: 31.0000, Lon: 120.0000})
	near := g.Nearby(30.1500, 118.8600, 1000)
	ids := map[int64]bool{}
	for _, f := range near {
		ids[f.MemberID] = true
	}
	require.True(t, ids[1])
	require.True(t, ids[2])
	require.False(t, ids[3])
}

func TestRTreeSearch(t *testing.T) {
	rt := NewRTree()
	for i := 0; i < 40; i++ {
		lat := 30.1 + float64(i)*0.01
		lon := 118.8 + float64(i%5)*0.01
		rt.Insert(RTItem{ID: int64(i + 1), Box: CircleBBox(lat, lon, 120)})
	}
	require.Equal(t, 40, rt.Len())
	require.Len(t, rt.All(), 40)
	hits := rt.Search(BBox{MinLat: 30.0, MinLon: 118.7, MaxLat: 30.6, MaxLon: 119.0})
	require.NotEmpty(t, hits)
	near := rt.Search(CircleBBox(30.15, 118.82, 2000))
	require.NotEmpty(t, near)
}

func TestPointInRing(t *testing.T) {
	sq := [][2]float64{{0, 0}, {0, 1}, {1, 1}, {1, 0}}
	require.True(t, PointInRing(0.5, 0.5, sq))
	require.False(t, PointInRing(1.5, 0.5, sq))
}

func TestResampleAndProject(t *testing.T) {
	path := [][2]float64{{30.12, 118.85}, {30.14, 118.87}, {30.15, 118.88}}
	rs := Resample(path, 25, 200)
	require.Greater(t, len(rs), 3)
	_, _, along, remain := ProjectOnto(30.13, 118.86, path)
	require.Greater(t, along, 0.0)
	require.Greater(t, remain, 0.0)
}

func TestValidateLatLon(t *testing.T) {
	require.False(t, ValidateLatLon(91, 10))
	require.False(t, ValidateLatLon(0, 0))
	require.True(t, ValidateLatLon(30.1, 118.8))
}
