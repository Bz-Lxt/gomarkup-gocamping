package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	_ = os.Unsetenv("APP_ENV")
	c := Load()
	require.Equal(t, "local", c.TileProvider)
	require.Equal(t, "synthetic", c.DEMProvider)
	require.Equal(t, "mock", c.NotifyProvider)
	require.NotEmpty(t, c.CORSOrigins)
}

func TestLoadProviders(t *testing.T) {
	t.Setenv("TILE_PROVIDER", "osm")
	t.Setenv("DEM_PROVIDER", "http")
	c := Load()
	require.Equal(t, "osm", c.TileProvider)
	require.Equal(t, "http", c.DEMProvider)
}
