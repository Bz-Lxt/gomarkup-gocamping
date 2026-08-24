package model

import "time"

const (
	TripDraft    = "draft"
	TripActive   = "active"
	TripPaused   = "paused"
	TripFinished = "finished"
)

type Trip struct {
	ID         int64      `json:"id"`
	TeamID     int64      `json:"team_id"`
	RouteID    *int64     `json:"route_id"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func CanTransit(from, to string) bool {
	switch from {
	case TripDraft:
		return to == TripActive
	case TripActive:
		return to == TripPaused || to == TripFinished
	case TripPaused:
		return to == TripActive || to == TripFinished
	default:
		return false
	}
}
