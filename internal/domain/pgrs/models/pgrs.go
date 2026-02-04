package models

import (
	"time"

	shared "auren-platform/internal/shared/models"
)

const (
	StatusRascunho          = "RASCUNHO"
	StatusEmGeracao         = "EM_GERACAO"
	StatusAguardandoRevisao = "AGUARDANDO_REVISAO"
	StatusEmCorrecao        = "EM_CORRECAO"
	StatusAprovado          = "APROVADO"
	StatusPublicado         = "PUBLICADO"
	StatusArquivado         = "ARQUIVADO"
)

type PGRS struct {
	shared.Base

	UnidadeID uint  `gorm:"not null" json:"unidade_id"`
	RevisorID *uint `json:"revisor_id,omitempty"`

	Numero string `gorm:"uniqueIndex" json:"numero"`
	Versao int    `gorm:"default:1" json:"versao"`
	Status string `gorm:"not null;default:'RASCUNHO'" json:"status"`

	DadosFormulario string `gorm:"type:jsonb" json:"dados_formulario"`

	ResultadoAuditoria string     `gorm:"type:text" json:"resultado_auditoria"`
	RelatorioGerado    string     `gorm:"type:text" json:"relatorio_gerado"`
	ObservacoesRevisor string     `gorm:"type:text" json:"observacoes_revisor,omitempty"`
	DataRevisao        *time.Time `json:"data_revisao,omitempty"`
	DataAprovacao      *time.Time `json:"data_aprovacao,omitempty"`

	ArquivoDOCX   string `json:"arquivo_docx,omitempty"`
	HashDocumento string `json:"hash_documento,omitempty"`
}

func (p *PGRS) PodeSerAprovado() bool {
	return p.Status == StatusAguardandoRevisao
}
