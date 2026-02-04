package models

import (
	"time"

	shared "auren-platform/internal/shared/models"
)

// Alias local para shared.Base
type Base = shared.Base

// Proposta representa proposta comercial

type Proposta struct {
	Base
	GeradorID       uint    `json:"gerador_id"`
	ClienteNome     string  `json:"cliente_nome"`
	Titulo          string  `json:"titulo"`
	ServicoTipo     string  `json:"servico_tipo"`
	Status          string  `json:"status"` // RASCUNHO, ENVIADA, ACEITA
	ValorTotal      float64 `json:"valor_total"`
	Moeda           string  `gorm:"default:'BRL'" json:"moeda"`
	DescricaoEscopo string  `gorm:"type:text" json:"descricao_escopo"`
	Especificidades string  `gorm:"type:text" json:"especificidades"`
	SumarioIA       string  `gorm:"type:text" json:"sumario_ia"`
	ContratoGerado  bool    `gorm:"default:false" json:"contrato_gerado"`
}

type Contrato struct {
	Base
	PropostaID uint      `json:"proposta_id"`
	GeradorID  uint      `json:"gerador_id"`
	Numero     string    `gorm:"uniqueIndex" json:"numero"`
	Conteudo   string    `gorm:"type:text" json:"conteudo"`
	DataInicio time.Time `json:"data_inicio"`
	DataFim    time.Time `json:"data_fim"`
	Assinado   bool      `gorm:"default:false" json:"assinado"`
	HashDoc    string    `json:"hash_doc"`
}
