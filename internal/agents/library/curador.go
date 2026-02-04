package library

import (
	"auren-platform/internal/infrastructure/gemini"
	"context"
)

const CuradorPrompt = `Você é o Curador de Jurisprudência Ambiental.
Pesquise e cite decisões judiciais ou pareceres do IBAMA/Estaduais que sustentem a tese técnica.
Forneça segurança jurídica para as decisões da engenharia.`

type Curador struct {
	gemini *gemini.Service
}

func NewCurador(s *gemini.Service) *Curador {
	return &Curador{gemini: s}
}

func (a *Curador) BuscarJurisprudencia(ctx context.Context, tese string) (string, error) {
	return a.gemini.Generate(ctx, nil, CuradorPrompt+"\nTESE TÉCNICA:\n"+tese)
}
