package admin

import (
	"auren-platform/internal/infrastructure/gemini"
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
)

// Controller monitora a saúde financeira e o fluxo dos projetos
type Controller struct {
	gemini *gemini.Service
}

func NewController(s *gemini.Service) *Controller {
	return &Controller{gemini: s}
}

func (a *Controller) AnalisarFluxo(ctx context.Context, dadosFinanceiros string) (string, error) {
	systemPrompt := `### CONTROLLER FINANCEIRO AUREN
Monitore o fluxo de caixa e calcule a saúde financeira dos projetos.
COMPARE: Previsto vs Realizado.
ALERTE: Inadimplências potenciais e margens de lucro abaixo do esperado.`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("DADOS FINANCEIROS:\n%s", dadosFinanceiros)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
