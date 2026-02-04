package events

import (
	"context"
	"log"
	"sync"
)

// Message segue o padrão de comunicação da Auren Platform
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client representa uma conexão ativa (Engenheiro, Auditor ou Painel)
type Client struct {
	ID   string
	Conn chan Message
}

// Hub centraliza a comunicação em tempo real
type Hub struct {
	Clients             map[string]*Client
	Broadcast           chan Message
	Register            chan *Client
	Unregister          chan *Client
	ActiveSessionCancel context.CancelFunc
	Mu                  sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	log.Println("📡 [HUB] Motor de eventos em tempo real iniciado.")
	for {
		select {
		case c := <-h.Register:
			h.Mu.Lock()
			h.Clients[c.ID] = c
			h.Mu.Unlock()
			log.Printf("🔌 [HUB] Novo terminal conectado: %s (Total: %d)", c.ID, len(h.Clients))

		case c := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[c.ID]; ok {
				delete(h.Clients, c.ID)
				close(c.Conn)
				log.Printf("🔌 [HUB] Terminal desconectado: %s", c.ID)
			}
			h.Mu.Unlock()

		case m := <-h.Broadcast:
			h.Mu.RLock()
			if m.Type == "error" {
				log.Printf("⚠️ [HUB-BROADCAST] Alerta enviado: %v", m.Payload)
			}

			for id, c := range h.Clients {
				select {
				case c.Conn <- m:
				default:
					log.Printf("🧹 [HUB] Removendo cliente lento/travado: %s", id)
					h.Mu.RUnlock() // Unlock before lock to avoid deadlock
					h.Mu.Lock()
					delete(h.Clients, id)
					close(c.Conn)
					h.Mu.Unlock()
					h.Mu.RLock() // Re-lock for the loop
				}
			}
			h.Mu.RUnlock()
		}
	}
}
