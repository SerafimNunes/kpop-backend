package translate

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiService(ctx context.Context, apiKey string) (*GeminiService, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente gemini: %v", err)
	}

	// Usando o modelo 2.0 Flash para velocidade e custo-benefício
	model := client.GenerativeModel("gemini-2.0-flash")

	// Configuração do sistema otimizada para K-Pop
	model.SystemInstruction = genai.NewUserContent(genai.Text(
		"Você é um tradutor simultâneo de K-pop especializado em lives. " +
			"Traduza o áudio coreano para português brasileiro de forma natural e informal. " +
			"Se houver música ou silêncio, ignore. Seja conciso e use termos do fandom quando apropriado."))

	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

// TranslateAudio: Tradução da Live (Coreano -> PT-BR)
func (s *GeminiService) TranslateAudio(ctx context.Context, audioData []byte) (string, error) {
	data := genai.Blob{
		MIMEType: "audio/webm", // Compatível com o que o navegador envia
		Data:     audioData,
	}

	resp, err := s.model.GenerateContent(ctx, data)
	if err != nil {
		return "", fmt.Errorf("erro na geração de conteúdo: %v", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", nil
	}

	var translatedText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			translatedText += string(text)
		}
	}

	return translatedText, nil
}

// TranslateText: Chat Reverso (PT-BR -> Coreano Casual)
func (s *GeminiService) TranslateText(ctx context.Context, text string) (string, error) {
	// Prompt específico para soar natural para o Idol
	prompt := fmt.Sprintf("Traduza para coreano casual e fofo (estilo Weverse) para um Idol de K-pop. Retorne apenas a tradução: %s", text)
	
	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		if t, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(t), nil
		}
	}
	return "", fmt.Errorf("não foi possível traduzir o texto")
}

func (s *GeminiService) Close() {
	s.client.Close()
}