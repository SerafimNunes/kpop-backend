package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"auren-platform/ai"
	"auren-platform/db"
	"auren-platform/media_analysis"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ServeWS recebe o GeminiService unificado e processa roteamento por intenção.
func ServeWS(h *Hub, brain *ai.GeminiService, ap *media_analysis.AudioProcessor, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Erro ao fazer upgrade: %v", err)
		return
	}

	client := &Client{
		ID:   uuid.New().String()[:8],
		Conn: make(chan Message, 512),
	}

	h.Register <- client

	var currentAction string
	var currentFileName string

	defer func() {
		h.Unregister <- client
		conn.Close()
	}()

	// Goroutine de escrita
	go func() {
		for msg := range client.Conn {
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}()

	// Loop de leitura principal
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		// ROTEAMENTO DE BINÁRIOS (Imagens de Campo vs PDFs de Inteligência)
		if messageType == websocket.BinaryMessage {
			if currentAction == "upload_knowledge" {
				go handleKnowledgeUpload(ctx, client, brain, p, currentFileName)
			} else {
				// Se cair aqui sem action prévia, assume campo (Vistoria 1 como fallback ou erro)
				go handleMediaEvidence(ctx, client, brain, p, 1)
			}
			currentAction = "" // Reseta após processar o binário
			continue
		}

		// PROCESSAMENTO DE MENSAGENS TEXTO (JSON)
		var req map[string]interface{}
		if err := json.Unmarshal(p, &req); err != nil {
			cancel()
			continue
		}

		action, _ := req["action"].(string)

		switch action {
		case "upload_knowledge":
			// Apenas prepara o estado para o binário que virá em seguida
			currentAction = "upload_knowledge"
			if name, ok := req["fileName"].(string); ok {
				currentFileName = name
			}
			client.Conn <- Message{Type: "status", Payload: "Auren pronta para receber documento técnico..."}

		case "process_insight":
			url, _ := req["url"].(string)
			go handleExternalInsight(ctx, client, brain, ap, url)

		case "consolidate_pgrs":
			data, _ := req["pgrsData"]
			go handlePGRSConsolidation(ctx, client, brain, data)
		}
	}
}

// handleKnowledgeUpload processa PDFs/Artigos e salva na tabela ConhecimentoTecnico
func handleKnowledgeUpload(ctx context.Context, client *Client, brain *ai.GeminiService, data []byte, fileName string) {
	client.Conn <- Message{Type: "status", Payload: "Bibliotecário Auren indexando documento..."}

	// 1. Persistência Física
	storagePath := fmt.Sprintf("./storage/knowledge/%s_%d.pdf", fileName, time.Now().Unix())
	_ = os.MkdirAll(filepath.Dir(storagePath), 0755)
	_ = os.WriteFile(storagePath, data, 0644)

	// 2. Análise via IA (Librarian Agent)
	librarian := ai.NewLibrarian(brain)
	// Como o Gemini pode ler PDFs binários, passamos o conteúdo.
	// Caso use análise de texto, converteríamos o PDF aqui.
	res, err := librarian.ProcessKnowledge(ctx, "[Documento Binário: "+fileName+"]", fileName)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Erro na análise do bibliotecário."}
		return
	}

	// 3. Persistência no Banco (Tabela Dedicada)
	conhecimento := db.ConhecimentoTecnico{
		Titulo:    fileName,
		Fonte:     fileName,
		Tipo:      "PDF_TECNICO",
		SumarioIA: res,
		FileHash:  db.GenerateHash(data),
	}
	db.DB.Create(&conhecimento)

	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handleMediaEvidence(ctx context.Context, client *Client, brain *ai.GeminiService, data []byte, vistoriaID uint) {
	client.Conn <- Message{Type: "status", Payload: "Auditor analisando evidência visual..."}

	storagePath := fmt.Sprintf("./storage/evidences/VISTORIA_%d_%d.media", vistoriaID, time.Now().UnixNano())
	_ = os.MkdirAll(filepath.Dir(storagePath), 0755)
	_ = os.WriteFile(storagePath, data, 0644)

	evidencia := db.Evidencia{
		VistoriaID: vistoriaID,
		Tipo:       "FIELD_EVIDENCE",
		StorageURL: storagePath,
		FileHash:   db.GenerateHash(data),
		FileSize:   int64(len(data)),
	}
	db.DB.Create(&evidencia)

	res, err := brain.AnalyzeVisualEvidence(ctx, data, "Analise esta evidência de campo sob a ótica das NBRs ambientais.")
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Falha na auditoria de campo."}
		return
	}

	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handleExternalInsight(ctx context.Context, client *Client, brain *ai.GeminiService, ap *media_analysis.AudioProcessor, url string) {
	client.Conn <- Message{Type: "status", Payload: "Bibliotecário processando fonte externa..."}

	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		audioChan := make(chan []byte, 1)
		go ap.StartDigitalEar(ctx, url, audioChan)

		select {
		case chunk := <-audioChan:
			res, err := brain.AnalyzeTechnicalAudio(ctx, chunk, "Extraia o conhecimento técnico deste vídeo.")
			if err != nil {
				client.Conn <- Message{Type: "error", Payload: "Erro ao processar conhecimento externo."}
				return
			}
			client.Conn <- Message{Type: "technical_insight", Payload: res}
		case <-ctx.Done():
			client.Conn <- Message{Type: "error", Payload: "Timeout no processamento de vídeo."}
		}
		return
	}

	res, err := brain.ExecuteSimplePrompt(ctx, "Resuma tecnicamente os pontos de conformidade desta fonte: "+url)
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Falha ao analisar URL."}
		return
	}
	client.Conn <- Message{Type: "technical_insight", Payload: res}
}

func handlePGRSConsolidation(ctx context.Context, client *Client, brain *ai.GeminiService, formData interface{}) {
	client.Conn <- Message{Type: "status", Payload: "Redator consolidando laudo técnico..."}
	jsonBytes, _ := json.Marshal(formData)
	res, err := brain.RefinarParaLaudo(ctx, string(jsonBytes))
	if err != nil {
		client.Conn <- Message{Type: "error", Payload: "Falha na redação do documento."}
		return
	}
	client.Conn <- Message{Type: "technical_insight", Payload: res}
}
