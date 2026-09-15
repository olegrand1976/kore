package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

const (
	EmbeddingDimensions = 768

	SourceTypeRequestAttachment = "request_attachment"

	AnalysisSectionFunctional   = "functional"
	AnalysisSectionTechnical    = "technical"
	AnalysisSectionRisks        = "risks"
	AnalysisSectionTestScenario = "testScenario"
)

func ValidAnalysisSection(section string) bool {
	switch section {
	case AnalysisSectionFunctional, AnalysisSectionTechnical, AnalysisSectionRisks, AnalysisSectionTestScenario:
		return true
	default:
		return false
	}
}

type DocumentChunk struct {
	ID         uuid.UUID
	TenantID   kernel.TenantID
	SourceType string
	SourceID   uuid.UUID
	DemandID   *uuid.UUID
	ChunkIndex int
	Content    string
	Embedding  []float32
	MimeType   string
	FileName   string
	CreatedAt  time.Time
	Score      float64 // populated on similarity search
}

type RAGSource struct {
	FileName   string  `json:"fileName"`
	ChunkIndex int     `json:"chunkIndex"`
	Score      float64 `json:"score,omitempty"`
}
