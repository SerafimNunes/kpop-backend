package library

import (
	"context"
	"fmt"
	"io" // Adicionado para suportar a leitura do stream de texto
	"log"
	"regexp"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/ledongthuc/pdf"
)

// LibrarianAgent é responsável pela gestão e extração de conhecimento de fontes técnicas
type LibrarianAgent struct {
	Svc *GeminiService
}

// NewLibrarian instancia um novo agente bibliotecário
func NewLibrarian(svc *GeminiService) *LibrarianAgent {
	return &LibrarianAgent{Svc: svc}
}

// ExtractTextFromPDF usa o ledongthuc/pdf para extração rápida e compatível
func (l *LibrarianAgent) ExtractTextFromPDF(filePath string) (string, error) {
	log.Printf("📄 [LIBRARIAN] Extraindo texto via ledongthuc/pdf: %s", filePath)

	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("erro ao abrir arquivo PDF: %v", err)
	}
	defer f.Close()

	// r.GetPlainText() retorna um Reader para o texto do PDF
	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("erro ao processar texto do PDF: %v", err)
	}

	// Correção: strings.Builder não tem ReadFrom.
	// Usamos io.ReadAll para ler todo o conteúdo do Reader fornecido pela biblioteca.
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("erro ao ler stream de texto do PDF: %v", err)
	}

	return string(content), nil
}

// applyDCTR (Document Context Token Reduction) - Mantendo sua lógica original
func (l *LibrarianAgent) applyDCTR(text string) string {
	// 1. Limpeza de metadados e URLs
	reLinks := regexp.MustCompile(`http[s]?://\S+|www\.\S+`)
	text = reLinks.ReplaceAllString(text, "")

	// 2. Corte de Bibliografia
	reRefs := regexp.MustCompile(`(?i)(Referências|Bibliográficas|References|Bibliography)`)
	loc := reRefs.FindStringIndex(text)
	if loc != nil {
		log.Println("✂️ [DCTR] Seção de referências localizada e removida.")
		text = text[:loc[0]]
	}

	// 3. Normalização de espaços
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// ProcessKnowledge executa a análise profunda
func (l *LibrarianAgent) ProcessKnowledge(ctx context.Context, rawContent string, source string) (string, error) {
	cleanedContent := l.applyDCTR(rawContent)

	log.Printf("🧠 [LIBRARIAN] Conteúdo reduzido de %d para %d caracteres.", len(rawContent), len(cleanedContent))

	systemInstruction := `### AGENTE BIBLIOTECÁRIO AUREN
Você é um Auditor Ambiental. Analise o documento técnico fornecido.
FOCO: Extrair artigos, leis, penalidades e obrigações operacionais.
REJEITE: Textos puramente teóricos ou opiniões de autores sem base legal clara.`

	fullPrompt := fmt.Sprintf("%s\n\n--- FONTE: %s ---\n\n--- CONTEÚDO ---\n%s\n\n--- TAREFA ---\nCrie um Sumário Executivo com: 1. Obrigações, 2. Riscos e 3. Próximos Passos.",
		systemInstruction, source, cleanedContent)

	return l.Svc.Generate(ctx, []genai.Part{genai.Text(fullPrompt)}, systemInstruction)
}
