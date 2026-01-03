package handler

import (
	"context"
	"encoding/json"
	"k-lens/db"
	"k-lens/hub"
	"k-lens/media"
	"k-lens/models"
	"k-lens/translate"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     checkOrigin,
}

// checkOrigin valida a origem da requisição WebSocket para evitar ataques cross-site
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	host := r.Host

	// Se não houver Origin header, aceita (pode ser browser com SameSite)
	if origin == "" {
		return true
	}

	// Extrai host do origin (formato: http://host:port)
	if strings.HasSuffix(origin, "//"+host) || strings.HasSuffix(origin, "//localhost"+strings.TrimPrefix(host, "localhost")) {
		return true
	}

	log.Printf("⚠️ [WebSocket] Origem bloqueada: %s (Host: %s)", origin, host)
	return false
}

var (
	semaphore    = make(chan struct{}, 5)
	globalGemini *translate.GeminiService
	videoCutter  = media.NewCutter()
)

func ServeWS(h *hub.Hub, gemini *translate.GeminiService, w http.ResponseWriter, r *http.Request) {
	globalGemini = gemini
	vars := mux.Vars(r)
	liveIDStr := vars["id"]
	liveID, _ := strconv.ParseUint(liveIDStr, 10, 32)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro upgrade WS: %v", err)
		return
	}

	processor := NewAudioProcessor(500.0)
	clientChan := make(chan hub.Message, 256)
	h.Register <- clientChan
	startTime := time.Now()

	var currentLiveURL string

	defer func() {
		h.Unregister <- clientChan
		conn.Close()
	}()

	// Goroutine para escutar clipes concluídos e avisar o celular via HUB
	go func() {
		for clipName := range videoCutter.NotifyChan {
			msg := hub.Message{
				Type:    "CLIP_READY",
				Payload: "Clipe disponível",
				Url:     "/recordings/" + clipName, // Agora o campo Url existe no hub.Message
				LiveID:  liveIDStr,
			}
			h.Broadcast <- msg
		}
	}()

	// Goroutine de escrita (Servidor -> App)
	go func() {
		for message := range clientChan {
			if message.LiveID != "" && message.LiveID != liveIDStr {
				continue
			}
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteJSON(message); err != nil {
				return
			}
		}
	}()

	// Loop de leitura (App -> Servidor)
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Validação básica de tamanho de mensagem para evitar DoS
		if len(p) == 0 {
			continue
		}

		if messageType == websocket.TextMessage {
			var raw map[string]interface{}
			if err := json.Unmarshal(p, &raw); err != nil {
				log.Printf("⚠️ [WebSocket] JSON inválido: %v", err)
				continue
			}

			// Validação de campos obrigatórios
			if raw == nil || len(raw) == 0 {
				log.Printf("⚠️ [WebSocket] Mensagem vazia ou nula")
				continue
			}

			if raw["action"] == "update_config" {
				duration, _ := strconv.Atoi(interfaceToString(raw["duration"]))
				ratio := interfaceToString(raw["ratio"])
				url := interfaceToString(raw["live_url"])

				videoCutter.UpdateConfig(duration, ratio)
				currentLiveURL = url
				continue
			}

			if raw["type"] == "MANUAL_CLIP" {
				ratio := interfaceToString(raw["ratio"])
				url := interfaceToString(raw["url"])
				if url == "" {
					url = currentLiveURL
				}

				log.Printf("🕹️ [MANUAL] Solicitado corte em %s", ratio)
				videoCutter.UpdateConfig(61, ratio)

				milliOffset := time.Since(startTime).Milliseconds()
				go videoCutter.CreateClip(liveIDStr, url, float64(milliOffset), "manual_premium")

				h.Broadcast <- hub.Message{
					Type: "translation", Payload: "🎬 SOLICITANDO CORTE (" + ratio + ")...", LiveID: liveIDStr,
				}
				continue
			}
		}

		if messageType == websocket.BinaryMessage && gemini != nil {
			if len(p) < 100 || !processor.ShouldProcess(p) {
				continue
			}

			go func(audioData []byte) {
				select {
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				default:
					return
				}

				// Timeout aumentado para 30 segundos para melhor robustez
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				log.Printf("⏱️ [Gemini] Processando áudio com timeout de 30s")

				resultado, err := gemini.TranslateAudio(ctx, audioData)
				if err != nil {
					log.Printf("❌ [Gemini] Erro na tradução de áudio: %v", err)
					return
				}
				if resultado == "" {
					log.Printf("⚠️ [Gemini] Resposta vazia do Gemini")
					return
				}

				milliOffset := time.Since(startTime).Milliseconds()

				if db.DB != nil {
					db.DB.Create(&models.CaptionLog{
						LiveArchiveID: uint(liveID),
						Timestamp:     milliOffset,
						Text:          resultado,
					})
				}

				lowResult := strings.ToLower(resultado)
				if strings.Contains(lowResult, "💜") || strings.Contains(lowResult, "tchau") || strings.Contains(lowResult, "obrigado") {
					if currentLiveURL != "" {
						log.Printf("🎬 [GATILHO IA] Criando clipe para: %s", resultado)
						go videoCutter.CreateClip(liveIDStr, currentLiveURL, float64(milliOffset), "highlight")
					}
				}

				h.Broadcast <- hub.Message{
					Type: "translation", Payload: resultado, LiveID: liveIDStr,
				}
			}(p)
		}
	}
}

func interfaceToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return strconv.FormatFloat(s, 'f', 0, 64)
	case int:
		return strconv.Itoa(s)
	default:
		return ""
	}
}

func ReverseTranslate(w http.ResponseWriter, r *http.Request) {
	if globalGemini == nil {
		http.Error(w, "Gemini não configurado", 500)
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", 400)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coreano, err := globalGemini.TranslateText(ctx, req.Text)
	if err != nil {
		coreano = "Erro na tradução"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"korean": coreano})
}
