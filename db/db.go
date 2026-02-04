package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Base com rastreabilidade completa para conformidade legal
type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// --- ESTRUTURA OPERACIONAL (GERADORES) ---

type Gerador struct {
	Base
	RazaoSocial      string     `gorm:"not null" json:"razao_social"`
	CNPJ             string     `gorm:"uniqueIndex;not null" json:"cnpj"`
	CNAE             string     `json:"cnae"`
	Endereco         string     `json:"endereco"`
	Municipio        string     `json:"municipio"`
	Estado           string     `json:"estado"`
	ResponsavelLegal string     `json:"responsavel_legal"`
	Unidades         []Unidade  `gorm:"foreignKey:GeradorID" json:"unidades"`
	Propostas        []Proposta `gorm:"foreignKey:GeradorID" json:"propostas"`
}

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

// --- MÓDULO BUSINESS (PROPOSTAS E CONTRATOS) ---

type Proposta struct {
	Base
	GeradorID       uint    `json:"gerador_id"`
	ClienteNome     string  `json:"cliente_nome"`
	Titulo          string  `json:"titulo"`
	ServicoTipo     string  `json:"servico_tipo"`
	Status          string  `json:"status"` // RASCUNHO, ENVIADA, ACEITA
	ValorTotal      float64 `json:"valor_total"`
	Moeda           string  `gorm:"default:'BRL'" json:"moeda"`
	DescricaoEscopo string  `gorm:"type:text" json:"descricao_escopo"`
	Especificidades string  `gorm:"type:text" json:"especificidades"`
	SumarioIA       string  `gorm:"type:text" json:"sumario_ia"`
	ContratoGerado  bool    `gorm:"default:false" json:"contrato_gerado"`
}

type Contrato struct {
	Base
	PropostaID   uint                   `gorm:"uniqueIndex" json:"proposta_id"`
	GeradorID    uint                   `json:"gerador_id"`
	Numero       string                 `gorm:"uniqueIndex" json:"numero"`
	Conteudo     string                 `gorm:"type:text" json:"conteudo"`
	DataInicio   time.Time              `json:"data_inicio"`
	DataFim      time.Time              `json:"data_fim"`
	Assinado     bool                   `gorm:"default:false" json:"assinado"`
	HashDoc      string                 `json:"hash_doc"`
	Faturamentos []LancamentoFinanceiro `gorm:"foreignKey:ContratoID" json:"faturamentos"`
}

type LancamentoFinanceiro struct {
	Base
	ContratoID  uint      `json:"contrato_id,omitempty"`
	Description string    `gorm:"not null" json:"description"`
	Category    string    `json:"category"`
	Amount      float64   `gorm:"not null" json:"amount"`
	Date        time.Time `gorm:"not null" json:"date"`
	Status      string    `gorm:"not null" json:"status"` // PENDENTE, PAGO
	Type        string    `gorm:"not null" json:"type"`   // entrada, saida
}

// ═══════════════════════════════════════════════════════════
// MÓDULO PGRS (Plano de Gerenciamento de Resíduos Sólidos)
// ═══════════════════════════════════════════════════════════

// Status possíveis para um PGRS
const (
	PGRS_STATUS_RASCUNHO           = "RASCUNHO"
	PGRS_STATUS_EM_GERACAO         = "EM_GERACAO"
	PGRS_STATUS_AGUARDANDO_REVISAO = "AGUARDANDO_REVISAO"
	PGRS_STATUS_EM_CORRECAO        = "EM_CORRECAO"
	PGRS_STATUS_APROVADO           = "APROVADO"
	PGRS_STATUS_PUBLICADO          = "PUBLICADO"
	PGRS_STATUS_ARQUIVADO          = "ARQUIVADO"
)

// PGRS representa um Plano de Gerenciamento de Resíduos Sólidos
type PGRS struct {
	Base

	// Relacionamentos
	UnidadeID uint    `gorm:"not null" json:"unidade_id"`
	Unidade   Unidade `gorm:"foreignKey:UnidadeID" json:"unidade,omitempty"`
	RevisorID *uint   `json:"revisor_id,omitempty"`

	// Metadados
	Numero string `gorm:"uniqueIndex" json:"numero"` // Ex: PGRS-2026-001
	Versao int    `gorm:"default:1" json:"versao"`
	Status string `gorm:"not null;default:'RASCUNHO'" json:"status"`

	// Dados do Formulário (JSON)
	DadosFormulario string `gorm:"type:jsonb" json:"dados_formulario"`

	// Resultados da IA
	ResultadoAuditoria string `gorm:"type:text" json:"resultado_auditoria"`
	RelatorioGerado    string `gorm:"type:text" json:"relatorio_gerado"`

	// Revisão
	ObservacoesRevisor string     `gorm:"type:text" json:"observacoes_revisor,omitempty"`
	DataRevisao        *time.Time `json:"data_revisao,omitempty"`
	DataAprovacao      *time.Time `json:"data_aprovacao,omitempty"`

	// Arquivo Final
	ArquivoDOCX   string `json:"arquivo_docx,omitempty"` // Path no storage
	HashDocumento string `json:"hash_documento,omitempty"`
}

// VersaoPGRS mantém histórico de versões
type VersaoPGRS struct {
	Base
	PGRSID          uint   `gorm:"not null" json:"pgrs_id"`
	NumeroVersao    int    `gorm:"not null" json:"numero_versao"`
	DadosFormulario string `gorm:"type:jsonb" json:"dados_formulario"`
	RelatorioGerado string `gorm:"type:text" json:"relatorio_gerado"`
	Observacoes     string `gorm:"type:text" json:"observacoes,omitempty"`
}

// LogPGRS registra todas as ações no documento
type LogPGRS struct {
	Base
	PGRSID    uint   `gorm:"not null" json:"pgrs_id"`
	Acao      string `gorm:"not null" json:"acao"` // CRIADO, EDITADO, APROVADO, etc.
	Descricao string `gorm:"type:text" json:"descricao"`
}

// --- MÓDULO VISTORIA & EVIDÊNCIAS ---

type Vistoria struct {
	Base
	UnidadeID uint      `gorm:"not null" json:"unidade_id"`
	Data      time.Time `json:"data"`
	Tecnico   string    `json:"tecnico"`
	Setor     string    `json:"setor"`

	CheckSegregacao    bool `gorm:"default:false" json:"check_segregacao"`
	CheckArmazenamento bool `gorm:"default:false" json:"check_armazenamento"`
	CheckIdentificacao bool `gorm:"default:false" json:"check_identificacao"`
	CheckContencao     bool `gorm:"default:false" json:"check_contencao"`

	Observacoes   string      `gorm:"type:text" json:"observacoes"`
	NotasTecnicas string      `gorm:"type:text" json:"notas_tecnicas"` // Sincronizado com SocketService
	TranscricaoIA string      `gorm:"type:text" json:"transcricao_ia"`
	Evidencias    []Evidencia `gorm:"foreignKey:VistoriaID;constraint:OnDelete:CASCADE" json:"evidencias"`
}

type Evidencia struct {
	Base
	VistoriaID uint    `gorm:"index" json:"vistoria_id"`
	Tipo       string  `gorm:"type:varchar(20)" json:"tipo"`
	MimeType   string  `json:"mime_type"`
	StorageURL string  `gorm:"type:text" json:"storage_url"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	FileHash   string  `gorm:"type:text" json:"file_hash"`
}

// --- MÓDULO AUDITORIA ---

type ProtocoloAuditoria struct {
	Base
	Nome      string              `gorm:"not null" json:"nome"`
	Versao    string              `json:"versao"`
	Ativo     bool                `gorm:"default:true" json:"ativo"`
	Perguntas []PerguntaAuditoria `gorm:"foreignKey:ProtocoloID" json:"perguntas"`
}

type PerguntaAuditoria struct {
	Base
	ProtocoloID uint   `json:"protocolo_id"`
	Categoria   string `json:"categoria"`
	Texto       string `gorm:"type:text" json:"texto"`
	Requisito   string `json:"requisito_legal"`
	Peso        int    `json:"peso"`
}

type AuditoriaExecucao struct {
	Base
	UnidadeID      uint                `json:"unidade_id"`
	ProtocoloID    uint                `json:"protocolo_id"`
	Status         string              `json:"status"`
	Score          float64             `json:"score"`
	ExposicaoLevel string              `json:"exposicao_level"`
	LastCheck      time.Time           `json:"last_check"`
	Respostas      []AuditoriaResposta `gorm:"foreignKey:AuditoriaID" json:"respostas"`
	PontosCriticos []PontoCritico      `gorm:"foreignKey:AuditoriaID" json:"pontos_criticos"`
}

type AuditoriaResposta struct {
	Base
	AuditoriaID uint   `json:"auditoria_id"`
	PerguntaID  uint   `json:"pergunta_id"`
	Conforme    bool   `json:"conforme"`
	Observacao  string `gorm:"type:text" json:"observacao"`
}

type PontoCritico struct {
	Base
	AuditoriaID   uint   `json:"auditoria_id"`
	Identificador string `json:"identificador"`
	Titulo        string `json:"title"`
	Descricao     string `json:"desc"`
	Status        string `json:"status"`
	Color         string `json:"color"`
}

// --- MÓDULO INTELLIGENCE ---

type IntelligenceSource struct {
	Base
	UnidadeID   uint                `json:"unidade_id"`
	SourceURL   string              `gorm:"type:text" json:"source_url"`
	SourceType  string              `json:"source_type"`
	TokenUsage  float64             `json:"token_usage"`
	Insight     IntelligenceInsight `gorm:"foreignKey:SourceID" json:"insight"`
	ChatHistory []ChatMessage       `gorm:"foreignKey:SourceID" json:"chat_history"`
}

type IntelligenceInsight struct {
	Base
	SourceID    uint   `gorm:"uniqueIndex" json:"source_id"`
	Summary     string `gorm:"type:text" json:"summary"`
	PgrsImpact  string `gorm:"type:text" json:"pgrs_impact"`
	LegalBase   string `gorm:"type:text" json:"legal_base"`
	Opportunity string `gorm:"type:text" json:"opportunity"`
}

type ChatMessage struct {
	Base
	SourceID uint   `json:"source_id"`
	Role     string `json:"role"`
	Content  string `gorm:"type:text" json:"content"`
}

// --- RESÍDUOS E INVENTÁRIO (PGRS) ---

type Residuo struct {
	Base
	Nome           string `gorm:"not null" json:"nome"`
	CodigoNBR      string `gorm:"index" json:"codigo_nbr"`
	Classe         string `json:"classe"`
	EstadoFisico   string `json:"estado_fisico"`
	Periculosidade bool   `json:"periculosidade"`
}

type InventarioResiduo struct {
	Base
	UnidadeID          uint   `gorm:"not null" json:"unidade_id"`
	Nome               string `json:"nome"`
	Classe             string `json:"classe"`
	DestinacaoPrevista string `json:"destino"`
	QuantidadeEstimada string `json:"quantidade"`
	ResiduoID          uint   `json:"residuo_id"`
	SetorOrigem        string `json:"setor_origem"`
}

func InitDB() {
	log.Println("🐘 [AUREN-DB] Sincronizando Schema Produção v2.9...")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("❌ [DB FATAL] Erro de conexão: %v", err)
	}

	err = db.AutoMigrate(
		&Gerador{}, &Unidade{}, &Residuo{},
		&InventarioResiduo{}, &Vistoria{},
		&Evidencia{}, &ProtocoloAuditoria{},
		&PerguntaAuditoria{}, &AuditoriaExecucao{},
		&AuditoriaResposta{}, &PontoCritico{},
		&Proposta{}, &Contrato{}, &LancamentoFinanceiro{},
		&IntelligenceSource{}, &IntelligenceInsight{}, &ChatMessage{},
		&PGRS{},
		&VersaoPGRS{},
		&LogPGRS{},
	)
	if err != nil {
		log.Fatalf("❌ [DB FATAL] Falha no AutoMigrate: %v", err)
	}

	DB = db
	log.Println("✅ [DB] Schema Auren Consolidado e Pronto para Operação.")
}
