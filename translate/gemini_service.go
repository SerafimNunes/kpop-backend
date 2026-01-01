package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Estrutura para facilitar o tráfego de dados
type RefinedTranslation struct {
	Original string `json:"original"`
	Traducao string `json:"traducao"`
}

type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiService inicia o serviço uma única vez (chamar no main.go)
func NewGeminiService(ctx context.Context, apiKey string) (*GeminiService, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-2.0-flash-exp")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("Você é uma tradutora especialista em K-Pop (ARMY). " +
				"Sua tarefa é receber uma transcrição bruta de uma live e refiná-la. " +
				"1. Corrija erros de transcrição. " +
				"2. Traduza para Português Brasileiro com gírias do fandom. " +
				"3. Mantenha honoríficos (Oppa, Unnie, Hyung). " +
				"Responda SEMPRE no formato JSON: {\"original\": \"...\", \"traducao\": \"...\"}"),
		},
	}

	return &GeminiService{client: client, model: model}, nil
}

// RefinarETraduzir pega o texto do Chirp e prepara para o fã
func (s *GeminiService) RefinarETraduzir(ctx context.Context, textoBruto string) (*RefinedTranslation, error) {
	resp, err := s.model.GenerateContent(ctx, genai.Text(textoBruto))
	if err != nil {
		return nil, err
	}

	rawJSON := s.extrairTexto(resp)

	// Limpa possíveis marcações de markdown ```json ... ```
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	var result RefinedTranslation
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("falha ao parsear JSON da IA: %v", err)
	}

	return &result, nil
}

func (s *GeminiService) extrairTexto(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	}
	return ""
}

func (s *GeminiService) Close() {
	s.client.Close()
}
