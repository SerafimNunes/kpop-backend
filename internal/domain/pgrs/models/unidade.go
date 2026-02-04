package models

// shared base is aliased in base.go

// Unidade é um estabelecimento do Gerador

type Unidade struct {
	Base
	GeradorID    uint   `gorm:"not null" json:"gerador_id"`
	Nome         string `gorm:"not null" json:"nome"`
	TipoOperacao string `json:"tipo_operacao"`

	ResponsavelTecnico   string `json:"responsavel_tecnico"`
	RegistroProfissional string `json:"registro_profissional"`

	Inventarios []InventarioResiduo  `gorm:"foreignKey:UnidadeID" json:"inventarios"`
	Vistorias   []Vistoria           `gorm:"foreignKey:UnidadeID" json:"vistorias"`
	Auditorias  []AuditoriaExecucao  `gorm:"foreignKey:UnidadeID" json:"auditorias"`
	Insights    []IntelligenceSource `gorm:"foreignKey:UnidadeID" json:"insights"`
}
