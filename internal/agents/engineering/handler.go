package engineering

import (
	"auren-platform/internal/infrastructure/gemini"
	"context"
	"fmt"
)

// HandleEngineeringAction coordena a execução das tarefas da equipe de engenharia ambiental
func HandleEngineeringAction(ctx context.Context, s *gemini.Service, action string, payload map[string]interface{}) (interface{}, error) {
	// Instanciando a equipe técnica especializada
	// Cada agente recebe a referência ao serviço de infraestrutura de IA
	auditor := NewAuditorCampo(s)
	redator := NewRedatorPGRS(s)
	especialistaPNRS := NewEspecialistaPNRS(s)
	monitorCond := NewMonitorCondicionantes(s)
	analistaRiscos := NewAnalistaRiscos(s)

	// Extração segura do conteúdo de dados do payload
	data, _ := payload["data"].(string)

	switch action {
	case "auditar_campo":
		// Foco em análise de evidências fotográficas e notas de campo
		res, err := auditor.AnalisarEvidencias(ctx, data)
		return map[string]string{"response": res}, err

	case "redigir_pgrs":
		// Elaboração de minutas para Planos de Gerenciamento de Resíduos Sólidos
		res, err := redator.ElaborarMinuta(ctx, data)
		return map[string]string{"response": res}, err

	case "classificar_residuo":
		// Classificação baseada na NBR 10.004 e PNRS
		res, err := especialistaPNRS.Classificar(ctx, data)
		return map[string]string{"response": res}, err

	case "extrair_condicionantes":
		// Extração de obrigações de licenças ambientais (LP, LI, LO)
		res, err := monitorCond.ExtrairObrigacoes(ctx, data)
		return map[string]string{"response": res}, err

	case "simular_riscos":
		// Simulação de cenários de impacto e passivos ambientais
		res, err := analistaRiscos.SimularCenario(ctx, data)
		return map[string]string{"response": res}, err

	default:
		return nil, fmt.Errorf("ação de engenharia '%s' não suportada pela malha técnica", action)
	}
}
