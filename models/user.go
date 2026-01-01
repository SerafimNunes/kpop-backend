package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Name          string `gorm:"size:100;not null"`
	Email         string `gorm:"unique;not null"`
	Role          string `gorm:"default:'FAN'"` // FAN, MODERATOR, ADMIN
	KCoins        int    `gorm:"default:0"`
	Karma         int    `gorm:"default:0"` // Status social por ajudar a comunidade
	ProfilePic    string
	Subscriptions []LiveArchive `gorm:"many2many:user_subscriptions;"`
}
