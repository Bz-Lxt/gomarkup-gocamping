package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTP struct {
	url    string
	client *http.Client
}

func NewHTTP(url string) *HTTP {
	return &HTTP{url: url, client: &http.Client{Timeout: 5 * time.Second}}
}

func (h *HTTP) Name() string { return "http" }

func (h *HTTP) Send(ctx context.Context, ev Event) error {
	if h.url == "" {
		return fmt.Errorf("notify http url empty")
	}
	raw, _ := json.Marshal(ev)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 422 {
		return fmt.Errorf("notify http auth/validation %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("notify http status %d", resp.StatusCode)
	}
	return nil
}
