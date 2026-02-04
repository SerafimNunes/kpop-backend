package models

// InventarioResiduo placeholder (fields simplified)
type InventarioResiduo struct {
	Base
	UnidadeID uint   `json:"unidade_id"`
	Nome      string `json:"nome"`
	Classe    string `json:"classe"`
}
