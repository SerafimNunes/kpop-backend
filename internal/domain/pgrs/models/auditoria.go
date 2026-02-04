package models

// AuditoriaExecucao placeholder
type AuditoriaExecucao struct {
	Base
	UnidadeID uint `json:"unidade_id"`
	Status    string
}

// PontoCritico placeholder
type PontoCritico struct {
	Base
	AuditoriaID   uint   `json:"auditoria_id"`
	Identificador string `json:"identificador"`
	Titulo        string `json:"title"`
	Descricao     string `json:"desc"`
	Status        string `json:"status"`
	Color         string `json:"color"`
}
