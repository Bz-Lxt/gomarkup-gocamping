package geo

import (
	"strconv"
	"sync"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

const (
	BucketLevel = 13 // ~1.2km cells
	FineLevel   = 16
)

type Bucket struct {
	CellID string
	Level  int
}

func CellIDAt(lat, lon float64, level int) s2.CellID {
	ll := s2.LatLngFromDegrees(lat, lon)
	return s2.CellIDFromLatLng(ll).Parent(level)
}

func CellToken(lat, lon float64, level int) string {
	return CellIDAt(lat, lon, level).ToToken()
}

func CoverRadius(lat, lon, radiusM float64, minLevel, maxLevel, maxCells int) []s2.CellID {
	if maxCells <= 0 {
		maxCells = 16
	}
	if minLevel <= 0 {
		minLevel = BucketLevel
	}
	if maxLevel <= 0 {
		maxLevel = FineLevel
	}
	angle := s1.Angle(radiusM / EarthRadiusM)
	center := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
	cap := s2.CapFromCenterAngle(center, angle)
	coverer := &s2.RegionCoverer{MinLevel: minLevel, MaxLevel: maxLevel, MaxCells: maxCells}
	return coverer.Covering(cap)
}

func CoverTokens(lat, lon, radiusM float64) []string {
	// Normalize every covering cell to the bucket level so lookups hit Grid keys.
	cells := CoverRadius(lat, lon, radiusM, BucketLevel, FineLevel, 20)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(cells))
	for _, c := range cells {
		tok := c.Parent(BucketLevel).ToToken()
		if _, ok := seen[tok]; ok {
			continue
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	return out
}

// Grid is an in-memory spatial hash of live member positions keyed by S2 buckets.
type Grid struct {
	mu   sync.RWMutex
	buck map[string]map[int64]LiveFix
}

type LiveFix struct {
	MemberID int64
	Lat      float64
	Lon      float64
	Elev     float64
	UnixMs   int64
	Cell     string
}

func NewGrid() *Grid {
	return &Grid{buck: make(map[string]map[int64]LiveFix)}
}

func (g *Grid) Upsert(fix LiveFix) {
	cell := CellToken(fix.Lat, fix.Lon, BucketLevel)
	fix.Cell = cell
	g.mu.Lock()
	defer g.mu.Unlock()
	// remove previous bucket if moved
	for tok, m := range g.buck {
		if tok == cell {
			continue
		}
		if _, ok := m[fix.MemberID]; ok {
			delete(m, fix.MemberID)
			if len(m) == 0 {
				delete(g.buck, tok)
			}
		}
	}
	if g.buck[cell] == nil {
		g.buck[cell] = make(map[int64]LiveFix)
	}
	g.buck[cell][fix.MemberID] = fix
}

func (g *Grid) Nearby(lat, lon, radiusM float64) []LiveFix {
	tokens := CoverTokens(lat, lon, radiusM)
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]LiveFix, 0, 16)
	seen := map[int64]struct{}{}
	for _, tok := range tokens {
		for _, fx := range g.buck[tok] {
			if _, ok := seen[fx.MemberID]; ok {
				continue
			}
			if HaversineM(lat, lon, fx.Lat, fx.Lon) <= radiusM {
				seen[fx.MemberID] = struct{}{}
				out = append(out, fx)
			}
		}
	}
	return out
}

func (g *Grid) Members(ids []int64) []LiveFix {
	want := map[int64]struct{}{}
	for _, id := range ids {
		want[id] = struct{}{}
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]LiveFix, 0, len(ids))
	for _, m := range g.buck {
		for id, fx := range m {
			if _, ok := want[id]; ok {
				out = append(out, fx)
			}
		}
	}
	return out
}

func (g *Grid) Remove(memberID int64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for tok, m := range g.buck {
		delete(m, memberID)
		if len(m) == 0 {
			delete(g.buck, tok)
		}
	}
}

func TokenLevel(tok string) int {
	id := s2.CellIDFromToken(tok)
	if !id.IsValid() {
		return -1
	}
	return id.Level()
}

func TokenDebug(tok string) string {
	id := s2.CellIDFromToken(tok)
	if !id.IsValid() {
		return tok
	}
	return tok + "@" + strconv.Itoa(id.Level())
}
