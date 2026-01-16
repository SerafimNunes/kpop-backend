package engineering

import (
	"context"
	"auren-platform/internal/engine"
)

const PNRSPrompt = `Você é o Especialista em Classificação de Resíduos (PNRS/CONAMA).
Classifique resíduos conforme NBR 10004 (Classe I, II A, II B).
Identifique incompatibilidade química no armazenamento e sugira a destinação final correta.`

type EspecialistaPNRS struct {
	gemini *engine.GeminiService
}

func NewEspecialistaPNRS(s *engine.GeminiService) *EspecialistaPNRS {
	return &EspecialistaPNRS{gemini: s}
}

func (a *EspecialistaPNRS) Classificar(ctx context.Context, descritivo string) (string, error) {
	return a.gemini.Generate(ctx, nil, PNRSPrompt+"\nRESÍDUO:\n"+descritivo)
}