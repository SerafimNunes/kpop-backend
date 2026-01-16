package admin

import (
	"context"
	"fmt"
	"auren-platform/internal/engine"
)

// HandleAdminAction gerencia TODA a equipe administrativa
func HandleAdminAction(ctx context.Context, s *engine.GeminiService, action string, payload map[string]interface{}) (interface{}, error) {
	// Instanciando a equipe completa
	taxAdvisor := NewTaxAdvisor(s)
	hunter := NewHunter(s)
	compliance := NewCompliance(s)
	controller := NewController(s)
	cs := NewCustomerSuccess(s)
	// propostas.go é chamado sob demanda se necessário, ou integrado aqui

	data, _ := payload["data"].(string)

	switch action {
	case "validar_impostos":
		res, err := taxAdvisor.ValidarImpostos(ctx, data)
		return map[string]string{"response": res}, err

	case "gerar_pitch":
		res, err := hunter.GerarPitch(ctx, data)
		return map[string]string{"response": res}, err
	
	case "analisar_contrato": // Novo
		res, err := compliance.AnalisarRisco(ctx, data)
		return map[string]string{"response": res}, err

	case "analisar_fluxo": // Novo
		res, err := controller.AnalisarFluxo(ctx, data)
		return map[string]string{"response": res}, err

	case "sugerir_upsell": // Novo
		res, err := cs.SugerirUpsell(ctx, data)
		return map[string]string{"response": res}, err

	default:
		return nil, fmt.Errorf("ação administrativa '%s' não suportada", action)
	}
}