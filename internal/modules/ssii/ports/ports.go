package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateMissionCommand struct {
	TenantID      kernel.TenantID
	ClientID      uuid.UUID
	StartDate     time.Time
	EndDate       *time.Time
	TJMAmount     int64
	Currency      string
	Technologies  []string
	ClientContact string
}

type UpdateEndDateCommand struct {
	TenantID  kernel.TenantID
	MissionID uuid.UUID
	EndDate   time.Time
}

type SSIIService interface {
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error)
	Create(ctx context.Context, cmd CreateMissionCommand) (domain.Mission, error)
	Stop(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error)
	UpdateEndDate(ctx context.Context, cmd UpdateEndDateCommand) (domain.Mission, error)
}

type SSIIRepository interface {
	SaveMission(ctx context.Context, m domain.Mission) error
	GetMission(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error)
	ListMissions(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error)
}
