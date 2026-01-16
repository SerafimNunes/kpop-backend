package library

import (
	"context"
	"fmt"
)

// HandleLibraryAction gerencia a equipe de pesquisa
func HandleLibraryAction(ctx context.Context, lib *Librarian, action string, payload map[string]interface{}) (interface{}, error) {
	// Librarian já vem instanciado do Core, precisamos dos novos
	curador := NewCurador(lib.gemini) // Acesso ao gemini via Librarian (pode exigir ajuste de visibilidade ou passar Gemini direto no handler)
	radar := NewRadar(lib.gemini)

	query, _ := payload["query"].(string)

	switch action {
	case "consultar_biblioteca":
		res, err := lib.Search(ctx, query)
		return map[string]interface{}{"results": res}, err

	case "buscar_jurisprudencia": // Novo
		res, err := curador.BuscarJurisprudencia(ctx, query)
		return map[string]string{"response": res}, err

	case "radar_legislativo": // Novo (Agora com agente dedicado)
		res, err := radar.AnalisarDiario(ctx, query)
		return map[string]string{"response": res}, err

	default:
		return nil, fmt.Errorf("ação de biblioteca '%s' não suportada", action)
	}
}