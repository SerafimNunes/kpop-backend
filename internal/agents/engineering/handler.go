package engineering

import (
	"context"
	"fmt"
	"auren-platform/internal/engine"
)

// HandleEngineeringAction gerencia TODA a equipe técnica
func HandleEngineeringAction(ctx context.Context, s *engine.GeminiService, action string, payload map[string]interface{}) (interface{}, error) {
	// Instanciando a equipe técnica
	auditor := NewAuditorCampo(s)
	redator := NewRedatorPGRS(s)
	especialistaPNRS := NewEspecialistaPNRS(s)
	monitorCond := NewMonitorCondicionantes(s)
	analistaRiscos := NewAnalistaRiscos(s)

	data, _ := payload["data"].(string)

	switch action {
	case "auditar_campo":
		res, err := auditor.AnalisarEvidencias(ctx, data)
		return map[string]string{"response": res}, err

	case "redigir_pgrs":
		res, err := redator.ElaborarMinuta(ctx, data)
		return map[string]string{"response": res}, err

	case "classificar_residuo": // Novo (Agora usando o agente dedicado)
		res, err := especialistaPNRS.Classificar(ctx, data)
		return map[string]string{"response": res}, err
	
	case "extrair_condicionantes": // Novo
		res, err := monitorCond.ExtrairObrigacoes(ctx, data)
		return map[string]string{"response": res}, err

	case "simular_riscos": // Novo
		res, err := analistaRiscos.SimularCenario(ctx, data)
		return map[string]string{"response": res}, err

	default:
		return nil, fmt.Errorf("ação de engenharia '%s' não suportada", action)
	}
}