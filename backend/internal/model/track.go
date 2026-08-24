package model

import "time"

type TrackPoint struct {
	ID         int64     `json:"id"`
	TripID     int64     `json:"trip_id"`
	MemberID   int64     `json:"member_id"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
	Elevation  *float64  `json:"elevation"`
	Accuracy   *float64  `json:"accuracy"`
	Speed      *float64  `json:"speed"`
	RecordedAt time.Time `json:"recorded_at"`
	Source     string    `json:"source"`
	IsNoise    bool      `json:"is_noise"`
	Fingerprint string   `json:"fingerprint"`
}

type RawPoint struct {
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	Elevation  *float64 `json:"elevation"`
	Accuracy   *float64 `json:"accuracy"`
	Speed      *float64 `json:"speed"`
	RecordedAt string   `json:"recorded_at"`
}

type TrackSegment struct {
	ID        int64     `json:"id"`
	TripID    int64     `json:"trip_id"`
	MemberID  int64     `json:"member_id"`
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
	DistanceM float64   `json:"distance_m"`
	IsGap     bool      `json:"is_gap"`
}

type TrackBatch struct {
	ID          int64     `json:"id"`
	TripID      int64     `json:"trip_id"`
	MemberID    int64     `json:"member_id"`
	BatchHash   string    `json:"batch_hash"`
	PointCount  int       `json:"point_count"`
	Accepted    int       `json:"accepted"`
	Rejected    int       `json:"rejected"`
	ProcessedAt time.Time `json:"processed_at"`
}

type MergeResult struct {
	Accepted      int           `json:"accepted"`
	Rejected      int           `json:"rejected"`
	Duplicates    int           `json:"duplicates"`
	DistanceM     float64       `json:"distance_m"`
	MovingSeconds float64       `json:"moving_seconds"`
	AvgSpeedKmh   float64       `json:"avg_speed_kmh"`
	AscentM       float64       `json:"ascent_m"`
	Idempotent    bool          `json:"idempotent"`
	Segments      []TrackSegment `json:"segments"`
}

type LivePosition struct {
	MemberID   int64     `json:"member_id"`
	Nickname   string    `json:"nickname"`
	Lat        float64   `json:"lat"`
	Lon        float64   `json:"lon"`
	Elevation  float64   `json:"elevation"`
	RecordedAt time.Time `json:"recorded_at"`
	S2Cell     string    `json:"s2_cell"`
}
