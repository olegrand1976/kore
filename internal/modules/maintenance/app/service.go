package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/internal/modules/maintenance/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo   ports.MaintenanceRepository
	feeder ports.CRAFeeder
}

func NewService(repo ports.MaintenanceRepository, feeder ports.CRAFeeder) ports.MaintenanceService {
	return &service{repo: repo, feeder: feeder}
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID) ([]domain.WorkRequest, error) {
	return s.repo.ListWorkRequests(ctx, tenant)
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkRequest, error) {
	return s.repo.GetWorkRequest(ctx, tenant, id)
}

func (s *service) Create(ctx context.Context, cmd ports.CreateWorkRequestCommand) (domain.WorkRequest, error) {
	wr := domain.NewWorkRequest(cmd.TenantID, cmd.ApplicationID, cmd.Subject)
	return wr, s.repo.SaveWorkRequest(ctx, wr)
}

func (s *service) Assign(ctx context.Context, cmd ports.AssignCommand) (domain.WorkRequest, error) {
	wr, err := s.repo.GetWorkRequest(ctx, cmd.TenantID, cmd.RequestID)
	if err != nil {
		return domain.WorkRequest{}, err
	}
	wr.Assign(cmd.AssigneeID)
	return wr, s.repo.SaveWorkRequest(ctx, wr)
}

func (s *service) Progress(ctx context.Context, cmd ports.ProgressCommand) (domain.WorkRequest, error) {
	wr, err := s.repo.GetWorkRequest(ctx, cmd.TenantID, cmd.RequestID)
	if err != nil {
		return domain.WorkRequest{}, err
	}
	if err := wr.Progress(cmd.ConsumptionDays); err != nil {
		return domain.WorkRequest{}, err
	}
	return wr, s.repo.SaveWorkRequest(ctx, wr)
}

func (s *service) Complete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkRequest, error) {
	wr, err := s.repo.GetWorkRequest(ctx, tenant, id)
	if err != nil {
		return domain.WorkRequest{}, err
	}
	if err := wr.Complete(); err != nil {
		return domain.WorkRequest{}, err
	}
	if err := s.repo.SaveWorkRequest(ctx, wr); err != nil {
		return domain.WorkRequest{}, err
	}
	if s.feeder != nil && wr.AssigneeID != nil && wr.CompletedAt != nil {
		days := wr.ConsumptionDays
		if days <= 0 {
			days = 1
		}
		minutes := int(days * 480)
		day := time.Date(wr.CompletedAt.Year(), wr.CompletedAt.Month(), wr.CompletedAt.Day(), 0, 0, 0, 0, time.UTC)
		if err := s.feeder.ProposeLines(ctx, []ports.ProposedLine{{
			TenantID:   tenant,
			UserID:     *wr.AssigneeID,
			SourceType: "work_request",
			SourceID:   wr.ID,
			Day:        day,
			Duration:   kernel.Duration{Minutes: minutes},
			Comment:    wr.Subject,
		}}); err != nil {
			return domain.WorkRequest{}, err
		}
	}
	return wr, nil
}

var _ ports.MaintenanceService = (*service)(nil)
