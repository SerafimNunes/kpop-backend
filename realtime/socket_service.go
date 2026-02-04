package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"auren-platform/core/events"
	"auren-platform/db"
	"auren-platform/media_analysis"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	htmlRegex = regexp.MustCompile("<[^>]*>")
)

// Librarian defines the interface for the librarian agent.
type Librarian interface {
	ProcessKnowledge(ctx context.Context, rawContent string, source string) (string, error)
	Search(ctx context.Context, query string) (interface{}, error)
	ExtractTextFromPDF(filePath string) (string, error)
}

// ServeWS agora recebe as dependências injetadas individualmente para evitar ciclos de importação com o pacote engine.
func ServeWS(
	hub *events.Hub,
	geminiSvc any, // Pode ser tipado como *gemini.Service se importar infra
	agents map[string]any, // Mapa contendo os agentes instanciados
	sap *media_analysis.StreamAudioProcessor,
	vp *media_analysis.VideoProcessor,
	fap *media_analysis.FileAudioProcessor,
	w http.ResponseWriter,
	r *http.Request,
) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Erro upgrade: %v", err)
		return
	}
	log.Println("[WS] Cliente conectado com sucesso.")

	// Uso obrigatório do prefixo events. devido ao package diferente
	client := &events.Client{
		ID:   uuid.New().String()[:8],
		Conn: make(chan events.Message, 512),
	}
	hub.Register <- client

	var currentAction string
	var currentFileName string
	var currentVistoriaID uint
	var currentMimeType string

	defer func() {
		hub.Unregister <- client
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
				// Handler deve ser adaptado para receber os agentes do mapa
				go handleKnowledgeUpload(ctx, cancel, client, agents, p, currentFileName)
			case "upload_evidence":
				go handleMediaEvidence(ctx, cancel, client, agents, vp, fap, p, currentVistoriaID, currentMimeType)
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

		log.Printf("[WS] Mensagem Recebida: %s", string(p))
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
			client.Conn <- events.Message{Type: "status", Payload: "Aguardando arquivo de mídia..."}
			cancel()

		case "vistoria_request_sync":
			go handleVistoriaSync(ctx, cancel, client)

		case "financeiro_request_sync":
			go handleFinanceSync(ctx, cancel, client)

		case "comercial_request_sync":
			go handleComercialSync(ctx, cancel, client)

		case "create_transaction":
			payload, _ := req["payload"].(map[string]interface{})
			go handleCreateTransaction(ctx, cancel, client, hub, payload)

		case "dashboard_request_sync":
			go handleDashboardSync(ctx, cancel, client)

		case "engenharia_request_sync":
			go handleEngenhariaSync(ctx, cancel, client)

		case "consolidate_inspection":
			payload, _ := req["payload"].(map[string]interface{})
			go handleConsolidateInspection(ctx, cancel, client, agents, payload)

		case "generate_proposal", "generate_contract":
			payload, _ := req["payload"].(map[string]interface{})
			go handlePropostas(ctx, cancel, client, agents, action, payload)

		case "process_intelligence_source":
			payload, _ := req["payload"].(map[string]interface{})
			go handleIntelligenceSource(ctx, cancel, client, agents, payload)

		case "ai_chat_query":
			payload, _ := req["payload"].(map[string]interface{})
			go handleAIChatQuery(ctx, cancel, client, agents, payload)

		case "upload_knowledge":
			currentAction = "upload_knowledge"
			if name, ok := req["fileName"].(string); ok {
				currentFileName = name
			}
			client.Conn <- events.Message{Type: "status", Payload: fmt.Sprintf("Aguardando upload de %s", currentFileName)}
			cancel()

		case "generate_pgrs_report":
			payload, _ := req["payload"].(map[string]interface{})
			go handleGeneratePGRSReport(ctx, cancel, client, agents, payload)

		default:
			cancel()
		}
	}
}

// --- HANDLERS ADAPTADOS ---

func handleIntelligenceSource(ctx context.Context, cancel context.CancelFunc, client *events.Client, agents map[string]any, payload map[string]interface{}) {
	defer cancel()

	librarian, ok := agents["librarian"].(Librarian)
	if !ok {
		client.Conn <- events.Message{Type: "error", Payload: "Agente de Biblioteca não configurado."}
		return
	}

	source, _ := payload["url"].(string)
	client.Conn <- events.Message{Type: "socket:ai_progress", Payload: map[string]interface{}{"percentage": 25, "message": "Analisando fonte..."}}

	// Aqui, precisaríamos de uma função para extrair o conteúdo da URL.
	// Por enquanto, vamos simular, e usar o ProcessKnowledge.
	// Em um caso real: content, err := web_scraper.Scrape(source)
	content := "Simulação de conteúdo extraído da URL: " + source

	result, err := librarian.ProcessKnowledge(ctx, content, source)
	if err != nil {
		client.Conn <- events.Message{Type: "error", Payload: fmt.Sprintf("Erro do Engine de IA: %v", err)}
		return
	}
	client.Conn <- events.Message{Type: "socket:ai_progress", Payload: map[string]interface{}{"percentage": 80, "message": "Gerando insights..."}}

	// A resposta do Gemini pode vir com markdown, vamos apenas enviar o resumo por enquanto
	insightData := map[string]string{
		"summary":     result,
		"pgrsImpact":  "Impacto a ser determinado pela IA.",
		"legalBase":   "Base legal a ser determinada pela IA.",
		"opportunity": "Oportunidade a ser determinada pela IA.",
	}

	client.Conn <- events.Message{Type: "socket:technical_insight", Payload: insightData}
}

func handleAIChatQuery(ctx context.Context, cancel context.CancelFunc, client *events.Client, agents map[string]any, payload map[string]interface{}) {
	defer cancel()
	librarian, ok := agents["librarian"].(Librarian)
	if !ok {
		client.Conn <- events.Message{Type: "error", Payload: "Agente de Biblioteca não configurado."}
		return
	}

	query, _ := payload["query"].(string)
	result, err := librarian.Search(ctx, query)
	if err != nil {
		client.Conn <- events.Message{Type: "error", Payload: fmt.Sprintf("Erro do Engine de IA: %v", err)}
		return
	}

	response, _ := result.(string)
	client.Conn <- events.Message{Type: "socket:ai_chat_response", Payload: map[string]string{"content": response}}
}

func handleKnowledgeUpload(ctx context.Context, cancel context.CancelFunc, client *events.Client, agents map[string]any, data []byte, fileName string) {
	defer cancel()

	librarian, ok := agents["librarian"].(Librarian)
	if !ok {
		client.Conn <- events.Message{Type: "error", Payload: "Agente de Biblioteca não configurado."}
		return
	}

	filePath := fmt.Sprintf("./storage/knowledge/%s", fileName)
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		client.Conn <- events.Message{Type: "error", Payload: fmt.Sprintf("Falha ao salvar arquivo: %v", err)}
		return
	}
	client.Conn <- events.Message{Type: "socket:ai_progress", Payload: map[string]interface{}{"percentage": 20, "message": "Arquivo salvo. Extraindo texto..."}}

	var rawContent string
	// Supondo que só estamos lidando com PDFs por agora, como em pesquisa.js
	if regexp.MustCompile(`\.pdf$`).MatchString(strings.ToLower(fileName)) {
		rawContent, err = librarian.ExtractTextFromPDF(filePath)
		if err != nil {
			client.Conn <- events.Message{Type: "error", Payload: fmt.Sprintf("Falha ao extrair texto do PDF: %v", err)}
			return
		}
	} else {
		rawContent = string(data)
	}

	client.Conn <- events.Message{Type: "socket:ai_progress", Payload: map[string]interface{}{"percentage": 50, "message": "Texto extraído. Processando com IA..."}}

	result, err := librarian.ProcessKnowledge(ctx, rawContent, fileName)
	if err != nil {
		client.Conn <- events.Message{Type: "error", Payload: fmt.Sprintf("Erro do Engine de IA: %v", err)}
		return
	}
	client.Conn <- events.Message{Type: "socket:ai_progress", Payload: map[string]interface{}{"percentage": 80, "message": "Gerando insights..."}}

	insightData := map[string]string{
		"summary":     result,
		"pgrsImpact":  "Impacto a ser determinado pela IA.",
		"legalBase":   "Base legal a ser determinada pela IA.",
		"opportunity": "Oportunidade a ser determinada pela IA.",
	}
	client.Conn <- events.Message{Type: "socket:technical_insight", Payload: insightData}
}

func handlePropostas(ctx context.Context, cancel context.CancelFunc, client *events.Client, agents map[string]any, action string, payload map[string]interface{}) {
	defer cancel()

	propostasAgent, ok := agents["propostas"].(interface {
		GerarProposta(context.Context, string, string, string) (string, error)
		AnalisarContrato(context.Context, string) (string, error)
	})
	if !ok {
		client.Conn <- events.Message{Type: "error", Payload: "Agente de Propostas não configurado."}
		return
	}

	client.Conn <- events.Message{Type: "status", Payload: "Analisando contexto e requisitos..."}

	var result string
	var err error

	if action == "generate_proposal" {
		cliente, _ := payload["cliente"].(string)
		servico, _ := payload["servico"].(string)
		detalhes, _ := payload["detalhes"].(string)
		result, err = propostasAgent.GerarProposta(ctx, cliente, servico, detalhes)
	} else { // generate_contract
		contentBase, _ := payload["contentBase"].(string)
		result, err = propostasAgent.AnalisarContrato(ctx, contentBase)
	}

	if err != nil {
		client.Conn <- events.Message{Type: "error", Payload: fmt.Sprintf("Erro do Engine de IA: %v", err)}
		return
	}

	client.Conn <- events.Message{Type: "status", Payload: "Estruturando documento final..."}
	time.Sleep(1 * time.Second) // Simula o trabalho de formatação

	// Remove tags HTML se houver, para garantir que o front-end formate
	cleanResult := htmlRegex.ReplaceAllString(result, "")

	client.Conn <- events.Message{Type: "technical_insight", Payload: cleanResult}
}

func handleFinanceSync(ctx context.Context, cancel context.CancelFunc, client *events.Client) {
	defer cancel()

	var transactions []db.LancamentoFinanceiro
	if err := db.DB.WithContext(ctx).Order("date desc").Limit(20).Find(&transactions).Error; err != nil {
		client.Conn <- events.Message{Type: "error", Payload: "Falha ao buscar transações."}
		return
	}

	var totalRevenue, totalExpenses float64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	db.DB.WithContext(ctx).Model(&db.LancamentoFinanceiro{}).Where("type = ? AND date >= ?", "entrada", startOfMonth).Select("coalesce(sum(amount), 0)").Row().Scan(&totalRevenue)
	db.DB.WithContext(ctx).Model(&db.LancamentoFinanceiro{}).Where("type = ? AND date >= ?", "saida", startOfMonth).Select("coalesce(sum(amount), 0)").Row().Scan(&totalExpenses)

	payload := map[string]interface{}{
		"metrics": []map[string]interface{}{
			{"label": "Receita Bruta", "value": totalRevenue},
			{"label": "Despesas", "value": totalExpenses},
		},
		"transactions": transactions,
	}

	client.Conn <- events.Message{Type: "financeiro_init", Payload: payload}
}

func handleCreateTransaction(ctx context.Context, cancel context.CancelFunc, client *events.Client, hub *events.Hub, payload map[string]interface{}) {
	defer cancel()
	// Lógica de DB omitida para brevidade, mas segue o mesmo padrão
	hub.Broadcast <- events.Message{Type: "financeiro_update", Payload: payload}
}

func handleVistoriaSync(ctx context.Context, cancel context.CancelFunc, client *events.Client) {
	defer cancel()

	var vistorias []db.Vistoria
	if err := db.DB.WithContext(ctx).Order("data desc").Limit(10).Find(&vistorias).Error; err != nil {
		client.Conn <- events.Message{Type: "error", Payload: "Falha ao buscar vistorias."}
		return
	}

	inspections := []map[string]interface{}{}
	for _, v := range vistorias {
		status := "Em Análise"
		if v.CheckSegregacao && v.CheckArmazenamento && v.CheckIdentificacao && v.CheckContencao {
			status = "Consolidado"
		}
		inspections = append(inspections, map[string]interface{}{
			"id":          fmt.Sprintf("VIS-%d", v.ID),
			"setor":       v.Setor,
			"responsavel": v.Tecnico,
			"data":        v.Data.Format("2006-01-02"),
			"status":      status,
		})
	}

	payload := map[string]interface{}{
		"inspections": inspections,
	}
	client.Conn <- events.Message{Type: "vistoria_init", Payload: payload}
}

func handleConsolidateInspection(ctx context.Context, cancel context.CancelFunc, client *events.Client, agents map[string]any, data map[string]interface{}) {
	defer cancel()
	// Exemplo de como usar o cliente injetado corretamente
	syncPayload := map[string]interface{}{"status": "consolidado"}
	client.Conn <- events.Message{Type: "inspection_analysis_result", Payload: syncPayload}
}

// Stubs para os handlers de IA (devem ser implementados extraindo o agente do map 'agents')

func handleMediaEvidence(ctx context.Context, cancel context.CancelFunc, client *events.Client, agents map[string]any, vp *media_analysis.VideoProcessor, fap *media_analysis.FileAudioProcessor, data []byte, vistoriaID uint, mimeType string) {
	defer cancel()
}

func handleDashboardSync(ctx context.Context, cancel context.CancelFunc, client *events.Client) {
	defer cancel()

	var criticalProjects, attentionProjects, compliantProjects int64
	db.DB.WithContext(ctx).Model(&db.AuditoriaExecucao{}).Where("exposicao_level = ?", "Crítico").Count(&criticalProjects)
	db.DB.WithContext(ctx).Model(&db.AuditoriaExecucao{}).Where("exposicao_level = ?", "Atenção").Count(&attentionProjects)
	db.DB.WithContext(ctx).Model(&db.AuditoriaExecucao{}).Where("status = ?", "Em Conformidade").Count(&compliantProjects)

	semaforo := []map[string]interface{}{
		{"label": "Projetos Críticos", "count": criticalProjects, "action": "engenharia", "colorClass": "text-red-500", "bgClass": "bg-red-500 animate-pulse"},
		{"label": "Prazos em Atenção", "count": attentionProjects, "action": "vistoria", "colorClass": "text-amber-500", "bgClass": "bg-amber-500"},
		{"label": "Em Conformidade", "count": compliantProjects, "action": "dashboard", "colorClass": "text-emerald-500", "bgClass": "bg-emerald-500"},
	}

	var proposals []db.Proposta
	db.DB.WithContext(ctx).Where("status IN ?", []string{"ENVIADA", "RASCUNHO"}).Order("created_at desc").Limit(5).Find(&proposals)
	leads := []map[string]interface{}{}
	for _, p := range proposals {
		leads = append(leads, map[string]interface{}{
			"id":      p.ID,
			"empresa": p.ClienteNome,
			"score":   95, // Mocked for now
			"tag":     p.ServicoTipo,
			"url":     "#",
		})
	}

	var insights []db.IntelligenceInsight
	db.DB.WithContext(ctx).Order("created_at desc").Limit(5).Find(&insights)
	radarNews := []map[string]interface{}{}
	for _, i := range insights {
		radarNews = append(radarNews, map[string]interface{}{
			"id":      i.ID,
			"tag":     i.LegalBase,
			"title":   i.Summary,
			"summary": i.PgrsImpact,
			"url":     "#",
		})
	}

	var projects []db.Proposta
	db.DB.WithContext(ctx).Order("updated_at desc").Limit(5).Find(&projects)
	projetos := []map[string]interface{}{}
	for _, p := range projects {
		statusClass := "bg-blue-50 text-blue-600 border-blue-100"
		if p.Status == "ACEITA" {
			statusClass = "bg-emerald-50 text-emerald-600 border-emerald-100"
		} else if p.Status == "RECUSADA" {
			statusClass = "bg-red-50 text-red-600 border-red-100"
		}

		progresso := 50
		if p.Status == "ACEITA" {
			progresso = 100
		}

		projetos = append(projetos, map[string]interface{}{
			"id":          fmt.Sprintf("PRJ-%d", p.ID),
			"cliente":     p.ClienteNome,
			"status":      p.Status,
			"statusClass": statusClass,
			"progresso":   progresso, // Mocked for now
			"barClass":    "bg-emerald-500",
			"responsavel": "Eng. Responsável", // Mocked for now
		})
	}

	payload := map[string]interface{}{
		"semaforo":  semaforo,
		"leads":     leads,
		"radarNews": radarNews,
		"projetos":  projetos,
		"stats": map[string]interface{}{
			"tempoIA": 0.5,
		},
	}

	client.Conn <- events.Message{Type: "dashboard_init", Payload: payload}
}

func handleComercialSync(ctx context.Context, cancel context.CancelFunc, client *events.Client) {
	defer cancel()
	// In a real scenario, this would be a DB query.
	// e.g., db.DB.Model(&db.Proposal{}).Where("status = ?", "active").Select("SUM(value)").Row().Scan(&pipelineValue)
	pipelineValue := 195300.50

	payload := map[string]interface{}{
		"pipelineValue": pipelineValue,
	}
	client.Conn <- events.Message{Type: "comercial_init", Payload: payload}
}

func handleEngenhariaSync(ctx context.Context, cancel context.CancelFunc, client *events.Client) {
	defer cancel()

	var proposals []db.Proposta
	db.DB.WithContext(ctx).Where("servico_tipo IN ?", []string{"PGRS Profissional", "Auditoria"}).Order("created_at desc").Limit(5).Find(&proposals)

	projects := []map[string]interface{}{}
	for _, p := range proposals {
		projects = append(projects, map[string]interface{}{
			"id":         fmt.Sprintf("PRJ-%d", p.ID),
			"cliente":    p.ClienteNome,
			"status":     p.Status,
			"compliance": 85, // Mocked for now
		})
	}

	payload := map[string]interface{}{
		"projects": projects,
	}
	client.Conn <- events.Message{Type: "engenharia_init", Payload: payload}
}

// handleGeneratePGRSReport coordena todo o processo de geração de um PGRS
func handleGeneratePGRSReport(
	ctx context.Context,
	cancel context.CancelFunc,
	client *events.Client,
	agents map[string]any,
	payload map[string]interface{},
) {
	defer cancel()

	log.Println("🚀 [PGRS] Iniciando geração de relatório PGRS...")

	// ═══════════════════════════════════════════════════════════
	// ETAPA 1: VALIDAÇÃO E EXTRAÇÃO DE DADOS
	// ═══════════════════════════════════════════════════════════

	client.Conn <- events.Message{Type: "status", Payload: "Validando dados do formulário..."}

	razaoSocial, _ := payload["razao"].(string)
	cnpj, _ := payload["cnpj"].(string)
	cnae, _ := payload["cnae"].(string)

	if razaoSocial == "" || cnpj == "" {
		log.Println("❌ [PGRS] Erro: Dados obrigatórios faltando")
		client.Conn <- events.Message{
			Type:    "error",
			Payload: "Dados obrigatórios faltando: Razão Social e CNPJ são obrigatórios.",
		}
		return
	}

	// Extrair inventário
	inventoryRaw, ok := payload["inventory"].([]interface{})
	if !ok || len(inventoryRaw) == 0 {
		log.Println("❌ [PGRS] Erro: Inventário vazio")
		client.Conn <- events.Message{
			Type:    "error",
			Payload: "Inventário de resíduos não pode estar vazio. Adicione pelo menos um item.",
		}
		return
	}

	// Converter inventário para formato estruturado
	var inventory []map[string]string
	for _, item := range inventoryRaw {
		if itemMap, ok := item.(map[string]interface{}); ok {
			inventory = append(inventory, map[string]string{
				"nome":    fmt.Sprintf("%v", itemMap["nome"]),
				"classe":  fmt.Sprintf("%v", itemMap["classe"]),
				"destino": fmt.Sprintf("%v", itemMap["destino"]),
			})
		}
	}

	// Dados de auditoria
	auditScore, _ := payload["auditScore"].(float64)
	exposureLevel, _ := payload["exposure"].(string)

	log.Printf("✅ [PGRS] Dados validados - Empresa: %s | CNPJ: %s | Inventário: %d itens", razaoSocial, cnpj, len(inventory))

	// ═══════════════════════════════════════════════════════════
	// ETAPA 2: PREPARAR DADOS ESTRUTURADOS
	// ═══════════════════════════════════════════════════════════

	pgrsData := map[string]interface{}{
		"razao_social":   razaoSocial,
		"cnpj":           cnpj,
		"cnae":           cnae,
		"inventory":      inventory,
		"audit_score":    auditScore,
		"exposure_level": exposureLevel,
		"data_geracao":   time.Now().Format("02/01/2006"),
		"num_residuos":   len(inventory),
	}

	pgrsDataJSON, err := json.Marshal(pgrsData)
	if err != nil {
		log.Printf("❌ [PGRS] Erro ao serializar dados: %v", err)
		client.Conn <- events.Message{Type: "error", Payload: "Erro ao processar dados do formulário."}
		return
	}

	// ═══════════════════════════════════════════════════════════
	// ETAPA 3: OBTER AGENTES NECESSÁRIOS
	// ═══════════════════════════════════════════════════════════

	type Auditor interface {
		AnalisarPGRSCompliance(ctx context.Context, pgrsDataJSON []byte) (string, error)
	}

	type Redator interface {
		GerarRelatorioPGRS(ctx context.Context, pgrsDataJSON []byte, auditResult string) (string, error)
	}

	auditorAgent, ok1 := agents["auditor"].(Auditor)
	redatorAgent, ok2 := agents["redator"].(Redator)

	if !ok1 || !ok2 {
		log.Println("❌ [PGRS] Erro: Agentes técnicos não disponíveis")
		client.Conn <- events.Message{
			Type:    "error",
			Payload: "Erro interno: Agentes de IA não estão configurados corretamente.",
		}
		return
	}

	// ═══════════════════════════════════════════════════════════
	// ETAPA 4: EXECUTAR AUDITORIA TÉCNICA
	// ═══════════════════════════════════════════════════════════

	client.Conn <- events.Message{
		Type:    "status",
		Payload: "Fase 1/3: Executando Auditoria de Compliance NBR 10.004...",
	}

	log.Println("🔍 [PGRS] Executando auditoria de compliance via IA...")

	auditResult, err := auditorAgent.AnalisarPGRSCompliance(ctx, pgrsDataJSON)
	if err != nil {
		log.Printf("❌ [PGRS] Erro na auditoria: %v", err)
		client.Conn <- events.Message{
			Type:    "error",
			Payload: fmt.Sprintf("Erro na auditoria técnica: %v. Tente novamente.", err),
		}
		return
	}

	log.Println("✅ [PGRS] Auditoria concluída com sucesso")

	// Enviar insight intermediário ao frontend
	client.Conn <- events.Message{
		Type:    "socket:technical_insight",
		Payload: auditResult,
	}

	// ═══════════════════════════════════════════════════════════
	// ETAPA 5: GERAR RELATÓRIO VIA IA
	// ═══════════════════════════════════════════════════════════

	client.Conn <- events.Message{
		Type:    "status",
		Payload: "Fase 2/3: Redação Técnica do Memorial Descritivo...",
	}

	log.Println("✍️  [PGRS] Gerando relatório técnico via Gemini AI...")

	relatorioFinal, err := redatorAgent.GerarRelatorioPGRS(ctx, pgrsDataJSON, auditResult)
	if err != nil {
		log.Printf("❌ [PGRS] Erro na redação: %v", err)
		client.Conn <- events.Message{
			Type:    "error",
			Payload: fmt.Sprintf("Erro ao gerar relatório: %v. Tente novamente.", err),
		}
		return
	}

	log.Println("✅ [PGRS] Relatório gerado com sucesso")

	// ═══════════════════════════════════════════════════════════
	// ETAPA 6: PERSISTIR NO BANCO DE DADOS (TODO)
	// ═══════════════════════════════════════════════════════════

	client.Conn <- events.Message{
		Type:    "status",
		Payload: "Fase 3/3: Salvando no Sistema...",
	}

	// Criar PGRS no banco
	pgrsRecord, err := db.CriarPGRS(1, pgrsData) // TODO: Pegar unidadeID real
	if err != nil {
		log.Printf("⚠️ [PGRS] Erro ao criar registro: %v", err)
	} else {
		// Salvar resultados da IA
		err = db.SalvarResultadoIA(pgrsRecord.ID, auditResult, relatorioFinal)
		if err != nil {
			log.Printf("⚠️ [PGRS] Erro ao salvar resultado IA: %v", err)
		} else {
			log.Printf("💾 [PGRS] Salvo com sucesso! ID: %d | Número: %s | Status: %s",
				pgrsRecord.ID, pgrsRecord.Numero, pgrsRecord.Status)
		}
	}

	// ═══════════════════════════════════════════════════════════
	// ETAPA 7: ENVIAR RESULTADO FINAL AO CLIENTE
	// ═══════════════════════════════════════════════════════════

	client.Conn <- events.Message{
		Type:    "socket:pgrs_report_ready",
		Payload: relatorioFinal,
	}

	log.Println("✅ [PGRS] Geração de PGRS concluída com sucesso!")
	log.Printf("📊 [PGRS] Resumo - Empresa: %s | Resíduos: %d | Score: %.1f%%", razaoSocial, len(inventory), auditScore)
}
