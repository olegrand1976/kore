package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/domain"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateLeaveTypeConfigCommand struct {
	TenantID      kernel.TenantID
	SocieteID     uuid.UUID
	Code          string
	Label         string
	TracksBalance bool
	Active        bool
	SortOrder     int
}

type UpdateLeaveTypeConfigCommand struct {
	TenantID      kernel.TenantID
	ID            uuid.UUID
	Label         string
	TracksBalance bool
	Active        bool
	SortOrder     int
}

type ResetLeaveTypeConfigsCommand struct {
	TenantID  kernel.TenantID
	SocieteID uuid.UUID
}

type LeaveTypeConfigRepository interface {
	ListBySociete(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, activeOnly bool) ([]domain.LeaveTypeConfig, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.LeaveTypeConfig, error)
	GetByCode(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, code string) (domain.LeaveTypeConfig, error)
	Save(ctx context.Context, cfg domain.LeaveTypeConfig) error
	Delete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
	IsCodeUsed(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, code string) (bool, error)
	UpsertDefaults(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, templates []domain.LeaveTypeTemplate) error
}

type OrgSocieteReader interface {
	GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (orgdomain.Societe, error)
	ResolveSocieteIDForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (uuid.UUID, error)
}

type LeaveTypeConfigService interface {
	List(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, activeOnly bool) ([]domain.LeaveTypeConfig, error)
	ListForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveTypeConfig, error)
	Create(ctx context.Context, cmd CreateLeaveTypeConfigCommand) (domain.LeaveTypeConfig, error)
	Update(ctx context.Context, cmd UpdateLeaveTypeConfigCommand) (domain.LeaveTypeConfig, error)
	Delete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
	ResetDefaults(ctx context.Context, cmd ResetLeaveTypeConfigsCommand) ([]domain.LeaveTypeConfig, error)
	BootstrapDefaults(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) error
	ValidateTypeForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, code string) error
}
