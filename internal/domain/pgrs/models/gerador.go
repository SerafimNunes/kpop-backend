package models

import (
	comercialmodels "auren-platform/internal/domain/comercial/models"
)

// Gerador representa a empresa cliente

type Gerador struct {
	Base
	RazaoSocial      string                     `gorm:"not null" json:"razao_social"`
	CNPJ             string                     `gorm:"uniqueIndex;not null" json:"cnpj"`
	CNAE             string                     `json:"cnae"`
	Endereco         string                     `json:"endereco"`
	Municipio        string                     `json:"municipio"`
	Estado           string                     `json:"estado"`
	ResponsavelLegal string                     `json:"responsavel_legal"`
	Unidades         []Unidade                  `gorm:"foreignKey:GeradorID" json:"unidades"`
	Propostas        []comercialmodels.Proposta `gorm:"foreignKey:GeradorID" json:"propostas"`
}
