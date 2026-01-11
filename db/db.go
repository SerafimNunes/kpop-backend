package db

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Base com rastreabilidade completa
type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// --- ESTRUTURA OPERACIONAL ---

type Gerador struct {
	Base
	RazaoSocial      string    `gorm:"not null" json:"razao_social"`
	CNPJ             string    `gorm:"uniqueIndex;not null" json:"cnpj"`
	CNAE             string    `json:"cnae"`
	Endereco         string    `json:"endereco"`
	Municipio        string    `json:"municipio"`
	Estado           string    `json:"estado"`
	ResponsavelLegal string    `json:"responsavel_legal"`
	Unidades         []Unidade `gorm:"foreignKey:GeradorID" json:"unidades"`
}

type Unidade struct {
	Base
	GeradorID    uint                `gorm:"not null" json:"gerador_id"`
	Nome         string              `gorm:"not null" json:"nome"`
	TipoOperacao string              `json:"tipo_operacao"`
	Inventarios  []InventarioResiduo `gorm:"foreignKey:UnidadeID" json:"inventarios"`
	Vistorias    []Vistoria          `gorm:"foreignKey:UnidadeID" json:"vistorias"`
	Auditorias   []AuditoriaExecucao `gorm:"foreignKey:UnidadeID" json:"auditorias"`
}

// --- MOTOR DE AUDITORIA DINÂMICA (AUREN INTELLIGENCE) ---

type ProtocoloAuditoria struct {
	Base
	Nome      string              `gorm:"not null" json:"nome"` // Ex: "Norma Federal CONAMA 313"
	Versao    string              `json:"versao"`
	Ativo     bool                `gorm:"default:true" json:"ativo"`
	IA_Update bool                `gorm:"default:true" json:"ia_update"` // IA pode atualizar este checklist?
	Perguntas []PerguntaAuditoria `gorm:"foreignKey:ProtocoloID" json:"perguntas"`
}

type PerguntaAuditoria struct {
	Base
	ProtocoloID uint   `json:"protocolo_id"`
	Categoria   string `json:"categoria"` // Ex: "Armazenamento", "Documentação"
	Texto       string `gorm:"type:text" json:"texto"`
	Requisito   string `json:"requisito_legal"`           // Ex: "NBR 12235"
	IALogic     string `gorm:"type:text" json:"ia_logic"` // Instrução para a IA validar a foto desta pergunta
	Peso        int    `json:"peso"`                      // Gravidade: 1 (Leve) a 3 (Crítico)
}

type AuditoriaExecucao struct {
	Base
	UnidadeID   uint                `json:"unidade_id"`
	ProtocoloID uint                `json:"protocolo_id"`
	Status      string              `json:"status"` // "Em Andamento", "Finalizada", "GAP_Report"
	Score       float64             `json:"score"`
	Respostas   []AuditoriaResposta `gorm:"foreignKey:AuditoriaID" json:"respostas"`
}

type AuditoriaResposta struct {
	Base
	AuditoriaID uint   `json:"auditoria_id"`
	PerguntaID  uint   `json:"pergunta_id"`
	Conforme    bool   `json:"conforme"`
	Observacao  string `gorm:"type:text" json:"observacao"`
	EvidenciaID uint   `json:"evidencia_id"` // FK para a tabela de Evidencias (Foto/Vídeo da prova)
}

// --- GESTÃO DE EVIDÊNCIAS E CAMPO ---

type Vistoria struct {
	Base
	UnidadeID  uint        `gorm:"not null" json:"unidade_id"`
	Data       time.Time   `json:"data"`
	Tecnico    string      `json:"tecnico"`
	Setor      string      `json:"setor"`
	Notas      string      `gorm:"type:text" json:"notas"`
	Evidencias []Evidencia `gorm:"foreignKey:VistoriaID" json:"evidencias"`
}

type Evidencia struct {
	Base
	VistoriaID uint   `gorm:"index" json:"vistoria_id"`
	Tipo       string `gorm:"type:varchar(20)" json:"tipo"` // FIELD_EVIDENCE, KNOWLEDGE_SOURCE
	MimeType   string `json:"mime_type"`
	StorageURL string `gorm:"type:text" json:"storage_url"`
	FileHash   string `gorm:"type:text" json:"file_hash"`
	FileSize   int64  `json:"file_size"`
}

// --- BASE DE CONHECIMENTO TÉCNICO (NOVA TABELA PARA O LIBRARIAN) ---

type ConhecimentoTecnico struct {
	Base
	Titulo      string `json:"titulo"`
	Fonte       string `json:"fonte"` // URL ou Nome do Arquivo
	Tipo        string `json:"tipo"`  // PDF, LEGISLAÇÃO, ARTIGO, YOUTUBE
	ConteudoRaw string `gorm:"type:text" json:"conteudo_raw"`
	SumarioIA   string `gorm:"type:text" json:"sumario_ia"`
	FileHash    string `gorm:"index" json:"file_hash"`
}

// --- RESÍDUOS E PNRS ---

type Residuo struct {
	Base
	Nome           string `gorm:"not null" json:"nome"`
	CodigoNBR      string `gorm:"index" json:"codigo_nbr"`
	CodigoIBAMA    string `json:"codigo_ibama"`
	Classe         string `json:"classe"`
	EstadoFisico   string `json:"estado_fisico"`
	Periculosidade bool   `json:"periculosidade"`
}

type InventarioResiduo struct {
	Base
	UnidadeID          uint    `gorm:"not null" json:"unidade_id"`
	ResiduoID          uint    `gorm:"not null" json:"residuo_id"`
	Residuo            Residuo `gorm:"foreignKey:ResiduoID" json:"residuo_detalhes"`
	SetorOrigem        string  `json:"setor_origem"`
	QuantidadeEstimada float64 `json:"quantidade_estimada"`
	UnidadeMedida      string  `json:"unidade_medida"`
	DestinacaoPrevista string  `json:"destinacao_prevista"`
}

type MTR struct {
	Base
	InventarioID uint      `gorm:"not null" json:"inventario_id"`
	DataColeta    time.Time `json:"data_coleta"`
	Status        string    `json:"status"`
}

// --- INICIALIZAÇÃO ---

func InitDB() {
	log.Println("🐘 [AUREN-DB] Conectando ao cluster PostgreSQL (Enterprise Mode)...")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("❌ [DB FATAL] Erro de conexão física: %v", err)
	}

	// Migração completa incluindo a nova tabela de Conhecimento
	err = db.AutoMigrate(
		&Gerador{}, &Unidade{}, &Residuo{},
		&InventarioResiduo{}, &MTR{}, &Vistoria{},
		&Evidencia{}, &ProtocoloAuditoria{},
		&PerguntaAuditoria{}, &AuditoriaExecucao{},
		&AuditoriaResposta{}, &ConhecimentoTecnico{},
	)
	if err != nil {
		log.Fatalf("❌ [DB FATAL] Falha no AutoMigrate: %v", err)
	}

	DB = db
	log.Println("✅ [DB] Auren Core Schema sincronizado com Motor de Auditoria e Base de Conhecimento.")
}

func GenerateHash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}