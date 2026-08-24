package replay

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

func TestBuildFrames(t *testing.T) {
	t0 := timeutil.Now().Add(-10 * time.Minute)
	pts := []model.TrackPoint{
		{MemberID: 1, Lat: 30, Lon: 118, RecordedAt: t0},
		{MemberID: 2, Lat: 30.01, Lon: 118.01, RecordedAt: t0.Add(30 * time.Second)},
		{MemberID: 1, Lat: 30.02, Lon: 118.02, RecordedAt: t0.Add(70 * time.Second)},
	}
	frames := Build(pts, 30*time.Second)
	require.GreaterOrEqual(t, len(frames), 2)
	require.NotEmpty(t, frames[len(frames)-1].Fixes)
}
