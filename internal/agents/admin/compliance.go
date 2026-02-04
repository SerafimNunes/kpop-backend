package admin

import (
	"auren-platform/internal/infrastructure/gemini"
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
)

// Compliance blinda a empresa contra riscos jurídicos em contratos complexos
type Compliance struct {
	gemini *gemini.Service
}

func NewCompliance(s *gemini.Service) *Compliance {
	return &Compliance{gemini: s}
}

// AnalisarContrato sincronizado com Handler e Socket
func (a *Compliance) AnalisarContrato(ctx context.Context, contratoTexto string) (string, error) {
	systemPrompt := `### ANALISTA DE COMPLIANCE CONTRATUAL SÊNIOR
Sua missão é blindar a Auren juridicamente. Analise minutas contratuais.
FOCO: Identificar cláusulas leoninas, multas desproporcionais e riscos de responsabilidade civil.
FORMATO: Aponte o risco e sugira imediatamente uma redação alternativa segura.`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("MINUTA PARA ANÁLISE:\n%s", contratoTexto)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
