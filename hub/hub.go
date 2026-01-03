package hub

import (
	"sync"
)

// Message define a estrutura de dados que viaja via WebSocket
type Message struct {
	Type    string      `json:"type"` // "translation", "ad", "system", "vip_alert"
	Payload interface{} `json:"payload"`
	LiveID  string      `json:"live_id,omitempty"` // Identificador da live para o "tubo" correto
}

// Hub mantém o conjunto de clientes ativos e faz o broadcast das mensagens
type Hub struct {
	// Clientes conectados: o canal de mensagens é a chave
	Clients map[chan Message]bool

	// Mensagens que chegam para serem enviadas a todos
	Broadcast chan Message

	// Canais para registrar e remover clientes (thread-safe)
	Register   chan chan Message
	Unregister chan chan Message

	mu sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan Message),
		Register:   make(chan chan Message),
		Unregister: make(chan chan Message),
		Clients:    make(map[chan Message]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client <- message:
					// Mensagem enviada com sucesso
				default:
					// Se o buffer do cliente estiver cheio, desconecta para não travar o hub
					close(client)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}
