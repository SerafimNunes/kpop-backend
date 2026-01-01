package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// SubtitleMessage define a estrutura da legenda que trafega pelo WebSocket
type SubtitleMessage struct {
	LiveID    uint   `json:"live_id"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	IsFinal   bool   `json:"is_final"` // True quando a frase termina de ser processada
}

// Client representa um usuário conectado (Fã ou Moderador)
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	LiveID uint // A qual live este cliente pertence
}

// Hub gerencia as conexões ativas e a distribuição de mensagens em tempo real
type Hub struct {
	// Clientes registrados organizados por ID da Live: Rooms[LiveID][Client]
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
			log.Printf("Usuário entrou na Live %d. Total na sala: %d", client.LiveID, len(h.Rooms[client.LiveID]))

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
			log.Printf("Usuário saiu da Live %d", client.LiveID)

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

// ReadPump lê mensagens do WebSocket (ex: comandos do moderador)
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// Configurações de limite de leitura e timeout
	c.Conn.SetReadLimit(512 * 1024) // 512KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erro de leitura: %v", err)
			}
			break
		}
	}
}

// WritePump envia mensagens do Hub para o dispositivo do fã/moderador
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
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
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
