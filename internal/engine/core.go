package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"

	// Imports dos Agentes conforme a nova estrutura do Manifesto 1.4
	"auren-platform/internal/agents/admin"
	"auren-platform/internal/agents/engineering"
	"auren-platform/internal/agents/library"
)

// GeminiService gerencia a conexão bruta com o Google Vertex AI
type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// AurenEngine é a estrutura central que o main.go utiliza (O Cérebro)
type AurenEngine struct {
	Gemini    *GeminiService
	Librarian *library.Librarian // Agente de Pesquisa/RAG
}

// ExecutionRequest define o contrato do Semantic Dispatcher
type ExecutionRequest struct {
	Module  string                 `json:"module"`
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}

// NewAurenEngine inicializa o serviço Gemini e orquestra os sub-serviços
func NewAurenEngine(ctx context.Context) (*AurenEngine, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")

	geminiSvc, err := NewGeminiService(ctx, projectID, location)
	if err != nil {
		return nil, err
	}

	// Inicializa o Bibliotecário (Módulo de Pesquisa)
	libAgent := library.NewLibrarian(geminiSvc)

	return &AurenEngine{
		Gemini:    geminiSvc,
		Librarian: libAgent,
	}, nil
}

// HandleExecution é o coração do Manifesto 1.4.
// Ele roteia a requisição para o agente especialista correto.
func (e *AurenEngine) HandleExecution(w http.ResponseWriter, r *http.Request) {
	var req ExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	log.Printf("🤖 Dispatcher: Módulo [%s] | Ação [%s]", req.Module, req.Action)

	var result interface{}
	var err error

	// Roteamento Semântico baseado nos Departamentos do Manifesto
	switch req.Module {
	case "administrative":
		result, err = admin.HandleAdminAction(r.Context(), e.Gemini, req.Action, req.Payload)
	case "engineering":
		result, err = engineering.HandleEngineeringAction(r.Context(), e.Gemini, req.Action, req.Payload)
	case "library":
		result, err = library.HandleLibraryAction(r.Context(), e.Librarian, req.Action, req.Payload)
	default:
		err = fmt.Errorf("módulo '%s' não reconhecido pela Auren Platform", req.Module)
	}

	if err != nil {
		log.Printf("❌ Erro na execução: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// --- Métodos do GeminiService (Infraestrutura de IA) ---

func NewGeminiService(ctx context.Context, projectID, location string) (*GeminiService, error) {
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

	return &GeminiService{client: client, model: model}, nil
}

func (s *GeminiService) Generate(ctx context.Context, parts []genai.Part, systemInstruction string) (string, error) {
	var executor *genai.GenerativeModel
	if systemInstruction != "" {
		executor = s.client.GenerativeModel("gemini-2.0-flash-001")
		executor.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemInstruction)}}
		executor.SetTemperature(0.0)
	} else {
		executor = s.model
	}

	resp, err := executor.GenerateContent(ctx, parts...)
	if err != nil {
		return "", err
	}
	return s.extractText(resp), nil
}

func (s *GeminiService) ExecuteSimplePrompt(ctx context.Context, prompt string) (string, error) {
	return s.Generate(ctx, []genai.Part{genai.Text(prompt)}, "")
}

func (s *GeminiService) extractText(resp *genai.GenerateContentResponse) string {
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

func (e *AurenEngine) Close() {
	if e.Gemini != nil {
		e.Gemini.Close()
	}
}

func (s *GeminiService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}
