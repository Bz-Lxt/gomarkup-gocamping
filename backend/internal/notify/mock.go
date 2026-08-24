package notify

import (
	"context"
	"sync"

	"gocamping/internal/logger"
)

type Mock struct {
	mu    sync.Mutex
	Events []Event
}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Name() string { return "mock" }

func (m *Mock) Send(_ context.Context, ev Event) error {
	m.mu.Lock()
	m.Events = append(m.Events, ev)
	m.mu.Unlock()
	logger.Info("notify.mock", "kind", ev.Kind, "channel", ev.Channel)
	return nil
}

func (m *Mock) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Events)
}
