package dem

import "context"

type Sample struct {
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Elevation float64 `json:"elevation"`
}

type Provider interface {
	Name() string
	Lookup(ctx context.Context, pts [][2]float64) ([]Sample, error)
}

func New(kind, httpURL string) Provider {
	switch kind {
	case "http":
		return NewHTTP(httpURL)
	default:
		return NewSynthetic()
	}
}
