package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// RefinedTranslation é a estrutura que garante que o Gemini nos entregue
// exatamente o que precisamos para a legenda e para o banco de dados.
type RefinedTranslation struct {
	Original string `json:"original"`
	Traducao string `json:"traducao"`
}

type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiService inicia o serviço configurado para o modo "Estúdio de Produção".
func NewGeminiService(ctx context.Context, apiKey string) (*GeminiService, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no Gemini: %v", err)
	}

	// Usando o modelo Flash 2.0 para a menor latência possível durante a live
	model := client.GenerativeModel("gemini-2.0-flash-exp")

	// Configuração do comportamento da IA para legendagem de vídeos (YouTube/TikTok)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("Você é uma tradutora de elite para lives de K-Pop e trabalha para o K-Live Lens. " +
				"Sua tarefa é receber a transcrição bruta (Coreano) e gerar legendas para fãs brasileiros. " +
				"REGRAS DE OURO: " +
				"1. CONCISÃO: Máximo de 12 palavras por frase. Se for longo, resuma o sentido. " +
				"2. CONTEXTO: Use gírias do fandom (ex: 'bias', 'comeback', 'stan'). " +
				"3. HONORÍFICOS: Mantenha termos como Oppa, Hyung, Unnie, Noona quando apropriado. " +
				"4. ADAPTAÇÃO: Converta expressões culturais coreanas para o português natural do Brasil. " +
				"5. MÚSICA: Se detectar que o Idol está cantando, coloque a frase entre notas musicais ♫. " +
				"FORMATO: Responda APENAS em JSON estrito, sem markdown: " +
				"{\"original\": \"texto em coreano\", \"traducao\": \"legenda em português\"}"),
		},
	}

	// Parâmetros de geração para evitar que a IA divague
	model.SetTemperature(0.3) // Menos criatividade, mais precisão técnica
	model.SetMaxOutputTokens(150)

	return &GeminiService{client: client, model: model}, nil
}

// RefinarETraduzir processa o texto vindo do Chirp v2 e entrega a legenda final
func (s *GeminiService) RefinarETraduzir(ctx context.Context, textoBruto string) (*RefinedTranslation, error) {
	if textoBruto == "" {
		return nil, fmt.Errorf("texto bruto vazio")
	}

	resp, err := s.model.GenerateContent(ctx, genai.Text(textoBruto))
	if err != nil {
		return nil, fmt.Errorf("erro na chamada do Gemini: %v", err)
	}

	// Extração segura do conteúdo
	rawJSON := s.extrairTexto(resp)
	if rawJSON == "" {
		return nil, fmt.Errorf("IA retornou resposta vazia")
	}

	// Limpeza de possíveis blocos de código markdown que a IA possa inserir
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	var result RefinedTranslation
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		// Fallback: Se o JSON quebrar, tentamos salvar ao menos o texto puro
		return &RefinedTranslation{
			Original: textoBruto,
			Traducao: "Erro ao processar tradução.",
		}, fmt.Errorf("falha ao converter JSON: %v", err)
	}

	return &result, nil
}

// extrairTexto navega pela estrutura de resposta do Google GenAI
func (s *GeminiService) extrairTexto(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) > 0 &&
		resp.Candidates[0].Content != nil &&
		len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	}
	return ""
}

// Close encerra a conexão com o cliente Google
func (s *GeminiService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
