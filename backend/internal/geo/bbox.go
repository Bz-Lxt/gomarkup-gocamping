package geo

type BBox struct {
	MinLat, MinLon, MaxLat, MaxLon float64
}

func (b BBox) Intersects(o BBox) bool {
	return b.MinLat <= o.MaxLat && b.MaxLat >= o.MinLat && b.MinLon <= o.MaxLon && b.MaxLon >= o.MinLon
}

func (b BBox) Contains(lat, lon float64) bool {
	return lat >= b.MinLat && lat <= b.MaxLat && lon >= b.MinLon && lon <= b.MaxLon
}

func (b BBox) Union(o BBox) BBox {
	if b.Empty() {
		return o
	}
	if o.Empty() {
		return b
	}
	out := b
	if o.MinLat < out.MinLat {
		out.MinLat = o.MinLat
	}
	if o.MinLon < out.MinLon {
		out.MinLon = o.MinLon
	}
	if o.MaxLat > out.MaxLat {
		out.MaxLat = o.MaxLat
	}
	if o.MaxLon > out.MaxLon {
		out.MaxLon = o.MaxLon
	}
	return out
}

func (b BBox) Empty() bool {
	return b.MinLat == 0 && b.MaxLat == 0 && b.MinLon == 0 && b.MaxLon == 0
}

func (b BBox) Area() float64 {
	if b.Empty() {
		return 0
	}
	return (b.MaxLat - b.MinLat) * (b.MaxLon - b.MinLon)
}

func (b BBox) Enlargement(o BBox) float64 {
	return b.Union(o).Area() - b.Area()
}
