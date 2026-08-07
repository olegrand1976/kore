package cra

import (
	"context"
	"strings"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/modules/ssii/domain"
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
	var minutes int
	for _, row := range rows {
		if row.MissionID != missionKey || row.Minutes <= 0 {
			continue
		}
		days[row.Day.Format("2006-01-02")] = struct{}{}
		minutes += row.Minutes
	}
	billableDays := float64(len(days))
	billableHours := float64(minutes) / 60.0
	currency := mission.Currency
	if currency == "" {
		currency = "EUR"
	}
	rateUnit := mission.RateUnit
	if rateUnit == "" {
		rateUnit = domain.RateUnitTJM
	}
	billing := ssiiports.MissionBilling{
		MissionID: missionID,
		ClientID:  mission.ClientID,
		Title:     strings.TrimSpace(mission.Title),
		RateUnit:  string(rateUnit),
		Days:      billableDays,
		Hours:     billableHours,
		UnitPrice: mission.TJMAmount,
		Currency:  currency,
	}
	switch rateUnit {
	case domain.RateUnitHourly:
		billing.Quantity = billableHours
		billing.TotalAmount = int64(billableHours * float64(mission.TJMAmount))
	default:
		billing.Quantity = billableDays
		billing.TotalAmount = int64(billableDays * float64(mission.TJMAmount))
	}
	return billing, nil
}

var _ ssiiports.MissionReader = (*MissionReader)(nil)
