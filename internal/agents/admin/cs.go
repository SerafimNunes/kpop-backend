package admin

import (
	"auren-platform/internal/infrastructure/gemini"
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
)

// CustomerSuccess foca na retenção e expansão da conta do cliente
type CustomerSuccess struct {
	gemini *gemini.Service
}

func NewCustomerSuccess(s *gemini.Service) *CustomerSuccess {
	return &CustomerSuccess{gemini: s}
}

func (a *CustomerSuccess) SugerirUpsell(ctx context.Context, contextoProjeto string) (string, error) {
	systemPrompt := `### AGENTE DE CUSTOMER SUCCESS AUREN
Analise o feedback ou o projeto entregue e identifique necessidades latentes para Upsell.
ESTRATÉGIA: Se o cliente resolveu um passivo de resíduos, sugira monitoramento contínuo ou treinamentos de conformidade.`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("CONTEXTO DO PROJETO ATUAL:\n%s", contextoProjeto)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
