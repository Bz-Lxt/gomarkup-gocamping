package ws

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 8 * time.Second
	pongWait   = 45 * time.Second
	pingPeriod = 20 * time.Second
	maxMsg     = 32 << 10
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	tripID int64
}

func (c *Client) readLoop() {
	defer func() {
		c.hub.leave(c)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMsg)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if msg.Type == TypePing {
			c.enqueue(Message{Type: TypePong, TripID: c.tripID, Payload: MustRaw(map[string]string{"t": "pong"})})
		}
	}
}

func (c *Client) writeLoop() {
	tk := time.NewTicker(pingPeriod)
	defer func() {
		tk.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-tk.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) enqueue(msg Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.send <- b:
	default:
	}
}
