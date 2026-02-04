package admin

import (
	"auren-platform/internal/infrastructure/gemini"
	"context"
	"fmt"
)

// HandleAdminAction gerencia a execução de tarefas da equipe administrativa (Manifesto 1.4)
func HandleAdminAction(ctx context.Context, s *gemini.Service, action string, payload map[string]interface{}) (interface{}, error) {
	// Instanciando a equipe de especialistas administrativos
	taxAdvisor := NewTaxAdvisor(s)
	hunter := NewHunter(s, nil)
	compliance := NewCompliance(s)
	controller := NewController(s)
	cs := NewCustomerSuccess(s)
	propostas := NewPropostas(s)

	// Extração segura do payload de dados
	data, _ := payload["data"].(string)

	switch action {
	case "validar_impostos":
		res, err := taxAdvisor.ValidarImpostos(ctx, data)
		return map[string]string{"response": res}, err

	case "gerar_pitch":
		res, err := hunter.GerarPitch(ctx, data)
		return map[string]string{"response": res}, err

	case "analisar_contrato":
		res, err := compliance.AnalisarContrato(ctx, data) // Sincronizado com o método do compliance.go
		return map[string]string{"response": res}, err

	case "analisar_fluxo":
		res, err := controller.AnalisarFluxo(ctx, data)
		return map[string]string{"response": res}, err

	case "sugerir_upsell":
		res, err := cs.SugerirUpsell(ctx, data)
		return map[string]string{"response": res}, err

	case "gerar_proposta":
		// Integração com propostas.go para fluxo comercial
		cliente, _ := payload["cliente"].(string)
		servico, _ := payload["servico"].(string)
		res, err := propostas.GerarProposta(ctx, cliente, servico, data)
		return map[string]string{"response": res}, err

	default:
		return nil, fmt.Errorf("ação administrativa '%s' não reconhecida pelo sistema", action)
	}
}
