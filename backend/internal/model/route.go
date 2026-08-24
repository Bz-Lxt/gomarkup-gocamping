package model

import "time"

type RouteBook struct {
	ID           int64      `json:"id"`
	OwnerID      int64      `json:"owner_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Visibility   string     `json:"visibility"`
	Version      int        `json:"version"`
	DistanceM    float64    `json:"distance_m"`
	AscentM      float64    `json:"ascent_m"`
	GeometryHash string     `json:"geometry_hash"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Waypoints    []Waypoint `json:"waypoints"`
}

type Waypoint struct {
	ID         int64      `json:"id"`
	RouteID    int64      `json:"route_id"`
	Seq        int        `json:"seq"`
	Type       string     `json:"type"` // camp|water|danger|waypoint
	Lat        float64    `json:"lat"`
	Lon        float64    `json:"lon"`
	Elevation  *float64   `json:"elevation"`
	RadiusM    *float64   `json:"radius_m"`
	Polygon    [][2]float64 `json:"polygon"` // [lat,lon] rings
	RiskWeight int        `json:"risk_weight"`
	Note       string     `json:"note"`
}

func (w Waypoint) IsPath() bool {
	return w.Type == "waypoint" || w.Type == "camp" || w.Type == "water"
}
