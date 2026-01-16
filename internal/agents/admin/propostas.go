package admin

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/vertexai/genai"
)

// BusinessAgent gerencia a redação de propostas comerciais e contratos jurídicos
type BusinessAgent struct {
	Svc *GeminiService
}

// NewBusinessAgent instancia o agente comercial
func NewBusinessAgent(svc *GeminiService) *BusinessAgent {
	return &BusinessAgent{Svc: svc}
}

// GenerateProposal rascunha uma proposta orçamentária personalizada
func (b *BusinessAgent) GenerateProposal(ctx context.Context, clienteNome string, servico string, detalhes string) (string, error) {
	log.Printf("💰 [BUSINESS] Gerando proposta para: %s - Serviço: %s", clienteNome, servico)

	systemInstruction := `### DIRETRIZ COMERCIAL AUREN PLATFORM
Você é o Diretor Comercial da Auren Platform. Sua missão é converter leads em clientes através de propostas técnicas impecáveis e persuasivas.

ESTRUTURA OBRIGATÓRIA DA RESPOSTA:
1. CABEÇALHO: Identificação Auren e Cliente.
2. OBJETIVO: Justificativa ambiental do serviço.
3. ESCOPO TÉCNICO: Lista detalhada de atividades e entregáveis.
4. METODOLOGIA: Como a Auren aplicará a tecnologia e normas CONAMA/PNRS.
5. INVESTIMENTO: Tabela de valores (simulados) e condições de pagamento.
6. PRAZOS: Cronograma por fases.

TOM DE VOZ: Executivo, focado em conformidade ESG e segurança jurídica.`

	prompt := fmt.Sprintf(`Gere uma proposta comercial de consultoria ambiental para:
Cliente: %s
Serviço Principal: %s
Contexto do Projeto: %s

Destaque o uso da inteligência artificial Auren para monitoramento de conformidade.`, clienteNome, servico, detalhes)

	return b.Svc.Generate(ctx, []genai.Part{genai.Text(prompt)}, systemInstruction)
}

// GenerateContract redige o contrato formal baseado em uma proposta aceita
func (b *BusinessAgent) GenerateContract(ctx context.Context, dadosProposta string) (string, error) {
	log.Println("📜 [BUSINESS] Redigindo contrato jurídico formal...")

	// Template Jurídico Base inserido nas instruções do sistema para manter o padrão Auren
	templateBase := `
	CLÁUSULA PRIMEIRA - DO OBJETO: Prestação de serviços de consultoria técnica ambiental.
	CLÁUSULA SEGUNDA - DA RESPONSABILIDADE TÉCNICA: A CONTRATADA emitirá a devida ART (Anotação de Responsabilidade Técnica).
	CLÁUSULA TERCEIRA - DAS OBRIGAÇÕES: Cumprimento rigoroso da legislação ambiental vigente e normas da ABNT.
	CLÁUSULA QUARTA - CONFIDENCIALIDADE E LGPD: Tratamento rigoroso de dados sensíveis e sigilo industrial.
	CLÁUSULA QUINTA - DOS HONORÁRIOS: Conforme cronograma financeiro aceito na proposta.
	CLÁUSULA SEXTA - RESCISÃO: Condições de distrato e multas contratuais.
	CLÁUSULA SÉTIMA - FORO: Comarca de Curitiba/PR (ou foro do cliente).`

	systemInstruction := fmt.Sprintf(`### DIRETRIZ JURÍDICA AUREN PLATFORM
Você é um Advogado Especialista em Direito Ambiental. Transforme o resumo comercial em um instrumento jurídico formal.

Utilize esta estrutura base para as cláusulas:
%s

REGRAS:
- Use linguagem jurídica (juridiquês moderado).
- Garanta que todas as obrigações citadas na proposta estejam refletidas no objeto.
- Formate o documento com títulos em negrito.`, templateBase)

	prompt := fmt.Sprintf("Redija o contrato de prestação de serviços completo baseado nestes dados de proposta: %s", dadosProposta)

	return b.Svc.Generate(ctx, []genai.Part{genai.Text(prompt)}, systemInstruction)
}
