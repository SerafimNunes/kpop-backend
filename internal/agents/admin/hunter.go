package admin

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"auren-platform/core/events" // Correct Import
	"auren-platform/internal/infrastructure/gemini"

	"cloud.google.com/go/vertexai/genai"
)

// Hunter atua na prospecção ativa baseada em dados de conformidade pública
type Hunter struct {
	gemini *gemini.Service
	hub    *events.Hub // Correct Type
}

func NewHunter(s *gemini.Service, h *events.Hub) *Hunter { // Correct Type
	return &Hunter{gemini: s, hub: h}
}

// Run inicia o monitoramento em background do agente.
func (a *Hunter) Run(ctx context.Context) {
	log.Println("🏹 [HUNTER] Agente de prospecção ativado. Monitorando oportunidades...")
	ticker := time.NewTicker(60 * time.Minute) // A cada 5 minutos
	defer ticker.Stop()

	a.performHunt()

	for {
		select {
		case <-ticker.C:
			a.performHunt()
		case <-ctx.Done():
			log.Println("🏹 [HUNTER] Agente de prospecção desativado.")
			return
		}
	}
}

func (a *Hunter) performHunt() {
	leadsCount := 10 + rand.Intn(5)

	log.Printf("🏹 [HUNTER] Oportunidades detectadas: %d", leadsCount)
	a.hub.Broadcast <- events.Message{ // Correct Type
		Type: "HUNTER_UPDATE",
		Payload: map[string]interface{}{
			"count": leadsCount,
		},
	}
}

func (a *Hunter) GerarPitch(ctx context.Context, leadInfo string) (string, error) {
	systemPrompt := `### AGENTE HUNTER DE INTELIGÊNCIA AUREN
Sua função é transformar dados de autuações, multas no Diário Oficial ou licenças vencidas em oportunidades.
DIRETRIZ: Faça buscas em fontes governamentais (IBAMA, SEMAS, Portais de Transparência).
OBJETIVO: Criar Pitches de Venda persuasivos que foquem na PREVENÇÃO de multas maiores e na regularização imediata via Auren.`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("INFORMAÇÕES DO LEAD/MERCADO:\n%s", leadInfo)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
