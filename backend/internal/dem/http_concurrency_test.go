package dem_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"gocamping/internal/dem"
)

type lookupLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type lookupRequest struct {
	Locations []lookupLocation `json:"locations"`
}

type lookupResult struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Elevation float64 `json:"elevation"`
}

type lookupResponse struct {
	Results []lookupResult `json:"results"`
}

func TestHTTPLookupPreservesInputOrderAcrossConcurrentBatches(t *testing.T) {
	points := make([][2]float64, 160)
	for i := range points {
		points[i] = [2]float64{30 + float64(i)/10000, 118 + float64(i)/10000}
	}

	secondBatchWritten := make(chan struct{})
	var secondBatchOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req lookupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := lookupResponse{Results: make([]lookupResult, len(req.Locations))}
		for i, loc := range req.Locations {
			resp.Results[i] = lookupResult{
				Latitude:  loc.Latitude,
				Longitude: loc.Longitude,
				Elevation: loc.Latitude * 10,
			}
		}
		writeResponse := func() {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		if len(req.Locations) > 0 && req.Locations[0].Latitude == points[0][0] {
			select {
			case <-secondBatchWritten:
				time.Sleep(250 * time.Millisecond)
			case <-time.After(500 * time.Millisecond):
			}
			writeResponse()
			return
		}
		writeResponse()
		secondBatchOnce.Do(func() { close(secondBatchWritten) })
	}))
	t.Cleanup(server.Close)

	samples, err := dem.NewHTTP(server.URL).Lookup(context.Background(), points)
	if err != nil {
		t.Fatalf("Lookup returned an error: %v", err)
	}
	if len(samples) != len(points) {
		t.Fatalf("Lookup returned %d samples for %d points", len(samples), len(points))
	}
	for i, sample := range samples {
		if sample.Lat != points[i][0] || sample.Lon != points[i][1] {
			t.Fatalf("sample %d = (%f, %f), want (%f, %f)", i, sample.Lat, sample.Lon, points[i][0], points[i][1])
		}
	}
}
