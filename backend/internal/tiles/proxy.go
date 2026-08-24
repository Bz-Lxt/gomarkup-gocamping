package tiles

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// OSM is the real tile path. Only used when TILE_PROVIDER=osm.
// Auth/validation failures are not retried.
type OSM struct {
	client *http.Client
}

func NewOSM() *OSM {
	return &OSM{client: &http.Client{Timeout: 6 * time.Second}}
}

func (o *OSM) Fetch(z, x, y int) ([]byte, error) {
	url := fmt.Sprintf("https://tile.openstreetmap.org/%d/%d/%d.png", z, x, y)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gocamping/1.0 (offline-first demo)")
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 422 {
		return nil, fmt.Errorf("osm auth/validation %d", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("osm status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}
