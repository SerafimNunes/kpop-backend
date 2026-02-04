package finance

import (
	"time"

	"gorm.io/gorm"
)

// Transaction representa uma entrada ou saída no fluxo de caixa.
type Transaction struct {
	gorm.Model
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Status      string    `json:"status"` // ex: "pago", "pendente"
	Type        string    `json:"type"`   // "entrada" ou "saida"
	Amount      float64   `json:"amount"`
}
