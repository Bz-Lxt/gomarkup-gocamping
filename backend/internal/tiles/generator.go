package tiles

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync"

	"gocamping/internal/dem"
)

const tileSize = 256

var cache sync.Map

func Key(z, x, y int) uint64 {
	return uint64(z)<<48 | uint64(x)<<24 | uint64(y)
}

func PNG(z, x, y int) []byte {
	k := Key(z, x, y)
	if v, ok := cache.Load(k); ok {
		return v.([]byte)
	}
	img := render(z, x, y)
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	b := buf.Bytes()
	cache.Store(k, b)
	return b
}

func render(z, x, y int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
	n := 1 << z
	if n <= 0 {
		n = 1
	}
	for py := 0; py < tileSize; py++ {
		for px := 0; px < tileSize; px++ {
			lon := tileLon(float64(x)+float64(px)/tileSize, n)
			lat := tileLat(float64(y)+float64(py)/tileSize, n)
			h := dem.ElevationAt(lat, lon)
			c := terrainColor(h, lat, lon)
			// grid lines for orientation
			if px == 0 || py == 0 {
				c = color.RGBA{40, 55, 42, 255}
			}
			img.SetRGBA(px, py, c)
		}
	}
	return img
}

func tileLon(x float64, n int) float64 {
	return x/float64(n)*360 - 180
}

func tileLat(y float64, n int) float64 {
	nlat := math.Pi - 2*math.Pi*y/float64(n)
	return 180 / math.Pi * math.Atan(math.Sinh(nlat))
}

func terrainColor(h, lat, lon float64) color.RGBA {
	// dusk-forest palette: valley moss → ridge ochre → peak mist
	t := (h - 20) / 1600
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	shade := 0.85 + 0.15*math.Sin(lat*40)*math.Cos(lon*30)
	r := uint8(clamp((0.18+0.55*t)*255*shade, 0, 255))
	g := uint8(clamp((0.32+0.28*(1-t))*255*shade, 0, 255))
	b := uint8(clamp((0.22+0.18*t)*255*shade, 0, 255))
	return color.RGBA{r, g, b, 255}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
