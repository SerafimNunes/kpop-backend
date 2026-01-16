package engineering

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/vertexai/genai"
)

type RedatorAgent struct {
	Svc *GeminiService
}

func NewRedator(svc *GeminiService) *RedatorAgent {
	return &RedatorAgent{Svc: svc}
}

// RefinarParaLaudo transforma notas brutas de campo em texto formal para o laudo PGRS.
func (a *RedatorAgent) RefinarParaLaudo(ctx context.Context, textoBruto string) (string, error) {
	systemPrompt := "Você é o Redator Técnico da Auren. Transforme notas de campo em texto profissional, impessoal e em conformidade com as normas ABNT."
	
	parts := []genai.Part{
		genai.Text("Texto para refinar: " + textoBruto),
	}

	return a.Svc.Generate(ctx, parts, systemPrompt)
}

// AnalyzeTechnicalAudio processa a transcrição e análise técnica de áudios
func (a *RedatorAgent) AnalyzeTechnicalAudio(ctx context.Context, audioData []byte, promptText string) (string, error) {
	if len(audioData) == 0 {
		return "", nil
	}
	log.Println("[AUREN-AI] Processando Áudio Técnico de Campo...")

	prompt := []genai.Part{
		genai.Blob{MIMEType: "audio/wav", Data: audioData},
		genai.Text("Extraia dados técnicos deste áudio: " + promptText),
	}

	return a.Svc.Generate(ctx, prompt, "Atue como um perito ambiental transcrevendo evidências de áudio.")
}

// AnalyzeText realiza análise simples de strings curtas.
func (a *RedatorAgent) AnalyzeText(ctx context.Context, text string) (string, error) {
		if len(text) < 20 {
			return "Input insuficiente para análise técnica.", nil
		}
		return a.Svc.ExecuteSimplePrompt(ctx, text)
	}
	
	// GeneratePGRSReport cria o laudo final do PGRS em formato Markdown.
	func (a *RedatorAgent) GeneratePGRSReport(ctx context.Context, pgrsDataJSON []byte, auditResult string) (string, error) {
		systemPrompt := "VOCÊ É O REDATOR-CHEFE DA AUREN, especialista em criar laudos técnicos de engenharia ambiental (PGRS). " +
			"Sua tarefa é gerar um MEMORIAL DESCRITIVO completo e bem formatado em **Markdown**, usando os dados do PGRS e os apontamentos da auditoria. " +
			"Siga estritamente a estrutura: 1. OBJETO, 2. QUADRO RESUMO DE RESÍDUOS (em uma tabela Markdown), 3. APONTAMENTOS DA AUDITORIA, 4. RESPONSABILIDADE TÉCNICA. " +
			"Seja formal, técnico e use os dados fornecidos para preencher todas as seções."
	
		prompt := fmt.Sprintf("Dados do PGRS em JSON:\n```json\n%s\n```\n\nApontamentos da Auditoria:\n```\n%s\n```\n\nPor favor, gere o Memorial Descritivo completo em Markdown.", string(pgrsDataJSON), auditResult)
	
		parts := []genai.Part{
			genai.Text(prompt),
		}
	
		return a.Svc.Generate(ctx, parts, systemPrompt)
	}
	