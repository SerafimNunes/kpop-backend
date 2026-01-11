package media_analysis

import (
	"time"

	"gorm.io/gorm"
)

// MediaSource representa um vídeo técnico para análise (Vistoria/Treinamento)
type MediaSource struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	SourceURL   string         `gorm:"uniqueIndex;not null" json:"source_url"`
	Type        string         `gorm:"default:'VIDEO'" json:"type"`
	IsProcessed bool           `gorm:"default:false" json:"is_processed"`
	CompanyID   uint           `gorm:"index" json:"company_id"` // Vinculado a uma empresa
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Insights []MediaInsight `gorm:"foreignKey:MediaID;constraint:OnDelete:CASCADE" json:"insights"`
}

// MediaInsight armazena as análises técnicas geradas pela IA e validadas pela RT
type MediaInsight struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	MediaID       uint      `gorm:"index" json:"media_id"`
	Timestamp     int64     `json:"timestamp"` // Momento exato do vídeo
	Summary       string    `gorm:"type:text" json:"summary"`
	Risks         string    `gorm:"type:text" json:"risks"`         // Riscos identificados
	Applicability string    `gorm:"type:text" json:"applicability"` // Normas aplicáveis
	ApprovedByRT  bool      `gorm:"default:false" json:"approved_by_rt"`
	CreatedAt     time.Time `json:"created_at"`
}
