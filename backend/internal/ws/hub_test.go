package ws

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThrottle1Hz(t *testing.T) {
	th := NewThrottle()
	now := time.Now()
	require.True(t, th.Allow(7, now))
	require.False(t, th.Allow(7, now.Add(200*time.Millisecond)))
	require.True(t, th.Allow(7, now.Add(time.Second)))
	require.True(t, th.Allow(8, now))
}

func TestRoomEmpty(t *testing.T) {
	h := NewHub()
	require.Equal(t, 0, h.RoomSize(1))
}
