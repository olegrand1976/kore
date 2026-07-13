package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/admin/domain"
	"github.com/kore/kore/internal/modules/admin/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo ports.AdminRepository
}

func NewService(repo ports.AdminRepository) ports.AdminService {
	return &service{repo: repo}
}

func (s *service) GetParameters(ctx context.Context, tenant kernel.TenantID, code string) (domain.ParameterSet, error) {
	return s.repo.GetParameterSet(ctx, tenant, code)
}

func (s *service) UpsertParameters(ctx context.Context, cmd ports.UpsertParametersCommand) (domain.ParameterSet, error) {
	ps, err := s.repo.GetParameterSet(ctx, cmd.TenantID, cmd.Code)
	if err != nil {
		ps = domain.NewParameterSet(cmd.TenantID, cmd.Code, cmd.Payload)
	} else {
		ps.Payload = cmd.Payload
		ps.UpdatedAt = time.Now().UTC()
	}
	return ps, s.repo.SaveParameterSet(ctx, ps)
}

func (s *service) ListTemplates(ctx context.Context, tenant kernel.TenantID) ([]domain.Template, error) {
	return s.repo.ListTemplates(ctx, tenant)
}

func (s *service) GetTemplate(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Template, error) {
	return s.repo.GetTemplate(ctx, tenant, id)
}

func (s *service) CreateTemplate(ctx context.Context, cmd ports.CreateTemplateCommand) (domain.Template, error) {
	t := domain.NewTemplate(cmd.TenantID, cmd.Type, cmd.Name, cmd.Content)
	return t, s.repo.SaveTemplate(ctx, t)
}

func (s *service) ListPhoneDirectory(ctx context.Context, tenant kernel.TenantID) ([]domain.PhoneDirectoryEntry, error) {
	return s.repo.ListPhoneDirectory(ctx, tenant)
}

func (s *service) CreatePhoneEntry(ctx context.Context, cmd ports.CreatePhoneEntryCommand) (domain.PhoneDirectoryEntry, error) {
	e := domain.NewPhoneDirectoryEntry(cmd.TenantID, cmd.Label, cmd.Phone)
	e.UserID = cmd.UserID
	if cmd.Visibility != "" {
		e.Visibility = cmd.Visibility
	}
	return e, s.repo.SavePhoneEntry(ctx, e)
}

var _ ports.AdminService = (*service)(nil)
