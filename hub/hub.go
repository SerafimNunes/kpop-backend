package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB para aguentar chunks de áudio
)

type SubtitleMessage struct {
	LiveID    uint   `json:"live_id"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	IsFinal   bool   `json:"is_final"`
}

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	LiveID uint
}

type Hub struct {
	Rooms      map[uint]map[*Client]bool
	Broadcast  chan SubtitleMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[uint]map[*Client]bool),
		Broadcast:  make(chan SubtitleMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.LiveID] == nil {
				h.Rooms[client.LiveID] = make(map[*Client]bool)
			}
			h.Rooms[client.LiveID][client] = true
			h.mu.Unlock()
			log.Printf("Live %d: Novo espectador conectado.", client.LiveID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Rooms[client.LiveID][client]; ok {
				delete(h.Rooms[client.LiveID], client)
				close(client.Send)
				if len(h.Rooms[client.LiveID]) == 0 {
					delete(h.Rooms, client.LiveID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.Lock()
			clients := h.Rooms[msg.LiveID]
			payload, _ := json.Marshal(msg)
			for client := range clients {
				select {
				case client.Send <- payload:
				default:
					close(client.Send)
					delete(h.Rooms[msg.LiveID], client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// WritePump envia mensagens do Hub para o navegador (Legendas e Pings)
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
