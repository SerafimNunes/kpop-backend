package admin

import (
	"context"
	"auren-platform/internal/engine"
)

const CompliancePrompt = `Você é o Analista de Compliance Contratual Sênior da Auren.
Sua missão é blindar a empresa juridicamente. Analise minutas contratuais (ex: Padrão Vale, Petrobras).
Identifique cláusulas leoninas, multas desproporcionais e riscos de responsabilidade civil.
Seja direto: Aponte o risco e sugira a redação segura.`

type Compliance struct {
	gemini *engine.GeminiService
}

func NewCompliance(s *engine.GeminiService) *Compliance {
	return &Compliance{gemini: s}
}

func (a *Compliance) AnalisarRisco(ctx context.Context, contratoTexto string) (string, error) {
	return a.gemini.Generate(ctx, nil, CompliancePrompt+"\nCONTRATO:\n"+contratoTexto)
}