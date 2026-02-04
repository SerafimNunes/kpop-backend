package models

import "time"

// Vistoria placeholder
type Vistoria struct {
	Base
	UnidadeID uint      `json:"unidade_id"`
	Data      time.Time `json:"data"`
	Tecnico   string    `json:"tecnico"`
}
