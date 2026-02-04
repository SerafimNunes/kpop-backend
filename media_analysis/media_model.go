package media_analysis

import (
	"time"

	"gorm.io/gorm"
)

// MediaSource centraliza o armazenamento de ativos de mídia da plataforma
type MediaSource struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	SourceURL   string         `gorm:"uniqueIndex;not null" json:"source_url"`
	Type        string         `gorm:"default:'VIDEO'" json:"type"` // VIDEO, AUDIO, IMAGE
	IsProcessed bool           `gorm:"default:false" json:"is_processed"`
	CompanyID   uint           `gorm:"index" json:"company_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relacionamento com as análises da IA (Cada Insight é um token salvo)
	Insights []MediaInsight `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE" json:"insights"`
}

// MediaInsight representa a destilação técnica da mídia feita pelos Agentes
type MediaInsight struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	MediaID       uint      `gorm:"index" json:"media_id"`
	Timestamp     int64     `json:"timestamp"` // Em milissegundos para precisão no frame
	Summary       string    `gorm:"type:text" json:"summary"`
	Risks         string    `gorm:"type:text" json:"risks"`         // Risco Ambiental/Ocupacional
	Applicability string    `gorm:"type:text" json:"applicability"` // Ex: NBR 10004, NR 35
	ApprovedByRT  bool      `gorm:"default:false" json:"approved_by_rt"`
	CreatedAt     time.Time `json:"created_at"`
}
