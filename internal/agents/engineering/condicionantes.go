package engineering

import (
	"auren-platform/internal/infrastructure/gemini"
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
)

// MonitorCondicionantes monitora licenças ambientais e prazos legais
type MonitorCondicionantes struct {
	gemini *gemini.Service
}

// NewMonitorCondicionantes cria uma nova instância do agente de monitoramento
func NewMonitorCondicionantes(s *gemini.Service) *MonitorCondicionantes {
	return &MonitorCondicionantes{gemini: s}
}

// ExtrairObrigacoes analisa o texto de licenças (LP, LI, LO) e gera cronogramas
func (a *MonitorCondicionantes) ExtrairObrigacoes(ctx context.Context, licencaTexto string) (string, error) {
	if licencaTexto == "" {
		return "Nenhum texto de licença fornecido para análise.", nil
	}

	systemPrompt := `### AGENTE DE MONITORAMENTO DE CONDICIONANTES AUREN
Você é um especialista em licenciamento ambiental. Sua missão é:
1. Ler o texto da Licença Ambiental fornecida.
2. Extrair TODAS as condicionantes, obrigações e prazos.
3. Gerar um cronograma de alertas e ações necessárias para evitar infrações ou multas.`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("CONTEÚDO DA LICENÇA:\n%s", licencaTexto)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
