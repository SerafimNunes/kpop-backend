package engineering

import (
	"context"
	"fmt"

	"auren-platform/internal/infrastructure/gemini"

	"cloud.google.com/go/vertexai/genai"
)

// AuditorCampo renomeado para sincronizar com o Handler
type AuditorCampo struct {
	gemini *gemini.Service
}

func NewAuditorCampo(s *gemini.Service) *AuditorCampo {
	return &AuditorCampo{gemini: s}
}

// AnalisarEvidencias processa fotos/vídeos e notas (Método principal usado pelo Handler)
func (a *AuditorCampo) AnalisarEvidencias(ctx context.Context, data string) (string, error) {
	systemPrompt := "VOCÊ É O AUDITOR DE CAMPO DA AUREN. Identifique inconformidades visuais NBR 12.235 e NBR 7.500."

	parts := []genai.Part{
		genai.Text(fmt.Sprintf("Dados para auditoria: %s", data)),
	}

	return a.gemini.Generate(ctx, parts, systemPrompt)
}

// AnalisarPGRSCompliance verifica os dados de um formulário contra as normas
func (a *AuditorCampo) AnalisarPGRSCompliance(ctx context.Context, pgrsDataJSON []byte) (string, error) {
	systemPrompt := `Você é o AUDITOR-CHEFE DE COMPLIANCE AMBIENTAL da Auren Consultoria.

TAREFA: Analisar dados de PGRS e identificar conformidades e não-conformidades com:
- Lei 12.305/2010 (Política Nacional de Resíduos Sólidos)
- NBR 10.004:2004 (Classificação de Resíduos Sólidos)
- NBR 12.235 (Armazenamento de Resíduos)
- Resoluções CONAMA relevantes

ANÁLISE OBRIGATÓRIA:
1. Verificar se todos os resíduos estão classificados corretamente
2. Verificar se as destinações são adequadas para cada classe
3. Identificar riscos de passivos ambientais
4. Avaliar exposição jurídica da empresa

FORMATO DE SAÍDA OBRIGATÓRIO (Markdown):

## 📊 SCORE DE COMPLIANCE
[Percentual de 0-100%]

## ✅ CONFORMIDADES IDENTIFICADAS
- [Lista de pontos positivos]

## ⚠️ NÃO-CONFORMIDADES CRÍTICAS
- [Lista de problemas graves que precisam correção imediata]

## 📋 RECOMENDAÇÕES PRIORITÁRIAS
1. [Ação prioritária 1]
2. [Ação prioritária 2]
3. [Ação prioritária 3]

## ⚖️ EXPOSIÇÃO JURÍDICA
[MÍNIMA / MODERADA / ALTA] - [Justificativa breve]

## 📚 BASE LEGAL
- [Artigos e normas aplicáveis]

Seja técnico, objetivo e sempre cite a base legal.`

	prompt := fmt.Sprintf("Dados do PGRS para análise:\n\n%s\n\nExecute a análise de compliance completa.", string(pgrsDataJSON))

	return a.gemini.Generate(ctx, []genai.Part{genai.Text(prompt)}, systemPrompt)
}
