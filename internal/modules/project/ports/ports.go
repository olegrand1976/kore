package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateEpicCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	Title         string
	Description   string
	Priority      string
}

type UpdateEpicCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	EpicID        uuid.UUID
	Title         *string
	Description   *string
	Status        *domain.EpicStatus
	Priority      *string
	TargetSprint  **uuid.UUID
}

type CreateSprintCommand struct {
	TenantID       kernel.TenantID
	ApplicationID  uuid.UUID
	Name           string
	Goal           string
	StartDate      time.Time
	EndDate        time.Time
	CapacityPoints *int16
}

type PlanSprintCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	SprintID      uuid.UUID
	DemandIDs     []uuid.UUID
}

type ReorderBacklogCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	DemandIDs     []uuid.UUID
}

type UpdateKanbanConfigCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	Columns       []domain.KanbanColumn
}

type ApplicationReader interface {
	GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (orgdomain.Application, error)
}

type ProjectService interface {
	CreateEpic(ctx context.Context, cmd CreateEpicCommand) (domain.Epic, error)
	ListEpics(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Epic, error)
	UpdateEpic(ctx context.Context, cmd UpdateEpicCommand) (domain.Epic, error)
	CreateSprint(ctx context.Context, cmd CreateSprintCommand) (domain.Sprint, error)
	ListSprints(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Sprint, error)
	StartSprint(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (domain.Sprint, error)
	CloseSprint(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (domain.Sprint, error)
	PlanSprint(ctx context.Context, cmd PlanSprintCommand) error
	ListBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, backlogOnly bool) ([]domain.BacklogItem, error)
	ReorderBacklog(ctx context.Context, cmd ReorderBacklogCommand) error
	GetKanbanConfig(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.KanbanConfig, error)
	SaveKanbanConfig(ctx context.Context, cmd UpdateKanbanConfigCommand) (domain.KanbanConfig, error)
	GetSprintBurndown(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (domain.BurndownSeries, error)
	GetVelocity(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, lastNSprints int) (domain.VelocityReport, error)
	ListAgileApplications(ctx context.Context, tenant kernel.TenantID) ([]orgdomain.Application, error)
	HasProjectArtifacts(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (bool, error)
}

type Repository interface {
	SaveEpic(ctx context.Context, epic domain.Epic) error
	GetEpic(ctx context.Context, tenant kernel.TenantID, appID, id uuid.UUID) (domain.Epic, error)
	ListEpics(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Epic, error)
	SaveSprint(ctx context.Context, sprint domain.Sprint) error
	GetSprint(ctx context.Context, tenant kernel.TenantID, appID, id uuid.UUID) (domain.Sprint, error)
	ListSprints(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Sprint, error)
	GetActiveSprint(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (*domain.Sprint, error)
	SaveKanbanConfig(ctx context.Context, cfg domain.KanbanConfig) error
	GetKanbanConfig(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.KanbanConfig, error)
	AssignDemandsToSprint(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID, demandIDs []uuid.UUID) error
	ReorderBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, demandIDs []uuid.UUID) error
	ListBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, backlogOnly bool) ([]domain.BacklogItem, error)
	CountProjectArtifacts(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (int, error)
	GetSprintBurndown(ctx context.Context, tenant kernel.TenantID, sprint domain.Sprint) (domain.BurndownSeries, error)
	GetVelocity(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, lastN int) (domain.VelocityReport, error)
	ListAgileApplications(ctx context.Context, tenant kernel.TenantID) ([]orgdomain.Application, error)
}
