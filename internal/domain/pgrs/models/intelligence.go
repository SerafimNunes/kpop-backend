package models

// Intelligence placeholders
type IntelligenceSource struct {
	Base
	UnidadeID uint `json:"unidade_id"`
}

type IntelligenceInsight struct {
	Base
	SourceID uint `json:"source_id"`
}

type ChatMessage struct {
	Base
	SourceID uint   `json:"source_id"`
	Role     string `json:"role"`
	Content  string `json:"content"`
}
