package notify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockRecords(t *testing.T) {
	m := NewMock()
	require.Equal(t, "mock", m.Name())
	require.NoError(t, m.Send(context.Background(), Event{Channel: "sos", Kind: "manual"}))
	require.Equal(t, 1, m.Len())
}

func TestFactory(t *testing.T) {
	p := New("mock", "")
	require.Equal(t, "mock", p.Name())
	h := New("http", "http://127.0.0.1:9/hook")
	require.Equal(t, "http", h.Name())
}
