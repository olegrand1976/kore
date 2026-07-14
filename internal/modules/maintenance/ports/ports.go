package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateWorkRequestCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	Subject       string
	Description   string
	Priority      string
	DueAt         *time.Time
}

type AssignCommand struct {
	TenantID   kernel.TenantID
	RequestID  uuid.UUID
	AssigneeID uuid.UUID
}

type ProgressCommand struct {
	TenantID        kernel.TenantID
	RequestID       uuid.UUID
	ConsumptionDays float64
}

type MaintenanceService interface {
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.WorkRequest, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkRequest, error)
	Create(ctx context.Context, cmd CreateWorkRequestCommand) (domain.WorkRequest, error)
	Assign(ctx context.Context, cmd AssignCommand) (domain.WorkRequest, error)
	Progress(ctx context.Context, cmd ProgressCommand) (domain.WorkRequest, error)
	Complete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkRequest, error)
}

type MaintenanceRepository interface {
	SaveWorkRequest(ctx context.Context, wr domain.WorkRequest) error
	GetWorkRequest(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkRequest, error)
	ListWorkRequests(ctx context.Context, tenant kernel.TenantID) ([]domain.WorkRequest, error)
}
