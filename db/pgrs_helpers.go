package db

import (
	"encoding/json"
	"fmt"
	"time"
)

// CriarPGRS cria um novo PGRS em estado RASCUNHO
func CriarPGRS(unidadeID uint, dadosFormulario map[string]interface{}) (*PGRS, error) {
	dadosJSON, err := json.Marshal(dadosFormulario)
	if err != nil {
		return nil, err
	}

	// Gerar número único
	var count int64
	DB.Model(&PGRS{}).
		Where("EXTRACT(YEAR FROM created_at) = ?", time.Now().Year()).
		Count(&count)

	numero := fmt.Sprintf("PGRS-%d-%03d", time.Now().Year(), count+1)

	pgrs := &PGRS{
		UnidadeID:       unidadeID,
		Numero:          numero,
		Versao:          1,
		Status:          PGRS_STATUS_RASCUNHO,
		DadosFormulario: string(dadosJSON),
	}

	result := DB.Create(pgrs)
	return pgrs, result.Error
}

// SalvarResultadoIA salva o resultado da IA no PGRS
func SalvarResultadoIA(pgrsID uint, auditoria string, relatorio string) error {
	return DB.Model(&PGRS{}).
		Where("id = ?", pgrsID).
		Updates(map[string]interface{}{
			"resultado_auditoria": auditoria,
			"relatorio_gerado":    relatorio,
			"status":              PGRS_STATUS_AGUARDANDO_REVISAO,
		}).Error
}

// AtualizarStatusPGRS muda o status e registra no log
func AtualizarStatusPGRS(pgrsID uint, novoStatus string, observacoes string) error {
	var pgrs PGRS
	if err := DB.First(&pgrs, pgrsID).Error; err != nil {
		return err
	}

	statusAnterior := pgrs.Status
	pgrs.Status = novoStatus

	// Atualizar campos específicos conforme status
	switch novoStatus {
	case PGRS_STATUS_APROVADO:
		now := time.Now()
		pgrs.DataAprovacao = &now
	case PGRS_STATUS_EM_CORRECAO:
		now := time.Now()
		pgrs.DataRevisao = &now
		pgrs.ObservacoesRevisor = observacoes
	}

	if err := DB.Save(&pgrs).Error; err != nil {
		return err
	}

	// Registrar no log
	log := LogPGRS{
		PGRSID:    pgrs.ID,
		Acao:      fmt.Sprintf("STATUS_%s_PARA_%s", statusAnterior, novoStatus),
		Descricao: observacoes,
	}
	DB.Create(&log)

	return nil
}

// ListarPGRS retorna PGRS filtrados
func ListarPGRS(status string, limit int) ([]PGRS, error) {
	var pgrs []PGRS

	query := DB.Model(&PGRS{}).Preload("Unidade")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&pgrs).Error

	return pgrs, err
}
