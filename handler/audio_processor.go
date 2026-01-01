package handler

import (
	"context"
	"kpop-backend/hub"
	"kpop-backend/translate"
	"log"
	"sync"
)

type AudioProcessor struct {
	GeminiSvc *translate.GeminiService
	Hub       *hub.Hub
	// Adicionar aqui o cliente do Google Chirp v2 quando configurarmos as credenciais
	mu sync.Mutex
}

func NewAudioProcessor(g *translate.GeminiService, h *hub.Hub) *AudioProcessor {
	return &AudioProcessor{
		GeminiSvc: g,
		Hub:       h,
	}
}

// ProcessAudioChunk recebe o áudio, filtra silêncio/música e orquestra as IAs
func (ap *AudioProcessor) ProcessAudioChunk(ctx context.Context, liveID uint, audioData []byte) {
	// 1. LÓGICA DE PERCEPÇÃO (VAD Local)
	// Aqui implementaríamos um check de amplitude simples ou integração com lib VAD
	if !isSpeech(audioData) {
		return // Ignora música, bateria ou silêncio para poupar tokens/créditos
	}

	// 2. TRANSCRIÇÃO (Google Chirp v2)
	// Por agora, simulamos a saída do Chirp.
	// Em breve faremos o streaming gRPC real para o Google.
	rawText := "Texto bruto vindo do Chirp"

	// 3. REFINAMENTO (Gemini 2.0 Flash)
	refined, err := ap.GeminiSvc.RefinarETraduzir(ctx, rawText)
	if err != nil {
		log.Printf("Erro no refinamento Gemini: %v", err)
		return
	}

	// 4. DISTRIBUIÇÃO VIA HUB
	ap.Hub.Broadcast <- hub.SubtitleMessage{
		LiveID:    liveID,
		Text:      refined.Traducao,
		Timestamp: 0, // Implementar sync de tempo real
		IsFinal:   true,
	}
}

// isSpeech faz a triagem inicial do áudio para evitar processar ruído
func isSpeech(data []byte) bool {
	// Implementação inicial: verificar se o buffer não está vazio ou abaixo de um threshold
	// No futuro, usamos uma lib de FFT para detectar frequências de voz humana
	return len(data) > 500
}
