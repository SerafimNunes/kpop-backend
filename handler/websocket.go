package handler

import (
	"log"
	"net/http"
	"strconv"

	"kpop-backend/hub" // Certifique-se que o caminho do módulo está correto
	"kpop-backend/translate"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Configuração do Upgrader para permitir conexões de diferentes origens (CORS)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // No MVP permitimos todas, no futuro restringimos ao app
	},
}

// ServeWS lida com as requisições de WebSocket vindas do app ou web
func ServeWS(h *hub.Hub, svc *translate.GeminiService, w http.ResponseWriter, r *http.Request) {
	// 1. Extrai o LiveID da URL (ex: /ws/live/10)
	vars := mux.Vars(r)
	liveIDStr := vars["id"]
	liveID, err := strconv.ParseUint(liveIDStr, 10, 32)
	if err != nil {
		log.Printf("ID de live inválido: %v", err)
		return
	}

	// 2. Faz o Upgrade da conexão HTTP para WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro no upgrade do WebSocket: %v", err)
		return
	}

	// 3. Cria a instância do cliente
	client := &hub.Client{
		Hub:    h,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		LiveID: uint(liveID),
	}

	// 4. Registra no Hub
	client.Hub.Register <- client

	// 5. Inicia as rotinas de leitura e escrita
	go client.WritePump()
	go client.ReadPump()
}
