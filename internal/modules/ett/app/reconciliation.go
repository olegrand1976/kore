package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/modules/ett/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ReconciliationService struct {
	ett ports.ETTRepository
	cra craports.CRAReader
}

func NewReconciliationService(ett ports.ETTRepository, cra craports.CRAReader) *ReconciliationService {
	return &ReconciliationService{ett: ett, cra: cra}
}

func (s *ReconciliationService) CompareMonth(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month string) (ports.ReconciliationReport, error) {
	report := ports.ReconciliationReport{UserID: userID, Month: month}
	parsed, err := cradomain.ParseMonth(month)
	if err != nil {
		return report, err
	}
	if ts, err := s.cra.TimesheetOf(ctx, tenant, userID, parsed); err == nil {
		for _, week := range ts.Weeks {
			for _, line := range week.Lines {
				report.CRAHours += float64(line.Duration.Minutes) / 60
			}
		}
	}
	from, to := monthBounds(month)
	records, err := s.ett.ListRecords(ctx, ports.RecordsQuery{
		TenantID: tenant,
		UserID:   &userID,
		From:     &from,
		To:       &to,
	})
	if err == nil {
		for _, rec := range records {
			if rec.ClockIn != nil && rec.ClockOut != nil {
				report.ETTHours += rec.ClockOut.Sub(*rec.ClockIn).Hours()
			} else if rec.EffectiveHours > 0 {
				report.ETTHours += rec.EffectiveHours
			}
		}
	}
	report.DeltaHours = report.CRAHours - report.ETTHours
	if abs(report.DeltaHours) > 1 {
		report.Alert = true
		report.AlertMessage = "Écart CRA / ETT supérieur à 1 h"
	}
	return report, nil
}

func monthBounds(month string) (time.Time, time.Time) {
	t, _ := time.Parse("2006-01", month)
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
