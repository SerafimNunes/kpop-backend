package models

// LogPGRS registra ações no documento
// Note: base alias is declared in base.go

type LogPGRS struct {
	Base
	PGRSID    uint   `gorm:"not null" json:"pgrs_id"`
	Acao      string `gorm:"not null" json:"acao"`
	Descricao string `gorm:"type:text" json:"descricao"`
}
