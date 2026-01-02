package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique" json:"username"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"-"`    // O "-" impede que a senha vaze no JSON
	Role     string `json:"role"` // "admin", "producer", "fan"
}
