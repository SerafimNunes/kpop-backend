package user

import (
	"time"

	"gorm.io/gorm"
)

// Role representa os papéis definidos no item 3.1 da Fonte da Verdade
type Role struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"` // ADMIN, GERENTE, RT, CLIENTE
}

// User centraliza os dados de perfil e autenticação oficial
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"not null" json:"name"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"` // Omitido no JSON por segurança
	RoleID       uint           `gorm:"not null" json:"role_id"`
	Role         Role           `gorm:"foreignKey:RoleID" json:"role"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
