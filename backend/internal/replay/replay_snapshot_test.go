package replay_test

import (
	"testing"
	"time"

	"gocamping/internal/model"
	"gocamping/internal/replay"
)

func TestBuildPreservesHistoricalFrameSnapshots(t *testing.T) {
	start := time.Date(2026, time.August, 26, 9, 0, 0, 0, time.UTC)
	points := []model.TrackPoint{
		{MemberID: 7, Lat: 30.1000, Lon: 118.8000, RecordedAt: start},
		{MemberID: 7, Lat: 30.3000, Lon: 118.9000, RecordedAt: start.Add(20 * time.Second)},
	}

	frames := replay.Build(points, 10*time.Second)
	if len(frames) != 3 {
		t.Fatalf("expected 3 replay frames, got %d", len(frames))
	}
	wantLat := []float64{30.1000, 30.1000, 30.3000}
	for i, want := range wantLat {
		if len(frames[i].Fixes) != 1 {
			t.Fatalf("frame %d: expected one member fix, got %d", i, len(frames[i].Fixes))
		}
		if got := frames[i].Fixes[0].Lat; got != want {
			t.Errorf("frame %d at %s: latitude = %.4f, want %.4f", i, frames[i].At.Format(time.RFC3339), got, want)
		}
	}
}
