package tiles

import (
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
	if err != nil {
		logger.Warn("tile fallback local", "err", err, "z", z, "x", x, "y", y)
		body = PNG(z, x, y)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = io.Copy(w, bytesReader(body))
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

type br struct{ b []byte }

func bytesReader(b []byte) *br { return &br{b: b} }

func (r *br) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}
