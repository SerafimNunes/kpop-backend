package models

import "gorm.io/gorm"

// LiveArchive representa uma live salva para replay futuro
type LiveArchive struct {
	gorm.Model
	Title       string       `json:"title"`
	Platform    string       `json:"platform"` // YouTube, Weverse, etc.
	OriginalURL string       `json:"original_url"`
	Captions    []CaptionLog `gorm:"foreignKey:LiveArchiveID"`
}

// CaptionLog guarda cada frase traduzida vinculada a um tempo da live
type CaptionLog struct {
	gorm.Model
	LiveArchiveID uint   `json:"live_archive_id"`
	Timestamp     int64  `json:"timestamp"` // Tempo em milissegundos
	OriginalText  string `json:"original_text"`
	RefinedText   string `json:"refined_text"`
}