package ws

// TestJoin registers a TestClient in a room so tests can observe broadcasts
// without a real websocket connection. It is only used by tests.
func (h *Hub) TestJoin(tripID int64, c *TestClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[tripID] == nil {
		h.rooms[tripID] = map[*Client]struct{}{}
	}
	// Wrap the TestClient in a real Client whose send channel is shared.
	rc := &Client{hub: h, send: c.Send, tripID: tripID}
	h.rooms[tripID][rc] = struct{}{}
}

// TestClient is a lightweight observer for Hub broadcasts in tests.
type TestClient struct {
	Send chan []byte
}
