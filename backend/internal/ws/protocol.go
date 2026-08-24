package ws

import "encoding/json"

const (
	TypePos      = "pos_update"
	TypeState    = "member_state"
	TypeRisk     = "risk_report"
	TypeSOS      = "sos"
	TypeETA      = "eta_update"
	TypePing     = "ping"
	TypePong     = "pong"
	TypeWelcome  = "welcome"
	TypeTrack    = "track_merged"
)

type Message struct {
	Type   string          `json:"type"`
	TripID int64           `json:"trip_id"`
	Payload json.RawMessage `json:"payload"`
}

func MustRaw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	if b == nil {
		b = []byte("{}")
	}
	return b
}
