package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/admin/domain"
	"github.com/kore/kore/pkg/kernel"
)

type UpsertParametersCommand struct {
	TenantID kernel.TenantID
	Code     string
	Payload  map[string]any
}

type CreateTemplateCommand struct {
	TenantID kernel.TenantID
	Type     string
	Name     string
	Content  map[string]any
}

type CreatePhoneEntryCommand struct {
	TenantID   kernel.TenantID
	UserID     *uuid.UUID
	Label      string
	Phone      string
	Visibility string
}

type AdminService interface {
	GetParameters(ctx context.Context, tenant kernel.TenantID, code string) (domain.ParameterSet, error)
	UpsertParameters(ctx context.Context, cmd UpsertParametersCommand) (domain.ParameterSet, error)
	ListTemplates(ctx context.Context, tenant kernel.TenantID) ([]domain.Template, error)
	GetTemplate(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Template, error)
	CreateTemplate(ctx context.Context, cmd CreateTemplateCommand) (domain.Template, error)
	ListPhoneDirectory(ctx context.Context, tenant kernel.TenantID) ([]domain.PhoneDirectoryEntry, error)
	CreatePhoneEntry(ctx context.Context, cmd CreatePhoneEntryCommand) (domain.PhoneDirectoryEntry, error)
}

type AdminRepository interface {
	GetParameterSet(ctx context.Context, tenant kernel.TenantID, code string) (domain.ParameterSet, error)
	SaveParameterSet(ctx context.Context, ps domain.ParameterSet) error
	ListTemplates(ctx context.Context, tenant kernel.TenantID) ([]domain.Template, error)
	GetTemplate(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Template, error)
	SaveTemplate(ctx context.Context, t domain.Template) error
	ListPhoneDirectory(ctx context.Context, tenant kernel.TenantID) ([]domain.PhoneDirectoryEntry, error)
	SavePhoneEntry(ctx context.Context, e domain.PhoneDirectoryEntry) error
}
