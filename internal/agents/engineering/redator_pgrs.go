package engineering

import (
	"context"
	"fmt"
	"log"

	"auren-platform/internal/infrastructure/gemini"

	"cloud.google.com/go/vertexai/genai"
)

// RedatorPGRS renomeado para sincronizar com o Handler
type RedatorPGRS struct {
	gemini *gemini.Service
}

func NewRedatorPGRS(s *gemini.Service) *RedatorPGRS {
	return &RedatorPGRS{gemini: s}
}

// ElaborarMinuta (Método principal usado pelo Handler)
func (r *RedatorPGRS) ElaborarMinuta(ctx context.Context, textoBruto string) (string, error) {
	log.Println("✍️ [ENGINEERING] Elaborando minuta técnica...")

	systemPrompt := "Você é o Redator Técnico da Auren. Transforme notas em texto profissional ABNT."

	parts := []genai.Part{
		genai.Text("Conteúdo bruto: " + textoBruto),
	}

	return r.gemini.Generate(ctx, parts, systemPrompt)
}

// GerarRelatorioPGRS cria o laudo final em Markdown
func (r *RedatorPGRS) GerarRelatorioPGRS(ctx context.Context, pgrsDataJSON []byte, auditResult string) (string, error) {
	systemPrompt := `Você é o REDATOR-CHEFE TÉCNICO da Auren Consultoria, especialista em documentação ambiental.

TAREFA: Gerar um MEMORIAL DESCRITIVO profissional de um PGRS (Plano de Gerenciamento de Resíduos Sólidos) em formato Markdown.

ESTRUTURA OBRIGATÓRIA:

# MEMORIAL DESCRITIVO - PGRS
**Plano de Gerenciamento de Resíduos Sólidos**

---

## 1. IDENTIFICAÇÃO DO EMPREENDIMENTO

[Dados da empresa: razão social, CNPJ, CNAE, endereço]

## 2. OBJETIVO

Este Plano de Gerenciamento de Resíduos Sólidos (PGRS) tem como objetivo estabelecer as diretrizes e procedimentos para o gerenciamento adequado dos resíduos sólidos gerados pela empresa, em conformidade com a Política Nacional de Resíduos Sólidos (Lei 12.305/2010) e demais normas aplicáveis.

## 3. DIAGNÓSTICO DE RESÍDUOS

### 3.1 Inventário de Resíduos

[Tabela em Markdown com: Nome do Resíduo | Classe NBR 10.004 | Quantidade Mensal | Destinação]

### 3.2 Classificação Técnica

[Explicação da classificação de cada resíduo segundo NBR 10.004]

## 4. SEGREGAÇÃO E ACONDICIONAMENTO

[Descrever como os resíduos serão segregados na fonte e acondicionados]

## 5. ARMAZENAMENTO TEMPORÁRIO

[Descrever área de armazenamento, capacidade, características físicas]

## 6. TRANSPORTE INTERNO E EXTERNO

[Descrever procedimentos de transporte, transportadoras autorizadas]

## 7. DESTINAÇÃO FINAL

[Para cada tipo de resíduo, descrever a destinação final ambientalmente adequada]

## 8. METAS E PROCEDIMENTOS

### 8.1 Metas de Redução
[Estabelecer metas quantitativas de redução de geração]

### 8.2 Reciclagem e Reaproveitamento
[Percentuais e ações para maximizar reciclagem]

## 9. RESPONSABILIDADE TÉCNICA

[Informar responsável técnico com CREA]

## 10. CRONOGRAMA DE IMPLEMENTAÇÃO

[Prazos para implementação das ações]

---

**Data de Elaboração:** [Data atual]  
**Validade:** 12 meses a partir da data de elaboração

---

DIRETRIZES DE REDAÇÃO:
- Use linguagem técnica profissional
- Seja específico e detalhado
- Cite sempre as normas aplicáveis (Lei 12.305/2010, NBR 10.004, etc.)
- Formate tabelas em Markdown correto
- Use negrito para termos importantes
- Seja objetivo mas completo

NÃO:
- Não use termos genéricos demais
- Não deixe seções vazias
- Não invente dados que não foram fornecidos`

	prompt := fmt.Sprintf(`Dados do formulário PGRS:
%s

Resultado da Auditoria de Compliance:
%s

Gere o Memorial Descritivo completo seguindo rigorosamente a estrutura definida.`, string(pgrsDataJSON), auditResult)

	return r.gemini.Generate(ctx, []genai.Part{genai.Text(prompt)}, systemPrompt)
}
