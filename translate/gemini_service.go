package translate

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiService(ctx context.Context) (*GeminiService, error) {
	// A Vertex AI exige o ID do projeto e a localização (ex: us-central1)
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION") // Ex: us-central1

	// O caminho para o arquivo JSON de credenciais que você vai gerar
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// Inicializa o cliente usando as credenciais do Google Cloud
	client, err := genai.NewClient(ctx, projectID, location, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente Vertex AI: %v", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash-001")

	// Instrução de Sistema Robusta (Chirp v2 style)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(
			"Você é um tradutor simultâneo especializado em lives de K-pop (K-LENS STUDIO). " +
				"Sua tarefa é converter áudio coreano para português brasileiro natural. " +
				"REGRAS CRÍTICAS: " +
				"1. Se houver música predominante, responda apenas: [MÚSICA]. " +
				"2. Se houver silêncio ou apenas ruído de fundo, responda: [SILÊNCIO]. " +
				"3. Se houver fala, seja informal, use gírias do fandom (bias, comeback, etc). " +
				"4. Seja extremamente conciso para caber em legendas rápidas.",
		)},
	}

	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

func (s *GeminiService) TranslateAudio(ctx context.Context, audioData []byte) (string, error) {
	// Na Vertex AI, enviamos o blob de áudio como parte do conteúdo
	prompt := []genai.Part{
		genai.Blob{
			MIMEType: "audio/webm",
			Data:     audioData,
		},
		genai.Text("Traduza o áudio acima."),
	}

	resp, err := s.model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", fmt.Errorf("erro vertex ai audio: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", nil
	}

	var output string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			output += string(text)
		}
	}

	return output, nil
}

func (s *GeminiService) TranslateText(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf("Traduza para coreano casual/fofo de Weverse (apenas o texto): %s", text)
	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if t, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(t), nil
		}
	}
	return "", fmt.Errorf("falha na tradução de texto via Vertex")
}

func (s *GeminiService) Close() {
	s.client.Close()
}
