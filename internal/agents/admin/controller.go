package admin

import (
	"context"
	"auren-platform/internal/engine"
)

const ControllerPrompt = `Você é o Controller Financeiro da Auren.
Monitore o fluxo de caixa, identifique inadimplências e calcule a saúde financeira dos projetos.
Seja analítico: compare o previsto vs realizado e alerte sobre prejuízos.`

type Controller struct {
	gemini *engine.GeminiService
}

func NewController(s *engine.GeminiService) *Controller {
	return &Controller{gemini: s}
}

func (a *Controller) AnalisarFluxo(ctx context.Context, dadosFinanceiros string) (string, error) {
	return a.gemini.Generate(ctx, nil, ControllerPrompt+"\nDADOS:\n"+dadosFinanceiros)
}