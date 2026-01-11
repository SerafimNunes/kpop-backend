package ai

import (
	"context"
	"fmt"
)

// LibrarianAgent é responsável pela gestão e extração de conhecimento de fontes técnicas
type LibrarianAgent struct {
	Svc *GeminiService
}

// NewLibrarian instancia um novo agente bibliotecário
func NewLibrarian(svc *GeminiService) *LibrarianAgent {
	return &LibrarianAgent{Svc: svc}
}

// ProcessKnowledge executa a análise profunda de conteúdo técnico e legislações
func (l *LibrarianAgent) ProcessKnowledge(ctx context.Context, content string, source string) (string, error) {
	// Definição de Instrução de Sistema Robusta (System Prompt)
	systemInstruction := `### INSTRUÇÃO DE SISTEMA: AGENTE BIBLIOTECÁRIO AUREN
Você é um Engenheiro Ambiental Sênior e Auditor de Conformidade.
Sua missão é realizar o 'Data Mining' de documentos, vídeos ou textos técnicos.
FOCO: Extrair obrigações legais, pontos críticos de fiscalização e oportunidades de otimização operacional.
RESTRIÇÃO: Responda de forma técnica, concisa e baseada apenas nos fatos apresentados.
IMPORTANTE: Se o conteúdo parecer um artigo de legislação, identifique os artigos e penalidades mencionadas.`

	fullPrompt := fmt.Sprintf("%s\n\n--- CONTEXTO DA FONTE ---\nFONTE: %s\nCONTEÚDO PARA ANÁLISE: %s\n\n--- TAREFA ---\nAnalise o conteúdo acima e retorne um sumário executivo com: 1. Pontos Críticos, 2. Base Legal Relacionada e 3. Sugestão de Ação Preventiva.", 
		systemInstruction, source, content)

	// Utiliza o serviço Gemini já configurado para processar o prompt estruturado
	return l.Svc.AnalyzeText(ctx, fullPrompt)
}