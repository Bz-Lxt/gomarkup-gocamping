package notify

import "context"

type Event struct {
	Channel string                 `json:"channel"`
	Kind    string                 `json:"kind"`
	Payload map[string]interface{} `json:"payload"`
}

type Provider interface {
	Name() string
	Send(ctx context.Context, ev Event) error
}

func New(kind, httpURL string) Provider {
	if kind == "http" {
		return NewHTTP(httpURL)
	}
	return NewMock()
}
