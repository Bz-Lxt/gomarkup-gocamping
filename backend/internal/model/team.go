package model

import "time"

type Team struct {
	ID         int64        `json:"id"`
	LeaderID   int64        `json:"leader_id"`
	RouteID    *int64       `json:"route_id"`
	Name       string       `json:"name"`
	InviteCode string       `json:"invite_code"`
	Status     string       `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	Members    []TeamMember `json:"members"`
}

type TeamMember struct {
	ID       int64     `json:"id"`
	TeamID   int64     `json:"team_id"`
	UserID   int64     `json:"user_id"`
	Nickname string    `json:"nickname"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	State    string    `json:"state"`
}
