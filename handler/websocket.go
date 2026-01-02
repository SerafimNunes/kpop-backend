package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"kpop-backend/hub"
	"kpop-backend/translate"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(h *hub.Hub, svc *translate.GeminiService, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	liveIDStr := vars["id"]
	liveID, err := strconv.ParseUint(liveIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro upgrade WS: %v", err)
		return
	}

	processor := NewAudioProcessor(svc, h)
	client := &hub.Client{
		Hub:    h,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		LiveID: uint(liveID),
	}

	client.Hub.Register <- client

	// Mock de legendas para teste visual
	go processor.StartMockSubtitles(client.LiveID)

	// Inicia a bomba de escrita (Hub -> Navegador)
	go client.WritePump()

	// Inicia a bomba de leitura (Microfone -> Backend)
	go func() {
		defer func() {
			client.Hub.Unregister <- client
			conn.Close()
		}()

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Conexão encerrada pelo cliente: %v", err)
				break
			}

			if messageType == websocket.BinaryMessage {
				// Processa o chunk de áudio PCM vindo do celular
				go processor.ProcessAudioChunk(context.Background(), client.LiveID, p)
			}
		}
	}()
}

func ReverseTranslate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Mock até configurar a API real do Gemini para tradução reversa
	koreanText := "안녕하세요! (Refinado: " + req.Text + ")"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"korean": koreanText,
	})
}
