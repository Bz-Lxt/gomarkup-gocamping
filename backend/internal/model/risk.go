package model

import "time"

type RiskHit struct {
	WaypointID int64   `json:"waypoint_id"`
	Type       string  `json:"type"`
	Note       string  `json:"note"`
	Weight     int     `json:"weight"`
	DistanceM  float64 `json:"distance_m"`
}

type RiskReport struct {
	ID          int64     `json:"id"`
	TripID      int64     `json:"trip_id"`
	Level       string    `json:"level"`
	Score       float64   `json:"score"`
	Dispersion  float64   `json:"dispersion"`
	MaxSlope    float64   `json:"max_slope"`
	WaterDistM  float64   `json:"water_dist_m"`
	SunsetLeftH float64   `json:"sunset_left_h"`
	Hits        []RiskHit `json:"hits"`
	CentroidLat float64   `json:"centroid_lat"`
	CentroidLon float64   `json:"centroid_lon"`
	FarthestM   float64   `json:"farthest_m"`
	ComputedAt  time.Time `json:"computed_at"`
}

type ETAResult struct {
	RemainingM     float64 `json:"remaining_m"`
	AvgSpeedKmh    float64 `json:"avg_speed_kmh"`
	CorrectedKmh   float64 `json:"corrected_kmh"`
	ETA            string  `json:"eta"`
	ETALow         string  `json:"eta_low"`
	ETAHigh        string  `json:"eta_high"`
	Confidence     float64 `json:"confidence"`
	MovingSeconds  float64 `json:"moving_seconds"`
}
