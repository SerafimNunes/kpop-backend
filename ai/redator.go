package ai

import (
	"context"
	"log"

	"cloud.google.com/go/vertexai/genai"
)

// RefinarParaLaudo transforma notas brutas de campo em texto formal para o laudo PGRS.
func (s *GeminiService) RefinarParaLaudo(ctx context.Context, textoBruto string) (string, error) {
	systemPrompt := "Você é o Redator Técnico da Auren. Transforme notas de campo em texto profissional, impessoal e em conformidade com as normas ABNT."
	
	parts := []genai.Part{
		genai.Text("Texto para refinar: " + textoBruto),
	}

	return s.Generate(ctx, parts, systemPrompt)
}

// AnalyzeTechnicalAudio processa a transcrição e análise técnica de áudios
func (s *GeminiService) AnalyzeTechnicalAudio(ctx context.Context, audioData []byte, promptText string) (string, error) {
	if len(audioData) == 0 {
		return "", nil
	}
	log.Println("[AUREN-AI] Processando Áudio Técnico de Campo...")

	prompt := []genai.Part{
		genai.Blob{MIMEType: "audio/wav", Data: audioData},
		genai.Text("Extraia dados técnicos deste áudio: " + promptText),
	}

	return s.Generate(ctx, prompt, "Atue como um perito ambiental transcrevendo evidências de áudio.")
}

// AnalyzeText realiza análise simples de strings curtas.
func (s *GeminiService) AnalyzeText(ctx context.Context, text string) (string, error) {
	if len(text) < 20 {
		return "Input insuficiente para análise técnica.", nil
	}
	return s.ExecuteSimplePrompt(ctx, text)
}