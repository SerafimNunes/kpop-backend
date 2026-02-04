package library

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"auren-platform/internal/infrastructure/gemini"

	"cloud.google.com/go/vertexai/genai"
	"github.com/ledongthuc/pdf"
)

// Librarian é o Agente de Pesquisa/RAG (Público para ser visto pelo Core)
type Librarian struct {
	Gemini *gemini.Service
}

// NewLibrarian instancia o agente usando a infraestrutura de IA universal
func NewLibrarian(svc *gemini.Service) *Librarian {
	return &Librarian{Gemini: svc}
}

// Search implementa a interface de busca rápida exigida pelos Handlers
func (l *Librarian) Search(ctx context.Context, query string) (interface{}, error) {
	log.Printf("🔍 [LIBRARIAN] Executando busca semântica: %s", query)

	// Utiliza o Gemini para processar a busca no contexto de conhecimento
	prompt := fmt.Sprintf("Como especialista em conformidade, responda: %s", query)
	res, err := l.Gemini.Generate(ctx, []genai.Part{genai.Text(prompt)}, "Busca Técnica Auren")
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ExtractTextFromPDF extração de texto para alimentar a base de conhecimento
func (l *Librarian) ExtractTextFromPDF(filePath string) (string, error) {
	log.Printf("📄 [LIBRARIAN] Extraindo texto do PDF: %s", filePath)

	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo PDF: %v", err)
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("erro ao processar texto do PDF: %v", err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("erro ao ler stream de texto do PDF: %v", err)
	}

	return string(content), nil
}

// applyDCTR (Document Context Token Reduction) - Redução inteligente de tokens
func (l *Librarian) applyDCTR(text string) string {
	// 1. Limpeza de metadados e URLs
	reLinks := regexp.MustCompile(`http[s]?://\S+|www\.\S+`)
	text = reLinks.ReplaceAllString(text, "")

	// 2. Corte de Bibliografia
	reRefs := regexp.MustCompile(`(?i)(Referências|Bibliográficas|References|Bibliography)`)
	loc := reRefs.FindStringIndex(text)
	if loc != nil {
		log.Println("✂️ [DCTR] Seção de referências removida.")
		text = text[:loc[0]]
	}

	// 3. Normalização de espaços
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// ProcessKnowledge executa a análise profunda de documentos para o storage/knowledge
func (l *Librarian) ProcessKnowledge(ctx context.Context, rawContent string, source string) (string, error) {
	cleanedContent := l.applyDCTR(rawContent)

	log.Printf("🧠 [LIBRARIAN] Conteúdo otimizado: %d para %d caracteres.", len(rawContent), len(cleanedContent))

	systemInstruction := `### AGENTE BIBLIOTECÁRIO AUREN
Você é um Auditor Ambiental. Analise o documento técnico fornecido.
FOCO: Extrair artigos, leis, penalidades e obrigações operacionais.
REJEITE: Textos puramente teóricos ou opiniões de autores sem base legal clara.`

	fullPrompt := fmt.Sprintf("%s\n\n--- FONTE: %s ---\n\n--- CONTEÚDO ---\n%s\n\n--- TAREFA ---\nCrie um Sumário Executivo com: 1. Obrigações, 2. Riscos e 3. Próximos Passos.",
		systemInstruction, source, cleanedContent)

	return l.Gemini.Generate(ctx, []genai.Part{genai.Text(fullPrompt)}, systemInstruction)
}
