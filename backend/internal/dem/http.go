package dem

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTP talks to an Open-Elevation compatible lookup API.
// Only used when DEM_PROVIDER=http. Auth/validation errors are never retried.
type HTTP struct {
	url    string
	client *http.Client
}

func NewHTTP(url string) *HTTP {
	if url == "" {
		url = "https://api.open-elevation.com/api/v1/lookup"
	}
	return &HTTP{url: url, client: &http.Client{Timeout: 8 * time.Second}}
}

func (h *HTTP) Name() string { return "http" }

type oeReq struct {
	Locations []oeLoc `json:"locations"`
}
type oeLoc struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
type oeResp struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Elevation float64 `json:"elevation"`
	} `json:"results"`
}

func (h *HTTP) Lookup(ctx context.Context, pts [][2]float64) ([]Sample, error) {
	if len(pts) == 0 {
		return []Sample{}, nil
	}
	const chunk = 80

	// Pre-slice into batches so each goroutine writes into its own fixed
	// index slot.  We keep the calls parallel but assemble the final
	// output in the original route order, independent of which batch
	// finishes first.  A supplier-side latency wobble (e.g. batch 2
	// completing ~280 ms before batch 1) must not reorder samples.
	batches := 0
	type slot struct {
		samples []Sample
		err     error
	}
	slots := make([]slot, (len(pts)+chunk-1)/chunk)
	var wg sync.WaitGroup
	for i := 0; i < len(pts); i += chunk {
		j := i + chunk
		if j > len(pts) {
			j = len(pts)
		}
		idx := batches
		batches++
		wg.Add(1)
		go func(idx int, batch [][2]float64) {
			defer wg.Done()
			part, err := h.lookupOnce(ctx, batch)
			slots[idx].samples = part
			slots[idx].err = err
		}(idx, pts[i:j])
	}
	wg.Wait()

	out := make([]Sample, 0, len(pts))
	for i := 0; i < batches; i++ {
		if slots[i].err != nil {
			return nil, slots[i].err
		}
		out = append(out, slots[i].samples...)
	}
	return out, nil
}

func (h *HTTP) lookupOnce(ctx context.Context, pts [][2]float64) ([]Sample, error) {
	reqBody := oeReq{Locations: make([]oeLoc, len(pts))}
	for i, p := range pts {
		reqBody.Locations[i] = oeLoc{Latitude: p[0], Longitude: p[1]}
	}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dem http: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 422 {
		return nil, fmt.Errorf("dem http auth/validation %d: %s", resp.StatusCode, string(body))
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("dem http transient %d", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("dem http status %d", resp.StatusCode)
	}
	var parsed oeResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("dem http decode: %w", err)
	}
	out := make([]Sample, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Sample{Lat: r.Latitude, Lon: r.Longitude, Elevation: r.Elevation})
	}
	return out, nil
}
