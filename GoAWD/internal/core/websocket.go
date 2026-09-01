package core

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"goawd/internal/types"
	"goawd/internal/ws"
)

type Client struct {
	conn *ws.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client

	pending   map[string]bool
	pendingMu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client, 8),
		unregister: make(chan *Client, 8),
		pending:    make(map[string]bool),
	}
}

func (h *Hub) Run(ctx context.Context) {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				close(client.send)
				client.conn.Close()
			}
			return

		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected, total: %d", len(h.clients))
			}

		case <-ticker.C:
			h.flushPending()
		}
	}
}

func (h *Hub) Notify(msgType string) {
	h.pendingMu.Lock()
	defer h.pendingMu.Unlock()
	h.pending[msgType] = true
}

func (h *Hub) flushPending() {
	h.pendingMu.Lock()
	if len(h.pending) == 0 {
		h.pendingMu.Unlock()
		return
	}
	pending := h.pending
	h.pending = make(map[string]bool)
	h.pendingMu.Unlock()

	for msgType := range pending {
		msg := types.WSMessage{Operation: types.WSOpReload, Type: msgType}
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		for client := range h.clients {
			select {
			case client.send <- data:
			default:
			}
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	client.readPump(h)
}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteText(string(msg)); err != nil {
			break
		}
	}
}
