package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/pkg/kernel"
)

type leaveTypeConfigService struct {
	repo ports.LeaveTypeConfigRepository
	org  ports.OrgSocieteReader
}

func NewLeaveTypeConfigService(repo ports.LeaveTypeConfigRepository, org ports.OrgSocieteReader) ports.LeaveTypeConfigService {
	return &leaveTypeConfigService{repo: repo, org: org}
}

func (s *leaveTypeConfigService) List(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, activeOnly bool) ([]domain.LeaveTypeConfig, error) {
	return s.repo.ListBySociete(ctx, tenant, societeID, activeOnly)
}

func (s *leaveTypeConfigService) ListForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveTypeConfig, error) {
	societeID, err := s.org.ResolveSocieteIDForUser(ctx, tenant, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListBySociete(ctx, tenant, societeID, true)
}

func (s *leaveTypeConfigService) Create(ctx context.Context, cmd ports.CreateLeaveTypeConfigCommand) (domain.LeaveTypeConfig, error) {
	code := normalizeLeaveCode(cmd.Code)
	if code == "" {
		return domain.LeaveTypeConfig{}, domain.ErrLeaveTypeNotFound
	}
	if _, err := s.repo.GetByCode(ctx, cmd.TenantID, cmd.SocieteID, code); err == nil {
		return domain.LeaveTypeConfig{}, domain.ErrLeaveTypeCodeExists
	}
	now := time.Now().UTC()
	cfg := domain.LeaveTypeConfig{
		ID:            uuid.New(),
		TenantID:      cmd.TenantID,
		SocieteID:     cmd.SocieteID,
		Code:          code,
		Label:         strings.TrimSpace(cmd.Label),
		TracksBalance: cmd.TracksBalance,
		Active:        cmd.Active,
		SortOrder:     cmd.SortOrder,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if cfg.Label == "" {
		cfg.Label = code
	}
	if err := s.repo.Save(ctx, cfg); err != nil {
		return domain.LeaveTypeConfig{}, err
	}
	return cfg, nil
}

func (s *leaveTypeConfigService) Update(ctx context.Context, cmd ports.UpdateLeaveTypeConfigCommand) (domain.LeaveTypeConfig, error) {
	cfg, err := s.repo.Get(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.LeaveTypeConfig{}, err
	}
	cfg.Label = strings.TrimSpace(cmd.Label)
	if cfg.Label == "" {
		cfg.Label = cfg.Code
	}
	cfg.TracksBalance = cmd.TracksBalance
	cfg.Active = cmd.Active
	cfg.SortOrder = cmd.SortOrder
	cfg.UpdatedAt = time.Now().UTC()
	if err := s.repo.Save(ctx, cfg); err != nil {
		return domain.LeaveTypeConfig{}, err
	}
	return cfg, nil
}

func (s *leaveTypeConfigService) Delete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error {
	cfg, err := s.repo.Get(ctx, tenant, id)
	if err != nil {
		return err
	}
	used, err := s.repo.IsCodeUsed(ctx, tenant, cfg.SocieteID, cfg.Code)
	if err != nil {
		return err
	}
	if used {
		cfg.Active = false
		cfg.UpdatedAt = time.Now().UTC()
		if err := s.repo.Save(ctx, cfg); err != nil {
			return err
		}
		return domain.ErrLeaveTypeInUse
	}
	return s.repo.Delete(ctx, tenant, id)
}

func (s *leaveTypeConfigService) ResetDefaults(ctx context.Context, cmd ports.ResetLeaveTypeConfigsCommand) ([]domain.LeaveTypeConfig, error) {
	societe, err := s.org.GetSociete(ctx, cmd.TenantID, cmd.SocieteID)
	if err != nil {
		return nil, err
	}
	templates, err := domain.DefaultLeaveTypesForCountry(societe.Pays)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpsertDefaults(ctx, cmd.TenantID, cmd.SocieteID, templates); err != nil {
		return nil, err
	}
	return s.repo.ListBySociete(ctx, cmd.TenantID, cmd.SocieteID, false)
}

func (s *leaveTypeConfigService) BootstrapDefaults(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) error {
	existing, err := s.repo.ListBySociete(ctx, tenant, societeID, false)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}
	_, err = s.ResetDefaults(ctx, ports.ResetLeaveTypeConfigsCommand{TenantID: tenant, SocieteID: societeID})
	return err
}

func (s *leaveTypeConfigService) ValidateTypeForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, code string) error {
	societeID, err := s.org.ResolveSocieteIDForUser(ctx, tenant, userID)
	if err != nil {
		return err
	}
	normalized := normalizeLeaveCode(code)
	cfg, err := s.repo.GetByCode(ctx, tenant, societeID, normalized)
	if err != nil || !cfg.Active {
		return domain.ErrUnknownLeaveType
	}
	return nil
}

func normalizeLeaveCode(code string) string {
	return strings.TrimSpace(strings.ToLower(code))
}
