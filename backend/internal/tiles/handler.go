package tiles

import (
	"bytes"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"gocamping/internal/logger"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	z, _ := strconv.Atoi(chi.URLParam(r, "z"))
	x, _ := strconv.Atoi(chi.URLParam(r, "x"))
	y, _ := strconv.Atoi(chi.URLParam(r, "y"))
	if z < 0 || z > 16 || x < 0 || y < 0 {
		http.Error(w, "bad tile", http.StatusBadRequest)
		return
	}
	body, err := resolve(z, x, y)
	fallback := false
	if err != nil || !validPNG(body) {
		logger.Warn("tile fallback local", "err", err, "z", z, "x", x, "y", y)
		body = PNG(z, x, y)
		fallback = true
	}
	w.Header().Set("Content-Type", "image/png")
	if fallback {
		// Short TTL for fallback tiles so the browser retries the
		// upstream provider and picks up the real tile once it
		// recovers, instead of caching the offline terrain for a day.
		w.Header().Set("Cache-Control", "public, max-age=300")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	}
	_, _ = io.Copy(w, bytes.NewReader(body))
}

func resolve(z, x, y int) ([]byte, error) {
	switch os.Getenv("TILE_PROVIDER") {
	case "osm":
		return NewOSM().Fetch(z, x, y)
	case "mapbox":
		return NewMapbox(os.Getenv("MAPBOX_TOKEN")).Fetch(z, x, y)
	default:
		return PNG(z, x, y), nil
	}
}

// validPNG reports whether b contains a complete, decodable PNG image.
//
// When an upstream tile provider (OSM, Mapbox) resets the connection
// after sending only the first few bytes, io.ReadAll returns the
// partial payload together with a read error.  The error alone is
// enough to trigger the local fallback, but we also validate the
// bytes themselves to catch the rarer case where the upstream closes
// cleanly with a truncated body (err == nil, but the PNG is still
// undecodable).  Serving such a short body would produce a white
// block on the map and — worse — be cached by the browser for a day.
func validPNG(b []byte) bool {
	if len(b) < 8 { // shorter than the PNG signature
		return false
	}
	_, err := png.Decode(bytes.NewReader(b))
	return err == nil
}
