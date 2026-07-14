package ssii

import (
	"context"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	ssiiports "github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type MissionRateReader struct {
	repo ssiiports.SSIIRepository
}

func NewMissionRateReader(repo ssiiports.SSIIRepository) craports.MissionRateReader {
	return &MissionRateReader{repo: repo}
}

func (r *MissionRateReader) GetMissionRate(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID) (craports.MissionRate, error) {
	m, err := r.repo.GetMission(ctx, tenant, missionID)
	if err != nil {
		return craports.MissionRate{}, err
	}
	currency := m.Currency
	if currency == "" {
		currency = "EUR"
	}
	return craports.MissionRate{TJMAmount: m.TJMAmount, Currency: currency}, nil
}

var _ craports.MissionRateReader = (*MissionRateReader)(nil)
