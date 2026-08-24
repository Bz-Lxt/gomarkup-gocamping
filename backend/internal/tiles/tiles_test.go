package tiles

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalPNG(t *testing.T) {
	a := PNG(8, 216, 107)
	b := PNG(8, 216, 107)
	require.Greater(t, len(a), 100)
	require.Equal(t, a, b)
	require.Equal(t, a[0:8], []byte{137, 80, 78, 71, 13, 10, 26, 10})
}

func TestTileKey(t *testing.T) {
	require.NotEqual(t, Key(1, 2, 3), Key(1, 2, 4))
}
