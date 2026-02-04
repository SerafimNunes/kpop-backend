package models

import (
	"time"

	shared "auren-platform/internal/shared/models"
)

// Alias local para shared.Base
type Base = shared.Base

// LancamentoFinanceiro representa um lançamento financeiro

type LancamentoFinanceiro struct {
	Base
	ContratoID  uint      `json:"contrato_id,omitempty"`
	Description string    `gorm:"not null" json:"description"`
	Category    string    `json:"category"`
	Amount      float64   `gorm:"not null" json:"amount"`
	Date        time.Time `gorm:"not null" json:"date"`
	Status      string    `gorm:"not null" json:"status"` // PENDENTE, PAGO
	Type        string    `gorm:"not null" json:"type"`   // entrada, saida
}
