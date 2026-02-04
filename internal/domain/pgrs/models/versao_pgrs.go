package models

// VersaoPGRS keeps history of versions
// (simplified - fields copied from existing db schema)

// Note: base alias is declared in base.go

type VersaoPGRS struct {
	Base
	PGRSID          uint   `gorm:"not null" json:"pgrs_id"`
	NumeroVersao    int    `gorm:"not null" json:"numero_versao"`
	DadosFormulario string `gorm:"type:jsonb" json:"dados_formulario"`
	RelatorioGerado string `gorm:"type:text" json:"relatorio_gerado"`
	Observacoes     string `gorm:"type:text" json:"observacoes,omitempty"`
}
