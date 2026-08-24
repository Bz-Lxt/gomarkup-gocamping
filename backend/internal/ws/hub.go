package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"gocamping/internal/logger"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: map[int64]map[*Client]struct{}{}}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, userID, tripID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Warn("ws upgrade", "err", err)
		return
	}
	c := &Client{hub: h, conn: conn, send: make(chan []byte, 32), userID: userID, tripID: tripID}
	h.join(c)
	c.enqueue(Message{Type: TypeWelcome, TripID: tripID, Payload: MustRaw(map[string]any{"user_id": userID, "trip_id": tripID})})
	go c.writeLoop()
	c.readLoop()
}

func (h *Hub) join(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.tripID] == nil {
		h.rooms[c.tripID] = map[*Client]struct{}{}
	}
	h.rooms[c.tripID][c] = struct{}{}
}

func (h *Hub) leave(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rm, ok := h.rooms[c.tripID]; ok {
		delete(rm, c)
		if len(rm) == 0 {
			delete(h.rooms, c.tripID)
		}
	}
	close(c.send)
}

func (h *Hub) Broadcast(tripID int64, typ string, payload any) {
	raw, err := json.Marshal(Message{Type: typ, TripID: tripID, Payload: MustRaw(payload)})
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[tripID] {
		select {
		case c.send <- raw:
		default:
		}
	}
}

func (h *Hub) RoomSize(tripID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[tripID])
}

// Throttle is a 1Hz per-member gate for live position broadcasts.
type Throttle struct {
	mu   sync.Mutex
	last map[int64]time.Time
}

func NewThrottle() *Throttle { return &Throttle{last: map[int64]time.Time{}} }

func (t *Throttle) Allow(memberID int64, now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if prev, ok := t.last[memberID]; ok && now.Sub(prev) < time.Second {
		return false
	}
	t.last[memberID] = now
	return true
}
