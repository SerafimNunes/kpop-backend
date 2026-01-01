package models

import "github.com/jinzhu/gorm"

type LiveArchive struct {
	gorm.Model
	GroupName    string `gorm:"index"`
	IdolName     string `gorm:"index"`
	Title        string
	ThumbnailURL string
	OriginalURL  string // Link para Weverse/YouTube/X
	Platform     string // YouTube, Weverse, etc
	IsLive       bool   `gorm:"default:false"`
	SponsorID    *uint  // ID do usuário que pagou pela tradução (se houver)
	Sponsor      User   `gorm:"foreignkey:SponsorID"`
	CaptionLogs  []CaptionLog
}

type CaptionLog struct {
	gorm.Model
	LiveArchiveID uint   `gorm:"index"`
	Timestamp     int64  // Tempo em milissegundos desde o início da live
	Text          string `gorm:"type:text"` // Texto traduzido para PT-BR
}
