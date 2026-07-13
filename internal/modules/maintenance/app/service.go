package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/internal/modules/maintenance/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo ports.MaintenanceRepository
}

func NewService(repo ports.MaintenanceRepository) ports.MaintenanceService {
	return &service{repo: repo}
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
	return wr, s.repo.SaveWorkRequest(ctx, wr)
}

var _ ports.MaintenanceService = (*service)(nil)
