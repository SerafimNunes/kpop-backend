package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"auren-platform/core/events" // Correct Import
	"auren-platform/internal/infrastructure/gemini"

	"auren-platform/internal/agents/admin"
	"auren-platform/internal/agents/engineering"
	"auren-platform/internal/agents/library"
)

// AurenEngine orquestra todos os departamentos e agentes.
type AurenEngine struct {
	// Core Services
	Gemini *gemini.Service
	Hub    *events.Hub // Correct Type

	// Admin Agents
	PropostasAgent  *admin.Propostas
	ComplianceAgent *admin.Compliance
	TaxAdvisorAgent *admin.TaxAdvisor
	HunterAgent     *admin.Hunter

	// Engineering Agents
	AuditorCampoAgent *engineering.AuditorCampo
	RedatorPGRSAgent  *engineering.RedatorPGRS

	// Library Agents
	LibrarianAgent *library.Librarian
	RadarAgent     *library.Radar
}

type ExecutionRequest struct {
	Module  string                 `json:"module"`
	Action  string                 `json:"action"`
	Payload map[string]interface{} `json:"payload"`
}

// NewAurenEngine inicializa o motor central e todos os seus agentes.
func NewAurenEngine(ctx context.Context, hub *events.Hub) (*AurenEngine, error) { // Correct Type
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")

	geminiSvc, err := gemini.NewService(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("falha ao iniciar gemini.Service: %w", err)
	}

	engine := &AurenEngine{
		Gemini: geminiSvc,
		Hub:    hub,

		PropostasAgent:    admin.NewPropostas(geminiSvc),
		ComplianceAgent:   admin.NewCompliance(geminiSvc),
		TaxAdvisorAgent:   admin.NewTaxAdvisor(geminiSvc),
		HunterAgent:       admin.NewHunter(geminiSvc, hub),
		AuditorCampoAgent: engineering.NewAuditorCampo(geminiSvc),
		RedatorPGRSAgent:  engineering.NewRedatorPGRS(geminiSvc),
		LibrarianAgent:    library.NewLibrarian(geminiSvc),
		RadarAgent:        library.NewRadar(geminiSvc, hub),
	}

	log.Println("✅ [ENGINE] Auren Semantic Engine e todos os agentes foram inicializados.")
	return engine, nil
}

// HandleExecution lida com as requisições HTTP da API (legado).
func (e *AurenEngine) HandleExecution(w http.ResponseWriter, r *http.Request) {
	var req ExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	log.Printf("🤖 Dispatcher HTTP: Módulo [%s] | Ação [%s]", req.Module, req.Action)

	var result interface{}
	var err error

	switch req.Module {
	case "library":
		result, err = library.HandleLibraryAction(r.Context(), e.LibrarianAgent, req.Action, req.Payload)
	default:
		err = fmt.Errorf("módulo '%s' via HTTP está obsoleto ou não é reconhecido", req.Module)
	}

	if err != nil {
		log.Printf("❌ Erro HTTP: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (e *AurenEngine) Close() {
	if e.Gemini != nil {
		e.Gemini.Close()
	}
}
