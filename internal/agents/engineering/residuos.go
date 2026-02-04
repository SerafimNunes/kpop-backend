package engineering

import (
	"auren-platform/internal/infrastructure/gemini"
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
)

// EspecialistaPNRS atua na classificação e gestão de resíduos sólidos
type EspecialistaPNRS struct {
	gemini *gemini.Service
}

// NewEspecialistaPNRS cria uma nova instância do especialista em PNRS
func NewEspecialistaPNRS(s *gemini.Service) *EspecialistaPNRS {
	return &EspecialistaPNRS{gemini: s}
}

// Classificar analisa o descritivo do resíduo com base na NBR 10.004
func (a *EspecialistaPNRS) Classificar(ctx context.Context, descritivo string) (string, error) {
	if descritivo == "" {
		return "Descrição do resíduo não informada.", nil
	}

	systemPrompt := `### ESPECIALISTA EM CLASSIFICAÇÃO DE RESÍDUOS AUREN
Você é um Engenheiro Ambiental sênior especialista em PNRS e CONAMA.
Sua tarefa é:
1. Classificar resíduos conforme NBR 10004 (Classe I - Perigosos, II A - Não Inertes, II B - Inertes).
2. Identificar riscos de incompatibilidade química no armazenamento.
3. Sugerir a destinação final ambientalmente correta (CADRI, incineração, aterro, etc).`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("DESCRITIVO DO RESÍDUO:\n%s", descritivo)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
