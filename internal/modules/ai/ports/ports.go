package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/ai/domain"
	congesdomain "github.com/kore/kore/internal/modules/conges/domain"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
	wfdomain "github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type AnalysisDraftCommand struct {
	TenantID      kernel.TenantID
	UserID        uuid.UUID
	DemandID      uuid.UUID
	Subject       string
	ApplicationID uuid.UUID
}

type AnalysisDraftResult struct {
	Draft     domain.AnalysisDraft `json:"draft"`
	RequestID uuid.UUID            `json:"requestId"`
}

type ClassifyDemandCommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	Subject  string
}

type ClassifyResult struct {
	Category   string    `json:"category"`
	Confidence float64   `json:"confidence"`
	RequestID  uuid.UUID `json:"requestId"`
}

type SimilarDemandsCommand struct {
	TenantID      kernel.TenantID
	UserID        uuid.UUID
	Subject       string
	ApplicationID *uuid.UUID
	Limit         int
}

type SimilarDemand struct {
	DemandID uuid.UUID `json:"demandId"`
	Subject  string    `json:"subject"`
	Score    float64   `json:"score"`
}

type CraPrefillCommand struct {
	TenantID    kernel.TenantID
	UserID      uuid.UUID
	TimesheetID uuid.UUID
	WeekNumber  int
}

type PrefillLine struct {
	Day      string  `json:"day"`
	Duration float64 `json:"duration"`
	Comment  string  `json:"comment"`
}

type CraPrefillResult struct {
	Lines     []PrefillLine `json:"lines"`
	RequestID uuid.UUID     `json:"requestId"`
}

type CraAnomaliesCommand struct {
	TenantID    kernel.TenantID
	UserID      uuid.UUID
	TimesheetID uuid.UUID
}

type CraAnomaly struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Day     string `json:"day,omitempty"`
}

type BudgetEstimateCommand struct {
	TenantID  kernel.TenantID
	UserID    uuid.UUID
	DemandID  uuid.UUID
	BudgetID  uuid.UUID
}

type BudgetEstimateResult struct {
	EffortDays float64   `json:"effortDays"`
	EffortUO   float64   `json:"effortUO"`
	Rationale  string    `json:"rationale"`
	RequestID  uuid.UUID `json:"requestId"`
}

type BudgetDemandSuggestCommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	BudgetID uuid.UUID
	Query    string
	Limit    int
}

type DemandSuggestion struct {
	DemandID uuid.UUID `json:"demandId"`
	Subject  string    `json:"subject"`
	Status   string    `json:"status"`
}

type DashboardBriefingCommand struct {
	TenantID           kernel.TenantID
	UserID             uuid.UUID
	Profile            string
	CraStatus          string
	LeavePending       int
	TmaOpen            int
	BudgetConsumption  float64
	BudgetOverrun      int
	PendingValidations int
}

type BriefingResult struct {
	Text      string    `json:"text"`
	RequestID uuid.UUID `json:"requestId"`
}

type CongesManagerCommand struct {
	TenantID       kernel.TenantID
	UserID         uuid.UUID
	LeaveRequestID uuid.UUID
}

type ManagerContextResult struct {
	Context   string    `json:"context"`
	RequestID uuid.UUID `json:"requestId"`
}

type WorkflowExplainCommand struct {
	TenantID   kernel.TenantID
	UserID     uuid.UUID
	InstanceID uuid.UUID
}

type PublicChatCommand struct {
	Message   string
	SessionID string
}

type ChatResult struct {
	Reply     string    `json:"reply"`
	RequestID uuid.UUID `json:"requestId"`
}

type EnableAICommand struct {
	TenantID          kernel.TenantID
	UserID            uuid.UUID
	NoticeAccepted    bool
	WorkersInformed   bool
}

type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
	MaxTokens    int
}

type CompletionResponse struct {
	Text  string
	Model string
}

type LLMProvider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
}

type Repository interface {
	IsCapabilityEnabled(ctx context.Context, code string) (bool, error)
	GetTenantSettings(ctx context.Context, tenant kernel.TenantID) (domain.TenantSettings, error)
	UpsertTenantSettings(ctx context.Context, settings domain.TenantSettings) error
	InsertRequestLog(ctx context.Context, log domain.RequestLog) error
	GetRequestLog(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestLog, error)
}

type TMAReader interface {
	GetDemand(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (tmadomain.Demand, error)
	ListDemands(ctx context.Context, tenant kernel.TenantID, visibleOnly bool) ([]tmadomain.Demand, error)
	GetAnalysis(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (tmadomain.AnalysisDossier, error)
}

type CRAReader interface {
	GetTimesheetByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (cradomain.Timesheet, error)
	ListRecentTimesheets(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, limit int) ([]cradomain.Timesheet, error)
}

type LeaveReader interface {
	GetLeave(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (congesdomain.LeaveRequest, error)
	ListLeaves(ctx context.Context, tenant kernel.TenantID, status *congesdomain.LeaveStatus) ([]congesdomain.LeaveRequest, error)
	ListBalances(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]congesdomain.LeaveBalance, error)
}

type WorkflowReader interface {
	GetInstance(ctx context.Context, tenant kernel.TenantID, id wfdomain.InstanceID) (wfdomain.WorkflowInstance, error)
	AvailableActions(ctx context.Context, tenant kernel.TenantID, instanceID wfdomain.InstanceID, actor authx.Identity) ([]wfdomain.ActionCode, error)
}

type AIService interface {
	SuggestAnalysisDraft(ctx context.Context, cmd AnalysisDraftCommand) (AnalysisDraftResult, error)
	ClassifyDemand(ctx context.Context, cmd ClassifyDemandCommand) (ClassifyResult, error)
	FindSimilarDemands(ctx context.Context, cmd SimilarDemandsCommand) ([]SimilarDemand, error)
	SuggestCraPrefill(ctx context.Context, cmd CraPrefillCommand) (CraPrefillResult, error)
	ListCraAnomalies(ctx context.Context, cmd CraAnomaliesCommand) ([]CraAnomaly, error)
	EstimateBudgetEffort(ctx context.Context, cmd BudgetEstimateCommand) (BudgetEstimateResult, error)
	SuggestBudgetDemands(ctx context.Context, cmd BudgetDemandSuggestCommand) ([]DemandSuggestion, error)
	DashboardBriefing(ctx context.Context, cmd DashboardBriefingCommand) (BriefingResult, error)
	CongesManagerContext(ctx context.Context, cmd CongesManagerCommand) (ManagerContextResult, error)
	ExplainWorkflow(ctx context.Context, cmd WorkflowExplainCommand) (domain.ExplainResult, error)
	PublicChat(ctx context.Context, cmd PublicChatCommand) (ChatResult, error)
	ExplainRequest(ctx context.Context, tenant kernel.TenantID, requestID uuid.UUID) (domain.ExplainResult, error)
	GetTenantSettings(ctx context.Context, tenant kernel.TenantID) (domain.TenantSettings, error)
	EnableAI(ctx context.Context, cmd EnableAICommand) error
}

type Clock interface {
	Now() time.Time
}
