package models

import "github.com/jinzhu/gorm"

type TranslationPoll struct {
	gorm.Model
	VideoURL     string `gorm:"unique;not null"`
	Title        string
	TargetCoins  int    // Meta de K-Coins para destravar
	CurrentCoins int    `gorm:"default:0"`
	Status       string `gorm:"default:'OPEN'"` // OPEN, PROCESSING, COMPLETED
}
