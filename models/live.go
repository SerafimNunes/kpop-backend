package models

import (
	"time"

	"gorm.io/gorm"
)

type LiveArchive struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title     string `json:"title"`
	IdolName  string `json:"idol_name"`
	Platform  string `json:"platform"`   // Ex: Weverse, YouTube
	VideoPath string `json:"video_path"` // Caminho do arquivo para o FFmpeg
}

type CaptionLog struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	LiveArchiveID uint   `json:"live_archive_id"`
	Timestamp     int64  `json:"timestamp"` // Milissegundos desde o início da live
	Text          string `json:"text"`      // Tradução da IA
	IsVipOnly     bool   `gorm:"default:false" json:"is_vip_only"`
}
