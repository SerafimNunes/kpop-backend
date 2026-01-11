package ai

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/vertexai/genai"
)

// Generate é o motor universal da plataforma.
// Ele aceita qualquer combinação de Part (Texto, Blob/Mídia).
// Adicionado log de telemetria para diagnóstico de tempo de resposta.
func (s *GeminiService) Generate(ctx context.Context, parts []genai.Part, systemInstruction string) (string, error) {
	var modelExecutor *genai.GenerativeModel

	start := time.Now()
	log.Println("🤖 [AUREN-AI] Iniciando execução de modelo multimodal...")

	if systemInstruction != "" {
		// Cria modelo temporário para instruções específicas de agentes
		modelExecutor = s.client.GenerativeModel("gemini-2.0-flash-001")
		modelExecutor.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(systemInstruction)},
		}
		modelExecutor.SetTemperature(0.0)
		log.Printf("🛠️ [AUREN-AI] Modelo configurado com System Instruction (Tamanho: %d)", len(systemInstruction))
	} else {
		modelExecutor = s.model
	}

	resp, err := modelExecutor.GenerateContent(ctx, parts...)
	if err != nil {
		log.Printf("[AUREN-INFRA] Falha na execução de IA: %v", err)
		return "", err
	}

	result := s.extractText(resp)
	log.Printf("✅ [AUREN-AI] Processamento concluído. Latência: %v | Resposta: %d bytes", time.Since(start), len(result))

	return result, nil
}

// ExecuteSimplePrompt para interações rápidas de texto puro.
func (s *GeminiService) ExecuteSimplePrompt(ctx context.Context, prompt string) (string, error) {
	start := time.Now()
	log.Printf("🤖 [AUREN-AI] Executando prompt simples: %s", prompt)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("❌ [AUREN-AI] Erro em prompt simples: %v", err)
		return "", err
	}

	result := s.extractText(resp)
	log.Printf("✅ [AUREN-AI] Resposta rápida recebida. Tempo: %v", time.Since(start))

	return result, nil
}

// AnalyzeVisualEvidence foca na análise técnica de mídia isolada
func (s *GeminiService) AnalyzeVisualEvidence(ctx context.Context, mediaData []byte, promptText string) (string, error) {
	mimeType := "image/jpeg"
	// Detecta se é vídeo pelo tamanho ou cabeçalho (simplificado para o seu fluxo original)
	if len(mediaData) > 5000000 {
		mimeType = "video/mp4"
	}

	log.Printf("📸 [AUREN-AI] Analisando evidência visual. Tipo: %s | Tamanho: %d bytes", mimeType, len(mediaData))

	prompt := []genai.Part{
		genai.Blob{MIMEType: mimeType, Data: mediaData},
		genai.Text(promptText),
	}

	return s.Generate(ctx, prompt, "")
}
