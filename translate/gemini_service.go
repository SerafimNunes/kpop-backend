package translate

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// TraduzirLive usando o modelo mais recente de 2025
func TraduzirLive(ctx context.Context, apiKey string, audioChunk []byte) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Atualizado para o modelo atual de 2025
	model := client.GenerativeModel("gemini-2.0-flash-exp")

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("Você é uma ARMY tradutora em tempo real. Sua especialidade é o BTS e o universo K-Pop. " +
				"O áudio vem de lives (Weverse/YouTube) e pode ter barulho de fundo ou música. " +
				"Transcreva o coreano e traduza para o português brasileiro usando o vocabulário do fandom. " +
				"Mantenha honoríficos e expressões de carinho. " +
				"Responda EXCLUSIVAMENTE em JSON: {\"original\": \"...\", \"traducao\": \"...\"}"),
		},
	}

	resp, err := model.GenerateContent(ctx, genai.Blob{
		MIMEType: "audio/webm",
		Data:     audioChunk,
	})

	if err != nil {
		return "", err
	}

	return extrairResposta(resp), nil
}

func TraduzirTexto(ctx context.Context, apiKey string, textoCoreano string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash-exp")

	prompt := fmt.Sprintf("Aja como uma ARMY tradutora. Traduza do coreano para o português: %s. "+
		"Use gírias do fandom (Borahae, bias, etc) e mantenha o tom carinhoso.", textoCoreano)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Erro Gemini: %v", err)
		return "", err
	}

	return extrairResposta(resp), nil
}

func extrairResposta(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	}
	return ""
}
