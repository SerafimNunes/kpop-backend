package handler

import (
	"context"
	"kpop-backend/db"
	"kpop-backend/hub"
	"kpop-backend/models"
	"kpop-backend/translate"
	"log"
	"sync"
	"time"
)

type AudioProcessor struct {
	GeminiSvc *translate.GeminiService
	Hub       *hub.Hub
	mu        sync.Mutex
}

func NewAudioProcessor(g *translate.GeminiService, h *hub.Hub) *AudioProcessor {
	return &AudioProcessor{
		GeminiSvc: g,
		Hub:       h,
	}
}

// ProcessAudioChunk orquestra o fluxo: VAD -> Chirp v2 (STT) -> Gemini (Tradução) -> DB/Web
func (ap *AudioProcessor) ProcessAudioChunk(ctx context.Context, liveID uint, audioData []byte) {
	// 1. LÓGICA DE PERCEPÇÃO (VAD Local)
	if !isSpeech(audioData) {
		return
	}

	// 2. TRANSCRIÇÃO (Placeholder para Google Chirp v2)
	// O Chirp v2 processará o áudio coreano aqui.
	rawText := "Texto capturado pelo Chirp v2"

	// 3. REFINAMENTO CONTEXTUAL (Gemini 2.0 Flash)
	// Usa a lógica que definimos para tradução não-estática.
	refined, err := ap.GeminiSvc.RefinarETraduzir(ctx, rawText)
	if err != nil {
		log.Printf("Erro no refinamento Gemini: %v", err)
		return
	}

	// 4. PERSISTÊNCIA E DISTRIBUIÇÃO
	timestamp := time.Now().UnixMilli()

	// Salva no Banco para o "Netflix de Lives"
	captionLog := models.CaptionLog{
		LiveArchiveID: liveID,
		Timestamp:     timestamp,
		OriginalText:  refined.Original,
		RefinedText:   refined.Traducao,
	}
	db.DB.Create(&captionLog)

	// Envia via WebSocket para o Web App (Mobile Friendly)
	ap.Hub.Broadcast <- hub.SubtitleMessage{
		LiveID:    liveID,
		Text:      refined.Traducao,
		Timestamp: timestamp,
		IsFinal:   true,
	}
}

// StartMockSubtitles - Útil para testar o layout roxo no celular sem áudio real
func (ap *AudioProcessor) StartMockSubtitles(liveID uint) {
	frases := []string{
		"Olá ARMYs! 💜",
		"O Chirp v2 está ouvindo...",
		"Gemini 2.0 traduzindo em tempo real...",
		"Este é o layout mobile-friendly!",
		"Saranghae! (Eu amo vocês)",
	}

	i := 0
	for {
		time.Sleep(4 * time.Second)
		msg := hub.SubtitleMessage{
			LiveID:    liveID,
			Text:      frases[i%len(frases)],
			Timestamp: time.Now().UnixMilli(),
			IsFinal:   true,
		}
		ap.Hub.Broadcast <- msg
		i++
	}
}

func isSpeech(data []byte) bool {
	// Filtro simples de silêncio/tamanho de pacote
	return len(data) > 500
}