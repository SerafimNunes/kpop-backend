package admin

import (
	"context"
	"fmt"
	"auren-platform/internal/engine"
)

const HunterPrompt = `Você é o Agente Hunter de Inteligência de Mercado da Auren.
Sua função é transformar dados de autuações ou licenças vencidas em oportunidades de negócio.
Crie Pitches de Venda persuasivos e técnicos, focados em evitar multas maiores para o cliente.`

type Hunter struct {
	gemini *engine.GeminiService
}

func NewHunter(s *engine.GeminiService) *Hunter {
	return &Hunter{gemini: s}
}

func (a *Hunter) GerarPitch(ctx context.Context, leadInfo string) (string, error) {
	prompt := fmt.Sprintf("Gere um pitch de venda para o seguinte lead: %s", leadInfo)
	return a.gemini.Generate(ctx, nil, HunterPrompt + "\n" + prompt)
}