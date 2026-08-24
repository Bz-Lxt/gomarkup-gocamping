package gpx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTrack(t *testing.T) {
	raw := []byte(`<?xml version="1.0"?><gpx version="1.1"><trk><trkseg>
	<trkpt lat="30.12" lon="118.85"><ele>400</ele></trkpt>
	<trkpt lat="30.13" lon="118.86"></trkpt>
	</trkseg></trk></gpx>`)
	wps, err := Parse(raw)
	require.NoError(t, err)
	require.Len(t, wps, 2)
	require.InDelta(t, 30.12, wps[0].Lat, 1e-6)
}

func TestParseRejectsBad(t *testing.T) {
	_, err := Parse([]byte(`not gpx`))
	require.Error(t, err)
	_, err = Parse([]byte(`<gpx><trkpt lat="99" lon="1"/></gpx>`))
	require.Error(t, err)
}
