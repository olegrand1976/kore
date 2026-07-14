package cra

import (
	"context"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	ssiiports "github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type missionActivityReader interface {
	ListDailyActivityInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]craports.DailyActivityRow, error)
}

type MissionReader struct {
	repo ssiiports.SSIIRepository
	cra  missionActivityReader
}

func NewMissionReader(repo ssiiports.SSIIRepository, cra missionActivityReader) ssiiports.MissionReader {
	return &MissionReader{repo: repo, cra: cra}
}

func (r *MissionReader) ActiveMissionDays(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID, period kernel.Period) (ssiiports.MissionBilling, error) {
	mission, err := r.repo.GetMission(ctx, tenant, missionID)
	if err != nil {
		return ssiiports.MissionBilling{}, err
	}
	rows, err := r.cra.ListDailyActivityInPeriod(ctx, tenant, period)
	if err != nil {
		return ssiiports.MissionBilling{}, err
	}
	missionKey := missionID.String()
	days := make(map[string]struct{})
	for _, row := range rows {
		if row.MissionID != missionKey || row.Minutes <= 0 {
			continue
		}
		days[row.Day.Format("2006-01-02")] = struct{}{}
	}
	billableDays := float64(len(days))
	currency := mission.Currency
	if currency == "" {
		currency = "EUR"
	}
	return ssiiports.MissionBilling{
		MissionID:   missionID,
		ClientID:    mission.ClientID,
		Days:        billableDays,
		TJMAmount:   mission.TJMAmount,
		Currency:    currency,
		TotalAmount: int64(billableDays * float64(mission.TJMAmount)),
	}, nil
}

var _ ssiiports.MissionReader = (*MissionReader)(nil)
