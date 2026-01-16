package engineering

import (
	"context"
	"auren-platform/internal/engine"
)

const CondicionantesPrompt = `Você é o Agente de Monitoramento de Condicionantes.
Leia a Licença Ambiental (LO, LI, LP) e extraia prazos e obrigações.
Gere um cronograma de alertas para evitar multas por perda de prazo.`

type MonitorCondicionantes struct {
	gemini *engine.GeminiService
}

func NewMonitorCondicionantes(s *engine.GeminiService) *MonitorCondicionantes {
	return &MonitorCondicionantes{gemini: s}
}

func (a *MonitorCondicionantes) ExtrairObrigacoes(ctx context.Context, licencaTexto string) (string, error) {
	return a.gemini.Generate(ctx, nil, CondicionantesPrompt+"\nLICENÇA:\n"+licencaTexto)
}