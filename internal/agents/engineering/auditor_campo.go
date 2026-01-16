package engineering

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

// AnalyzePGRSCompliance verifica os dados de um formulário PGRS contra as normas

func (a *AuditorAgent) AnalyzePGRSCompliance(ctx context.Context, pgrsDataJSON []byte) (string, error) {

	systemPrompt := "VOCÊ É O AUDITOR-CHEFE DA AUREN, especialista em legislação de resíduos sólidos (Lei 12.305/10 e NBR 10.004). " +

		"Sua tarefa é analisar os dados de um PGRS em formato JSON e apontar, em formato de lista (bullet points), TODAS as inconsistências, " +

		"pontos de melhoria ou riscos de não conformidade. Seja extremamente rigoroso e técnico. Se não houver nenhuma inconsistência, responda com 'Nenhuma inconsistência encontrada. O PGRS atende aos requisitos primários de conformidade.'"

	prompt := fmt.Sprintf("Analise os seguintes dados de um PGRS e aponte as não conformidades:\n\n```json\n%s\n```", string(pgrsDataJSON))

	parts := []genai.Part{

		genai.Text(prompt),
	}

	return a.Svc.Generate(ctx, parts, systemPrompt)

}

// AnalyzeInspectionNotes analisa as observações de texto livre de um técnico de campo.

func (a *AuditorAgent) AnalyzeInspectionNotes(ctx context.Context, notes string) (string, error) {

	systemPrompt := "VOCÊ É O AUDITOR DE CAMPO DA AUREN. Sua missão é interpretar as anotações de um técnico em campo e transformá-las em um resumo coeso, " +

		"identificando pontos de atenção e potenciais riscos ambientais. Formate a saída em um parágrafo curto e objetivo."

	parts := []genai.Part{

		genai.Text(fmt.Sprintf("Analisar as seguintes anotações de vistoria:\n\n\"%s\"", notes)),
	}

	return a.Svc.Generate(ctx, parts, systemPrompt)

}
