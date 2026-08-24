package tiles

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Mapbox is the optional satellite provider. Requires MAPBOX_TOKEN.
type Mapbox struct {
	token  string
	client *http.Client
}

func NewMapbox(token string) *Mapbox {
	return &Mapbox{token: token, client: &http.Client{Timeout: 8 * time.Second}}
}

func (m *Mapbox) Fetch(z, x, y int) ([]byte, error) {
	if m.token == "" {
		return nil, fmt.Errorf("mapbox token empty")
	}
	url := fmt.Sprintf("https://api.mapbox.com/v4/mapbox.satellite/%d/%d/%d.png?access_token=%s", z, x, y, m.token)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 422 {
		return nil, fmt.Errorf("mapbox auth/validation %d", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("mapbox status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 2<<20))
}
