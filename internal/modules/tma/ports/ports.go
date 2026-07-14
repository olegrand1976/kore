package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	budgetports "github.com/kore/kore/internal/modules/budget/ports"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateDemandCommand struct {
	TenantID         kernel.TenantID
	ApplicationID    uuid.UUID
	AuthorID         uuid.UUID
	Subject          string
	Description      string
	Priority         string
	DueAt            *time.Time
	RequiresChefGate bool
}

type ChefUtilisateurCommand struct {
	TenantID kernel.TenantID
	ID       uuid.UUID
	ActorID  uuid.UUID
}

type AssignCommand struct {
	TenantID   kernel.TenantID
	ID         uuid.UUID
	AssigneeID uuid.UUID
	ActorID    uuid.UUID
}

type AnalysisCommand struct {
	TenantID     kernel.TenantID
	DemandID     uuid.UUID
	ActorID      uuid.UUID
	Functional   string
	Technical    string
	Risks        string
	TestScenario string
}

type ReworkCommand struct {
	TenantID kernel.TenantID
	ID       uuid.UUID
	Reason   string
	ActorID  uuid.UUID
}

type ExportFilter struct {
	TenantID      kernel.TenantID
	ApplicationID *uuid.UUID
	Status        *domain.DemandStatus
	VisibleOnly   bool
}

type ProposedLine struct {
	TenantID   kernel.TenantID
	UserID     uuid.UUID
	SourceType string
	SourceID   uuid.UUID
	Day        time.Time
	Duration   kernel.Duration
	Comment    string
}

type StartWorkflowCommand struct {
	TenantID       kernel.TenantID
	DefinitionCode string
	EntityID       string
}

type FireTransitionCommand struct {
	TenantID   kernel.TenantID
	InstanceID uuid.UUID
	Action     string
	ActorID    uuid.UUID
}

type WorkflowInstance struct {
	ID           uuid.UUID
	CurrentState string
}

type NotificationEvent struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	Trigger  string
	Subject  string
	Body     string
	Vars     map[string]string
}

type TMAService interface {
	CreateDemand(ctx context.Context, cmd CreateDemandCommand) (domain.Demand, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Demand, error)
	ValidateCreation(ctx context.Context, cmd ChefUtilisateurCommand) error
	Assign(ctx context.Context, cmd AssignCommand) error
	TakeOver(ctx context.Context, tenant kernel.TenantID, id, userID uuid.UUID) error
	AddAnalysis(ctx context.Context, cmd AnalysisCommand) error
	Resolve(ctx context.Context, tenant kernel.TenantID, id, userID uuid.UUID) error
	Reopen(ctx context.Context, cmd ReworkCommand) error
	List(ctx context.Context, tenant kernel.TenantID, filter ExportFilter) ([]domain.Demand, error)
	GetAnalysis(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.AnalysisDossier, error)
	ExportXML(ctx context.Context, filter ExportFilter) ([]domain.XmlExportRow, error)
}

type DemandRepository interface {
	Save(ctx context.Context, d domain.Demand) error
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Demand, error)
	List(ctx context.Context, tenant kernel.TenantID, filter ExportFilter) ([]domain.Demand, error)
	SaveAnalysis(ctx context.Context, dossier domain.AnalysisDossier) error
	GetAnalysis(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.AnalysisDossier, error)
}

type WorkflowService interface {
	Start(ctx context.Context, cmd StartWorkflowCommand) (WorkflowInstance, error)
	Fire(ctx context.Context, cmd FireTransitionCommand) (WorkflowInstance, error)
}

type CRAFeeder interface {
	ProposeLines(ctx context.Context, lines []ProposedLine) error
}

type BudgetReader = budgetports.BudgetReader

type NotificationPublisher interface {
	Notify(ctx context.Context, evt NotificationEvent) error
}

type Clock interface {
	Now() time.Time
}
