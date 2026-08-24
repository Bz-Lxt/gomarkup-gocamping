package geo

// Neighborhood summarizes S2-bucket occupancy around a query point.
type Neighborhood struct {
	Tokens     []string
	MemberIDs  []int64
	Count      int
	RadiusM    float64
	CenterLat  float64
	CenterLon  float64
}

func (g *Grid) Inspect(lat, lon, radiusM float64) Neighborhood {
	near := g.Nearby(lat, lon, radiusM)
	n := Neighborhood{
		Tokens:    CoverTokens(lat, lon, radiusM),
		MemberIDs: make([]int64, 0, len(near)),
		Count:     len(near),
		RadiusM:   radiusM,
		CenterLat: lat,
		CenterLon: lon,
	}
	for _, f := range near {
		n.MemberIDs = append(n.MemberIDs, f.MemberID)
	}
	return n
}

func FineCell(lat, lon float64) string {
	return CellToken(lat, lon, FineLevel)
}

func SameBucket(aLat, aLon, bLat, bLon float64) bool {
	return CellToken(aLat, aLon, BucketLevel) == CellToken(bLat, bLon, BucketLevel)
}
