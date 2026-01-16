package admin

import (
	"context"
	"fmt"
	"auren-platform/internal/engine"
)

const TaxAdvisorPrompt = `Você é o Especialista Contábil da Auren Platform. 
Sua missão é garantir que a consultoria ambiental siga a legislação tributária brasileira.
Analise os valores e tipos de serviço e informe as retenções de impostos (ISS, PIS, COFINS, CSLL, IRRF).
Seja preciso e técnico. Utilize normas da Receita Federal.`

type TaxAdvisor struct {
	gemini *engine.GeminiService
}

func NewTaxAdvisor(s *engine.GeminiService) *TaxAdvisor {
	return &TaxAdvisor{gemini: s}
}

func (a *TaxAdvisor) ValidarImpostos(ctx context.Context, dados string) (string, error) {
	prompt := fmt.Sprintf("Analise tributariamente este cenário de serviço ambiental: %s", dados)
	return a.gemini.Generate(ctx, nil, TaxAdvisorPrompt + "\n" + prompt)
}