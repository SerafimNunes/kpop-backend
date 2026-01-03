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
	CheckOrigin:     func(r *http.Request) bool { return true },
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

	// Variável para armazenar a URL da live enviada pelo Studio
	var currentLiveURL string

	defer func() {
		h.Unregister <- clientChan
		conn.Close()
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

		// 1. TRATAMENTO DE MENSAGENS TEXTO (Comandos JSON)
		if messageType == websocket.TextMessage {
			var raw map[string]interface{}
			if err := json.Unmarshal(p, &raw); err != nil {
				continue
			}

			// A) ATUALIZAR CONFIGURAÇÃO (Frequente ao entrar na live)
			if raw["action"] == "update_config" {
				log.Printf("⚙️ Configuração atualizada: %v | %v | URL detectada", raw["ratio"], raw["duration"])

				duration, _ := strconv.Atoi(interfaceToString(raw["duration"]))
				ratio := interfaceToString(raw["ratio"])
				url := interfaceToString(raw["live_url"])

				videoCutter.UpdateConfig(duration, ratio)
				currentLiveURL = url
				continue
			}

			// B) NOVO: CORTE MANUAL (Botões Premium do Celular)
			if raw["type"] == "MANUAL_CLIP" {
				ratio := interfaceToString(raw["ratio"])
				url := interfaceToString(raw["url"])
				if url == "" {
					url = currentLiveURL
				}

				log.Printf("🕹️ [MANUAL] Solicitado corte em %s", ratio)

				// Atualiza o ratio apenas para este corte se necessário
				videoCutter.UpdateConfig(61, ratio)

				milliOffset := time.Since(startTime).Milliseconds()
				go videoCutter.CreateClip(liveIDStr, url, float64(milliOffset), "manual_premium")

				// Feedback visual para o celular
				h.Broadcast <- hub.Message{
					Type: "translation", Payload: "🎬 CORTE MANUAL INICIADO (" + ratio + ")", LiveID: liveIDStr,
				}
				continue
			}
		}

		// 2. TRATAMENTO DE ÁUDIO BINÁRIO (IA)
		if messageType == websocket.BinaryMessage && gemini != nil {
			// Se o buffer for muito pequeno ou silêncio detectado pelo processador, ignoramos
			if len(p) < 100 || !processor.ShouldProcess(p) {
				continue
			}

			go func(audioData []byte) {
				select {
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				default:
					return // Saturado
				}

				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				resultado, err := gemini.TranslateAudio(ctx, audioData)
				if err != nil || resultado == "" {
					return
				}

				milliOffset := time.Since(startTime).Milliseconds()

				// Persistência no Banco de Dados
				if db.DB != nil {
					db.DB.Create(&models.CaptionLog{
						LiveArchiveID: uint(liveID),
						Timestamp:     milliOffset,
						Text:          resultado,
					})
				}

				// Lógica de Clipe Automático com GATILHOS
				lowResult := strings.ToLower(resultado)
				if strings.Contains(lowResult, "💜") || strings.Contains(lowResult, "tchau") || strings.Contains(lowResult, "obrigado") {
					if currentLiveURL != "" {
						log.Printf("🎬 [GATILHO IA] Criando clipe para: %s", resultado)
						go videoCutter.CreateClip(liveIDStr, currentLiveURL, float64(milliOffset), "highlight")
					} else {
						log.Printf("⚠️ Gatilho ativado, mas URL da live não foi definida")
					}
				}

				// Envia tradução para a interface
				h.Broadcast <- hub.Message{
					Type: "translation", Payload: resultado, LiveID: liveIDStr,
				}
			}(p)
		}
	}
}

// Auxiliar para converter interface para string com segurança
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

// ReverseTranslate trata a tradução de PT-BR para Coreano (Botão do Studio)
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
