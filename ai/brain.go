package ai

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
)

// GeminiService é a estrutura centralizada para todo o pacote ai.
type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiService instancia o provedor de IA e configura as diretrizes da Auren.
func NewGeminiService(ctx context.Context, projectID, location string) (*GeminiService, error) {
	log.Printf("[AUREN-AI] Iniciando NewGeminiService para Projeto: %s, Local: %s", projectID, location)

	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if projectID == "" || location == "" {
		return nil, fmt.Errorf("configuração GCP incompleta")
	}

	client, err := genai.NewClient(ctx, projectID, location, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Printf("[AUREN-AI] ERRO CRÍTICO: Falha ao conectar na Vertex AI: %v", err)
		return nil, fmt.Errorf("falha Vertex AI: %v", err)
	}

	// Flash 2.0 é essencial para processamento de vídeo e latência baixa
	model := client.GenerativeModel("gemini-2.0-flash-001")
	model.SetTemperature(0.0)
	model.SetMaxOutputTokens(4000)

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(
			"VOCÊ É O CORE DE INTELIGÊNCIA DA AUREN PLATFORM. " +
				"SUA MISSÃO: Consolidar dados de campo (fotos/vídeos/áudio) com formulários técnicos e legislação ambiental brasileira. " +
				"DIRETRIZES DE PENSAMENTO: " +
				"1. CROSS-CHECK: Compare rigorosamente o formulário com as evidências visuais. Aponte discrepâncias. " +
				"2. RIGOR NBR: Para resíduos Classe I, cite obrigatoriamente NBR 12.235 e NBR 7.500. " +
				"3. VISÃO TÉCNICA: Identifique corrosão, falta de rotulagem GHS ou ausência de bacias de contenção. " +
				"4. SEGURANÇA: Se detectar risco de crime ambiental, use o prefixo [ALERTA CRÍTICO].",
		)},
	}

	log.Println("[AUREN-AI] GeminiService instanciado com sucesso.")
	return &GeminiService{client: client, model: model}, nil
}

// ConsolidatePGRS cruza os dados do formulário com as evidências coletadas
func (s *GeminiService) ConsolidatePGRS(ctx context.Context, pgrsDataJSON string, fieldNotes string, mediaData []byte, mimeType string) (string, error) {
	start := time.Now()
	log.Printf("[AUREN-AI] Iniciando Consolidação PGRS. Payload Size: %d bytes", len(mediaData))

	prompt := []genai.Part{
		genai.Text("### DADOS DO FORMULÁRIO (JSON): " + pgrsDataJSON),
		genai.Text("### NOTAS DE VISTORIA: " + fieldNotes),
		genai.Text("### TAREFA: Audite a fidedignidade. Gere o 'Diagnóstico de Situação Atual' para o laudo PGRS."),
	}

	if len(mediaData) > 0 {
		log.Printf("[AUREN-AI] Anexando mídia do tipo: %s", mimeType)
		prompt = append(prompt, genai.Blob{MIMEType: mimeType, Data: mediaData})
	}

	resp, err := s.model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Printf("[AUREN-AI] ERRO na geração de conteúdo (Consolidate): %v", err)
		return "", err
	}

	result := s.extractText(resp)
	log.Printf("[AUREN-AI] Consolidação finalizada em %v", time.Since(start))
	return result, nil
}

// extractText realiza o parse limpo da resposta da Vertex (Única implementação no pacote)
func (s *GeminiService) extractText(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		log.Println("[AUREN-AI] AVISO: IA retornou resposta vazia.")
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

// Close encerra os recursos do cliente.
func (s *GeminiService) Close() {
	if s.client != nil {
		log.Println("[AUREN-AI] Encerrando conexão GeminiService.")
		s.client.Close()
	}
}