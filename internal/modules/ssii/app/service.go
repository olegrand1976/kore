package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo ports.SSIIRepository
}

func NewService(repo ports.SSIIRepository) ports.SSIIService {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error) {
	return s.repo.ListMissions(ctx, tenant)
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error) {
	return s.repo.GetMission(ctx, tenant, id)
}

func (s *service) Create(ctx context.Context, cmd ports.CreateMissionCommand) (domain.Mission, error) {
	m := domain.NewMission(cmd.TenantID, cmd.ClientID, cmd.StartDate, cmd.TJMAmount)
	m.EndDate = cmd.EndDate
	m.Currency = cmd.Currency
	m.Technologies = cmd.Technologies
	m.ClientContact = cmd.ClientContact
	if m.Currency == "" {
		m.Currency = "EUR"
	}
	return m, s.repo.SaveMission(ctx, m)
}

func (s *service) Stop(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error) {
	m, err := s.repo.GetMission(ctx, tenant, id)
	if err != nil {
		return domain.Mission{}, err
	}
	if err := m.Stop(); err != nil {
		return domain.Mission{}, err
	}
	return m, s.repo.SaveMission(ctx, m)
}

func (s *service) UpdateEndDate(ctx context.Context, cmd ports.UpdateEndDateCommand) (domain.Mission, error) {
	m, err := s.repo.GetMission(ctx, cmd.TenantID, cmd.MissionID)
	if err != nil {
		return domain.Mission{}, err
	}
	m.SetEndDate(cmd.EndDate)
	return m, s.repo.SaveMission(ctx, m)
}

var _ ports.SSIIService = (*service)(nil)
