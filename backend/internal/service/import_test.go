package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gocamping/internal/model"
)

func TestImportGeoJSON(t *testing.T) {
	raw := []byte(`{"type":"FeatureCollection","features":[
	  {"type":"Feature","geometry":{"type":"Point","coordinates":[118.85,30.12]},"properties":{"type":"camp","note":"营地"}}
	]}`)
	wps, err := ImportGeoJSON(raw)
	require.NoError(t, err)
	require.Len(t, wps, 1)
	require.Equal(t, "camp", wps[0].Type)
	require.InDelta(t, 30.12, wps[0].Lat, 1e-9)
}

func TestImportRejectsBad(t *testing.T) {
	_, err := ImportGeoJSON([]byte(`{"type":"Nope"}`))
	require.Error(t, err)
	_, err = ImportGeoJSON([]byte(`{"type":"FeatureCollection","features":[{"geometry":{"type":"Point","coordinates":[200,30]},"properties":{}}]}`))
	require.Error(t, err)
}

func TestExportGPXContainsPoints(t *testing.T) {
	g := ExportGPX(&model.RouteBook{
		Title: "t",
		Waypoints: []model.Waypoint{{Type: "waypoint", Lat: 30.1, Lon: 118.8}},
	})
	require.Contains(t, g, "30.100000")
}
