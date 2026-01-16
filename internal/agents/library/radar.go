package library

import (
	"auren-platform/internal/engine"
	"context"
)

const RadarPrompt = `Você é o Agente Radar Legislativo.
Analise textos de Diários Oficiais e novas leis.
Filtre apenas o que impacta consultoria ambiental e resuma a mudança para a equipe.`

type Radar struct {
	gemini *engine.GeminiService
}

func NewRadar(s *engine.GeminiService) *Radar {
	return &Radar{gemini: s}
}

func (a *Radar) AnalisarDiario(ctx context.Context, textoDiario string) (string, error) {
	return a.gemini.Generate(ctx, nil, RadarPrompt+"\nTEXTO:\n"+textoDiario)
}