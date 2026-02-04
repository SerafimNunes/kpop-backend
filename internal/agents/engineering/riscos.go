package engineering

import (
	"auren-platform/internal/infrastructure/gemini"
	"cloud.google.com/go/vertexai/genai"
	"context"
	"fmt"
)

// AnalistaRiscos foca em segurança operacional e emergências ambientais
type AnalistaRiscos struct {
	gemini *gemini.Service
}

// NewAnalistaRiscos cria uma nova instância do analista de riscos
func NewAnalistaRiscos(s *gemini.Service) *AnalistaRiscos {
	return &AnalistaRiscos{gemini: s}
}

// SimularCenario cria modelos de impacto baseados em dados de instalações
func (a *AnalistaRiscos) SimularCenario(ctx context.Context, dadosInstalacao string) (string, error) {
	if dadosInstalacao == "" {
		return "Dados da instalação insuficientes para simulação.", nil
	}

	systemPrompt := `### ANALISTA DE RISCOS AMBIENTAIS AUREN
Você é especialista em Gestão de Riscos e Atendimento a Emergências.
Sua tarefa é:
1. Simular cenários de acidentes (vazamentos, incêndios, explosões) baseados nos dados da instalação.
2. Identificar receptores vulneráveis no entorno.
3. Sugerir medidas mitigadoras e preventivas para o PAE (Plano de Ação de Emergência).`

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("DADOS DA INSTALAÇÃO E ATIVIDADES:\n%s", dadosInstalacao)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}
