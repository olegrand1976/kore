package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateDemandCommand struct {
	TenantID         kernel.TenantID
	ApplicationID    uuid.UUID
	AuthorID         uuid.UUID
	AssigneeID       *uuid.UUID
	Subject          string
	Description      string
	Priority         string
	DueAt            *time.Time
	RequiresChefGate bool
	EpicID           *uuid.UUID
	StoryPoints      *int16
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
	SprintID      *uuid.UUID
	EpicID        *uuid.UUID
	BacklogOnly   bool
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
	InstanceID     *uuid.UUID
	InitialState   *string
	RequesterID    uuid.UUID
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
	SoftDelete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
	List(ctx context.Context, tenant kernel.TenantID, filter ExportFilter) ([]domain.Demand, error)
	GetAnalysis(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.AnalysisDossier, error)
	ExportXML(ctx context.Context, filter ExportFilter) ([]domain.XmlExportRow, error)
}

type DemandRepository interface {
	Save(ctx context.Context, d domain.Demand) error
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Demand, error)
	SoftDelete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID, deletedAt time.Time) error
	List(ctx context.Context, tenant kernel.TenantID, filter ExportFilter) ([]domain.Demand, error)
	// EnsureTicketNumbers assigns sequential ticket numbers to demands that still lack one.
	EnsureTicketNumbers(ctx context.Context, tenant kernel.TenantID) error
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

type NotificationPublisher interface {
	Notify(ctx context.Context, evt NotificationEvent) error
}

type Clock interface {
	Now() time.Time
}

type AgileArtifactValidator interface {
	EpicBelongsToApplication(ctx context.Context, tenant kernel.TenantID, appID, epicID uuid.UUID) (bool, error)
	SprintBelongsToApplication(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (bool, error)
}

type WipChecker interface {
	CheckWip(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, targetStatus string, excludeDemandID *uuid.UUID) error
}
