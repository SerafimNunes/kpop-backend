package media_analysis

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Expressão regular para remover múltiplos espaços em branco e quebras de linha
var whitespaceRegex = regexp.MustCompile(`\s{2,}`)

// FetchWebText busca o conteúdo de uma URL e extrai apenas o texto útil.
// Isso é crucial para remover HTML, scripts e menus, reduzindo o número
// de tokens enviados para a IA.
func FetchWebText(ctx context.Context, url string) (string, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	// Simula um navegador para evitar bloqueios simples
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	extractText(doc, &sb)

	// Limpa e normaliza o texto extraído
	text := whitespaceRegex.ReplaceAllString(sb.String(), "\n")
	return strings.TrimSpace(text), nil
}

// extractText percorre a árvore HTML e extrai o texto de tags relevantes.
func extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.ElementNode {
		// Ignora tags que tipicamente não contêm conteúdo principal
		if n.Data == "script" || n.Data == "style" || n.Data == "nav" || n.Data == "footer" || n.Data == "aside" || n.Data == "header" {
			return
		}
	}

	if n.Type == html.TextNode {
		// Adiciona o texto ao builder
		sb.WriteString(n.Data)
		sb.WriteString(" ") // Adiciona um espaço para separar palavras entre tags
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, sb)
	}
}
