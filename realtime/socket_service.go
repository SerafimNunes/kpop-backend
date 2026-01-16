package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"auren-platform/ai"
	"auren-platform/db"
	"auren-platform/media_analysis"

	"cloud.google.com/go/vertexai/genai"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	htmlRegex = regexp.MustCompile("<[^>]*>")
)

func ServeWS(h *Hub, engine *ai.GeminiService, sap *media_analysis.StreamAudioProcessor, vp *media_analysis.VideoProcessor, fap *media_analysis.FileAudioProcessor, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Erro upgrade: %v", err)
		return
	}

	client := &Client{
		ID:   uuid.New().String()[:8],
		Conn: make(chan Message, 512),
	}
	h.Register <- client

	var currentAction string
	var currentFileName string
	var currentVistoriaID uint
	var currentMimeType string

	defer func() {
		h.Unregister <- client
		conn.Close()
	}()

	go func() {
		for msg := range client.Conn {
			if err := conn.WriteJSON(msg); err != nil {
				break
			}
		}
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)

		if messageType == websocket.BinaryMessage {
			switch currentAction {
			case "upload_knowledge":
				go handleKnowledgeUpload(ctx, cancel, client, engine, p, currentFileName)
			case "upload_evidence":
				go handleMediaEvidence(ctx, cancel, client, engine, vp, fap, p, currentVistoriaID, currentMimeType)
			default:
				cancel()
			}
			currentAction = ""
			currentMimeType = ""
			continue
		}

		var req map[string]interface{}
		if err := json.Unmarshal(p, &req); err != nil {
			cancel()
			continue
		}

		action, _ := req["action"].(string)

		switch action {
		case "media_capture_start":
			currentAction = "upload_evidence"
			if id, ok := req["vistoriaID"].(float64); ok {
				currentVistoriaID = uint(id)
			}
			if mime, ok := req["mimeType"].(string); ok {
				currentMimeType = mime
			}
			client.Conn <- Message{Type: "status", Payload: "Aguardando arquivo de mídia..."}
			cancel()

		case "sync_vistoria":
			go handleVistoriaSync(ctx, cancel, client, req)

		case "upload_knowledge":
			currentAction = "upload_knowledge"
			currentFileName, _ = req["fileName"].(string)
			client.Conn <- Message{Type: "status", Payload: "Aguardando arquivo..."}
			cancel()

		case "process_intelligence_source":
			payload, _ := req["payload"].(map[string]interface{})
			url, _ := payload["url"].(string)
			go handleExternalInsight(ctx, cancel, client, engine, sap, url)

		case "generate_proposal":
			cliente, _ := req["cliente"].(string)
			servico, _ := req["servico"].(string)
			detalhes, _ := req["detalhes"].(string)
			go handleProposalGeneration(ctx, cancel, client, engine, cliente, servico, detalhes)

		case "generate_contract":
			propostaOriginal, _ := req["propostaAtual"].(string)
			go handleContractGeneration(ctx, cancel, client, engine, propostaOriginal)

		case "generate_pgrs_report":
			payload, _ := req["payload"].(map[string]interface{})
			go handlePGRSGeneration(ctx, cancel, client, engine, payload)

		case "consolidate_inspection":
			payload, _ := req["payload"].(map[string]interface{})
			go handleConsolidateInspection(ctx, cancel, client, engine, payload)

		default:
			cancel()
		}
	}
}

// --- HANDLERS ---

func handleMediaEvidence(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, vp *media_analysis.VideoProcessor, fap *media_analysis.FileAudioProcessor, data []byte, vistoriaID uint, mimeType string) {
	defer cancel()

	var processedData []byte
	var fileExtension string
	var evidenceType string
	var err error

	switch {
	case strings.HasPrefix(mimeType, "video/"):
		client.Conn <- Message{Type: "status", Payload: "Processando vídeo..."}
		processedData, err = vp.ProcessVideo(data)
		if err != nil {
			client.Conn <- Message{Type: "error", Payload: fmt.Sprintf("Erro ao processar vídeo: %v", err)}
			return
		}
		fileExtension = ".jpg"
		evidenceType = "video_frame"
	case strings.HasPrefix(mimeType, "audio/"):
		client.Conn <- Message{Type: "status", Payload: "Analisando áudio..."}
		classification, err := fap.ProcessAudioFile(data)
		if err != nil {
			client.Conn <- Message{Type: "error", Payload: fmt.Sprintf("Erro ao processar áudio: %v", err)}
			return
		}

		if classification == "voice" {
			// If voice is detected, we can save the original audio and send it for transcription
			processedData = data
			fileExtension = ".wav" // Assuming wav for now
			evidenceType = "audio"
			client.Conn <- Message{Type: "technical_insight", Payload: "Voz humana detectada no áudio."}
		} else {
			client.Conn <- Message{Type: "status", Payload: fmt.Sprintf("Áudio classificado como '%s'. Evidência não gerada.", classification)}
			return // Do not save if no voice is detected
		}

	case strings.HasPrefix(mimeType, "image/"):
		processedData = data
		fileExtension = ".jpg"
		evidenceType = "imagem"
	default:
		client.Conn <- Message{Type: "error", Payload: fmt.Sprintf("Tipo de mídia não suportado: %s", mimeType)}
		return
	}

	client.Conn <- Message{Type: "status", Payload: "Analisando evidência com IA..."}
	fileName := fmt.Sprintf("evid_%d_%d%s", vistoriaID, time.Now().Unix(), fileExtension)
	path := filepath.Join("./storage/evidences", fileName)
	if err := os.WriteFile(path, processedData, 0644); err != nil {
		client.Conn <- Message{Type: "error", Payload: fmt.Sprintf("Falha ao salvar evidência: %v", err)}
		return
	}

	auditor := ai.NewAuditor(engine)
	// We send the processed data (image) to the AI for analysis
	res, err := auditor.AnalyzeFieldEvidence(ctx, processedData, "image/jpeg", "Análise de conformidade de campo.")
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: fmt.Sprintf("Falha na análise da IA: %v", err)}
		return
	}

	db.DB.WithContext(ctx).Create(&db.Evidencia{
		VistoriaID: vistoriaID,
		Tipo:       evidenceType,
		StorageURL: path,
	})

	var vistoria db.Vistoria
	if err := db.DB.WithContext(ctx).First(&vistoria, vistoriaID).Error; err == nil {
		vistoria.NotasTecnicas += fmt.Sprintf("\n[IA %s]: %s", time.Now().Format("15:04"), res)
		db.DB.WithContext(ctx).Save(&vistoria)
	}

	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handleVistoriaSync(ctx context.Context, cancel context.CancelFunc, client *Client, data map[string]interface{}) {
	defer cancel()
	checklist, _ := data["checklist"].(map[string]interface{})
	unidadeID := uint(1) // Em produção, extrair do contexto do usuário

	var vistoria db.Vistoria
	db.DB.WithContext(ctx).FirstOrCreate(&vistoria, db.Vistoria{UnidadeID: unidadeID})

	vistoria.CheckSegregacao = checklist["segregacao"].(bool)
	vistoria.CheckArmazenamento = checklist["armazenamento"].(bool)
	vistoria.CheckIdentificacao = checklist["identificacao"].(bool)
	vistoria.CheckContencao = checklist["contencao"].(bool)
	vistoria.Data = time.Now()

	db.DB.WithContext(ctx).Save(&vistoria)
	client.Conn <- Message{Type: "sync_ok", Payload: "Vistoria sincronizada."}
}

func handleProposalGeneration(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, cliente, servico, detalhes string) {
	defer cancel()
	client.Conn <- Message{Type: "status", Payload: "Redigindo Proposta Comercial..."}

	business := ai.NewBusinessAgent(engine)
	res, err := business.GenerateProposal(ctx, cliente, servico, detalhes)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Erro na IA."}
		return
	}

	// Persistência Real
	novaProposta := db.Proposta{
		ClienteNome:     cliente,
		Titulo:          fmt.Sprintf("Proposta %s - %s", cliente, servico),
		ServicoTipo:     servico,
		DescricaoEscopo: detalhes,
		SumarioIA:       res,
		Status:          "RASCUNHO",
	}
	db.DB.WithContext(ctx).Create(&novaProposta)

	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handleContractGeneration(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, propostaOriginal string) {
	defer cancel()
	client.Conn <- Message{Type: "status", Payload: "Gerando Minuta Contratual..."}

	business := ai.NewBusinessAgent(engine)
	res, err := business.GenerateContract(ctx, propostaOriginal)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Erro ao gerar contrato."}
		return
	}

	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handleKnowledgeUpload(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, data []byte, fileName string) {
	defer cancel()
	storagePath := fmt.Sprintf("./storage/knowledge/%d_%s", time.Now().Unix(), fileName)
	_ = os.WriteFile(storagePath, data, 0644)

	librarian := ai.NewLibrarian(engine)
	text, _ := librarian.ExtractTextFromPDF(storagePath)
	res, _ := librarian.ProcessKnowledge(ctx, text, fileName)

	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handleExternalInsight(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, sap *media_analysis.StreamAudioProcessor, url string) {
	if strings.Contains(url, "youtu") {
		handleYoutubeInsight(ctx, cancel, client, engine, sap, url)
	} else {
		// This function now uses the utility function from media_analysis
		text, err := media_analysis.FetchWebText(ctx, url)
		if err != nil {
			client.Conn <- Message{Type: "error", Payload: "Falha ao buscar conteúdo da web."}
			return
		}
		handleWebDocumentInsight(ctx, cancel, client, engine, text)
	}
}

func handleWebDocumentInsight(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, text string) {
	defer cancel()
	
	if len(text) > 30000 { text = text[:30000] }

	prompt := fmt.Sprintf("Analise este conteúdo regulatório: %s", text)
	res, _ := engine.Generate(ctx, []genai.Part{genai.Text(prompt)}, "Resumo Técnico ESG")
	
	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handleYoutubeInsight(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, sap *media_analysis.StreamAudioProcessor, url string) {
	defer cancel()
	audioChan := make(chan []byte, 1)
	if err := sap.StartDigitalEar(ctx, url, audioChan); err != nil { 
		client.Conn <- Message{Type: "error", Payload: "Falha ao iniciar escuta no YouTube."}
		return 
	}

	select {
	case chunk := <-audioChan:
		res, _ := engine.Generate(ctx, []genai.Part{genai.Blob{MIMEType: "audio/wav", Data: chunk}}, "Extração de fatos ambientais")
		client.Conn <- Message{Type: "technical_insight", Payload: res}
	case <-ctx.Done():
	}
}

func handlePGRSGeneration(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, data map[string]interface{}) {
	defer cancel()
	client.Conn <- Message{Type: "status", Payload: "Dados recebidos. Iniciando auditoria de conformidade..."}

	auditor := ai.NewAuditor(engine)
	redator := ai.NewRedator(engine)

	jsonData, err := json.Marshal(data)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Falha ao processar os dados do formulário."}
		return
	}

	client.Conn <- Message{Type: "status", Payload: "Analisando inventário de resíduos contra a NBR 10.004..."}
	auditResult, err := auditor.AnalyzePGRSCompliance(ctx, jsonData)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Falha na auditoria da IA: " + err.Error()}
		return
	}

	client.Conn <- Message{Type: "technical_insight", Payload: auditResult}
	client.Conn <- Message{Type: "status", Payload: "Auditoria concluída com sucesso. Redigindo o memorial descritivo..."}

	report, err := redator.GeneratePGRSReport(ctx, jsonData, auditResult)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Falha na redação do laudo pela IA: " + err.Error()}
		return
	}

	client.Conn <- Message{Type: "pgrs_report_ready", Payload: report}
}

func handleConsolidateInspection(ctx context.Context, cancel context.CancelFunc, client *Client, engine *ai.GeminiService, data map[string]interface{}) {
	defer cancel()
	client.Conn <- Message{Type: "status", Payload: "Backend recebido. Persistindo dados da vistoria..."}

	setor, _ := data["setor"].(string)
	observacoes, _ := data["observacoes"].(string)
	checklist, _ := data["checklist"].(map[string]interface{})

	auditor := ai.NewAuditor(engine)
	iaAnalysis, err := auditor.AnalyzeInspectionNotes(ctx, observacoes)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Falha na análise da IA: " + err.Error()}
		return
	}

	client.Conn <- Message{Type: "status", Payload: "Análise da IA concluída. Consolidando..."}

	unidadeID := uint(1)
	var vistoria db.Vistoria
	db.DB.WithContext(ctx).FirstOrCreate(&vistoria, db.Vistoria{UnidadeID: unidadeID})

	vistoria.Setor = setor
	vistoria.Data = time.Now()
	vistoria.NotasTecnicas = observacoes
	vistoria.TranscricaoIA = iaAnalysis
	if chk, ok := checklist["segregacao"].(bool); ok {
		vistoria.CheckSegregacao = chk
	}
	if chk, ok := checklist["armazenamento"].(bool); ok {
		vistoria.CheckArmazenamento = chk
	}
	if chk, ok := checklist["identificacao"].(bool); ok {
		vistoria.CheckIdentificacao = chk
	}
	if chk, ok := checklist["vazamento"].(bool); ok {
		vistoria.CheckContencao = chk
	}
	db.DB.WithContext(ctx).Save(&vistoria)

	syncPayload := map[string]interface{}{
		"setor": setor,
		"residuos_identificados": []map[string]string{
			{"nome": "Resíduo Exemplo (derivado da análise IA)", "estimativa_mensal": "10kg"},
		},
		"observacoes_consolidadas": vistoria.NotasTecnicas + "\n\n**Auditoria IA:**\n" + vistoria.TranscricaoIA,
	}

	client.Conn <- Message{Type: "inspection_analysis_result", Payload: syncPayload}
}
