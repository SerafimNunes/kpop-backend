package admin

import (
	"context"
	"auren-platform/internal/engine"
)

const CSPrompt = `Você é o Agente de Customer Success (Pós-Venda) da Auren.
Analise o feedback do cliente ou o projeto entregue e sugira serviços complementares (Upsell).
Exemplo: Se entregamos PGRS, sugira Treinamento de Brigada.`

type CustomerSuccess struct {
	gemini *engine.GeminiService
}

func NewCustomerSuccess(s *engine.GeminiService) *CustomerSuccess {
	return &CustomerSuccess{gemini: s}
}

func (a *CustomerSuccess) SugerirUpsell(ctx context.Context, contextoProjeto string) (string, error) {
	return a.gemini.Generate(ctx, nil, CSPrompt+"\nCONTEXTO:\n"+contextoProjeto)
}