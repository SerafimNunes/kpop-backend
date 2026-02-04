package admin

import (
	"context"
	"fmt"
	"log"

	"auren-platform/internal/infrastructure/gemini" // Importação correta da Infra
	"cloud.google.com/go/vertexai/genai"
)

// Propostas gerencia a redação de propostas comerciais e contratos jurídicos
type Propostas struct {
	gemini *gemini.Service
}

// NewPropostas instancia o agente comercial (Sincronizado com o Handler)
func NewPropostas(s *gemini.Service) *Propostas {
	return &Propostas{gemini: s}
}

// GerarProposta rascunha uma proposta orçamentária personalizada
func (p *Propostas) GerarProposta(ctx context.Context, clienteNome string, servico string, detalhes string) (string, error) {
	log.Printf("💰 [ADMIN] Gerando proposta para: %s - Serviço: %s", clienteNome, servico)

	systemInstruction := `### DIRETRIZ COMERCIAL AUREN PLATFORM
Você é o Diretor Comercial da Auren Platform. Sua missão é converter leads em clientes através de propostas técnicas impecáveis e persuasivas.

ESTRUTURA OBRIGATÓRIA:
1. CABEÇALHO, 2. OBJETIVO, 3. ESCOPO TÉCNICO, 4. METODOLOGIA (IA Auren), 5. INVESTIMENTO, 6. PRAZOS.`

	prompt := fmt.Sprintf(`Gere uma proposta comercial para:
Cliente: %s
Serviço: %s
Contexto: %s`, clienteNome, servico, detalhes)

	return p.gemini.Generate(ctx, []genai.Part{genai.Text(prompt)}, systemInstruction)
}

// AnalisarContrato (Mapeado no Handler como ação de Compliance ou Propostas)
func (p *Propostas) AnalisarContrato(ctx context.Context, dadosProposta string) (string, error) {
	log.Println("📜 [ADMIN] Redigindo contrato jurídico formal...")

	systemInstruction := `### DIRETRIZ JURÍDICA AUREN PLATFORM
Você é um Advogado Especialista em Direito Ambiental. Transforme o resumo comercial em um instrumento jurídico formal.`

	prompt := fmt.Sprintf("Redija o contrato baseado nestes dados: %s", dadosProposta)

	return p.gemini.Generate(ctx, []genai.Part{genai.Text(prompt)}, systemInstruction)
}
