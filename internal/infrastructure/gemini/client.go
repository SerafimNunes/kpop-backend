package gemini

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
)

// Service gerencia a conexão bruta com o Google Vertex AI
type Service struct {
	Client *genai.Client
	Model  *genai.GenerativeModel
}

// NewService inicializa o cliente Gemini de forma independente
func NewService(ctx context.Context, projectID, location string) (*Service, error) {
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	client, err := genai.NewClient(ctx, projectID, location, option.WithCredentialsFile(creds))
	if err != nil {
		return nil, fmt.Errorf("falha Vertex AI: %v", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash-001")
	model.SetTemperature(0.0)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("VOCÊ É O CORE DA AUREN PLATFORM. Sua missão é auditoria e conformidade ambiental.")},
	}

	return &Service{Client: client, Model: model}, nil
}

// Generate é o método universal de inferência
func (s *Service) Generate(ctx context.Context, parts []genai.Part, systemInstruction string) (string, error) {
	var executor *genai.GenerativeModel
	if systemInstruction != "" {
		executor = s.Client.GenerativeModel("gemini-2.0-flash-001")
		executor.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemInstruction)}}
		executor.SetTemperature(0.0)
	} else {
		executor = s.Model
	}

	resp, err := executor.GenerateContent(ctx, parts...)
	if err != nil {
		return "", err
	}
	return s.extractText(resp), nil
}

// ExecuteSimplePrompt para diagnósticos rápidos
func (s *Service) ExecuteSimplePrompt(ctx context.Context, prompt string) (string, error) {
	return s.Generate(ctx, []genai.Part{genai.Text(prompt)}, "")
}

func (s *Service) extractText(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return ""
	}
	var b strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			b.WriteString(string(t))
		}
	}
	return strings.TrimSpace(b.String())
}

// Close encerra a conexão
func (s *Service) Close() {
	if s.Client != nil {
		s.Client.Close()
	}
}
