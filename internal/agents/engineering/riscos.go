package engineering

import (
	"context"
	"auren-platform/internal/engine"
)

const RiscosPrompt = `Você é o Analista de Riscos Ambientais e Emergências.
Simule cenários de acidentes (vazamento, incêndio, explosão) baseados nas instalações.
Sugira medidas mitigadoras para o PAE (Plano de Ação de Emergência).`

type AnalistaRiscos struct {
	gemini *engine.GeminiService
}

func NewAnalistaRiscos(s *engine.GeminiService) *AnalistaRiscos {
	return &AnalistaRiscos{gemini: s}
}

func (a *AnalistaRiscos) SimularCenario(ctx context.Context, dadosInstalacao string) (string, error) {
	return a.gemini.Generate(ctx, nil, RiscosPrompt+"\nINSTALAÇÃO:\n"+dadosInstalacao)
}