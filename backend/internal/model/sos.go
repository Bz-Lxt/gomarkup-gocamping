package model

import "time"

type SOSEvent struct {
	ID         int64      `json:"id"`
	TripID     int64      `json:"trip_id"`
	MemberID   int64      `json:"member_id"`
	Nickname   string     `json:"nickname,omitempty"`
	Type       string     `json:"type"`
	Lat        float64    `json:"lat"`
	Lon        float64    `json:"lon"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
}
