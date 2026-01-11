package ai

import (
	"context"
	"fmt"
	"cloud.google.com/go/vertexai/genai"
)

type AuditorAgent struct {
	Svc *GeminiService
}

func NewAuditor(svc *GeminiService) *AuditorAgent {
	return &AuditorAgent{Svc: svc}
}

// AnalyzeFieldEvidence processa fotos/vídeos e confronta com as NBRs
func (a *AuditorAgent) AnalyzeFieldEvidence(ctx context.Context, mediaData []byte, mimeType string, notes string) (string, error) {
	systemPrompt := "VOCÊ É O AUDITOR DE CAMPO DA AUREN. Sua missão é identificar inconformidades visuais. " +
		"Use rigor técnico NBR 12.235 e NBR 7.500. Se houver risco ambiental, use [ALERTA CRÍTICO]."

	parts := []genai.Part{
		genai.Blob{MIMEType: mimeType, Data: mediaData},
		genai.Text(fmt.Sprintf("Notas de campo: %s\n\nTarefa: Analise a conformidade desta evidência.", notes)),
	}

	return a.Svc.Generate(ctx, parts, systemPrompt)
}