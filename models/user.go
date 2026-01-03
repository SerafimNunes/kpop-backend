package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	GoogleID string `gorm:"uniqueIndex;not null" json:"google_id"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`

	// Nível de acesso para monetização
	IsVIP          bool   `gorm:"default:false" json:"is_vip"`
	SubscriptionID string `json:"subscription_id"` // ID do Stripe/MercadoPago
}
