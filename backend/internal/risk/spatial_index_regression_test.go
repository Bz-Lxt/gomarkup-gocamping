package risk_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gocamping/internal/model"
	"gocamping/internal/risk"
)

func TestEvaluateFindsEveryIndexedDanger(t *testing.T) {
	const side = 12
	dangers := make([]model.Waypoint, 0, side*side)
	for row := 0; row < side; row++ {
		for col := 0; col < side; col++ {
			dangers = append(dangers, model.Waypoint{
				ID:         int64(len(dangers) + 1),
				Type:       "danger",
				Lat:        25 + float64(row)*0.1,
				Lon:        100 + float64(col)*0.1,
				RiskWeight: 5,
			})
		}
	}

	now := time.Date(2026, time.August, 24, 12, 0, 0, 0, time.UTC)
	for _, danger := range dangers {
		report := risk.Evaluate(
			danger.Lat,
			danger.Lon,
			[][2]float64{{danger.Lat, danger.Lon}},
			dangers,
			nil,
			now,
		)
		found := false
		for _, hit := range report.Hits {
			if hit.WaypointID == danger.ID {
				found = true
				break
			}
		}
		require.Truef(t, found, "danger %d at its center was omitted from hits", danger.ID)
	}
}
