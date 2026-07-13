package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrAIDisabled      = errors.New("ai assistance disabled for tenant")
	ErrCapabilityOff   = errors.New("ai capability disabled")
	ErrRequestNotFound = errors.New("ai request not found")
)

type RiskClass string

const (
	RiskMinimal RiskClass = "minimal"
	RiskLimited RiskClass = "limited"
	RiskHigh    RiskClass = "high"
)

type Capability struct {
	Code            string
	RiskClass       RiskClass
	AnnexIII        bool
	Art63Assessment string
	Enabled         bool
	Wave            int
}

type TenantSettings struct {
	TenantID          kernel.TenantID
	Enabled           bool
	NoticeAcceptedAt  *time.Time
	NoticeAcceptedBy  *uuid.UUID
	WorkersInformedAt *time.Time
	LLMProvider       string
}

type RequestLog struct {
	ID             uuid.UUID
	TenantID       kernel.TenantID
	UserID         uuid.UUID
	CapabilityCode string
	EntityType     string
	EntityID       *uuid.UUID
	InputHash      string
	OutputJSON     []byte
	Model          string
	ExplainContext map[string]any
	CreatedAt      time.Time
}

type AnalysisDraft struct {
	Functional   string `json:"functional"`
	Technical    string `json:"technical"`
	Risks        string `json:"risks"`
	TestScenario string `json:"testScenario"`
}

type ExplainFactor struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ExplainResult struct {
	RequestID  uuid.UUID       `json:"requestId"`
	Capability string          `json:"capability"`
	Summary    string          `json:"summary"`
	Factors    []ExplainFactor `json:"factors"`
	Disclaimer string          `json:"disclaimer"`
}
