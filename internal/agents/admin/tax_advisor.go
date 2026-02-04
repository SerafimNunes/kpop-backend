package admin

import (
	"auren-platform/internal/infrastructure/gemini"
	"context"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
)

// TaxAdvisor garante a conformidade tributária dos serviços ambientais
type TaxAdvisor struct {
	gemini *gemini.Service
}

func NewTaxAdvisor(s *gemini.Service) *TaxAdvisor {
	return &TaxAdvisor{gemini: s}
}

func (a *TaxAdvisor) ValidarImpostos(ctx context.Context, dados string) (string, error) {
	systemPrompt := `### ESPECIALISTA CONTÁBIL AUREN PLATFORM
Sua missão é garantir que a consultoria ambiental siga a legislação tributária brasileira.
Analise os valores e tipos de serviço e informe as retenções de impostos (ISS, PIS, COFINS, CSLL, IRRF).
Seja preciso e técnico. Utilize normas da Receita Federal e legislações municipais para ISS.`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("CENÁRIO PARA ANÁLISE:\n%s", dados)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
