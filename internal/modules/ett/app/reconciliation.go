package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/modules/ett/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ReconciliationService struct {
	ett         ports.ETTRepository
	cra         craports.CRAReader
	prestations craports.CRAService
	users       userProfileReader
}

func NewReconciliationService(
	ett ports.ETTRepository,
	cra craports.CRAReader,
	prestations craports.CRAService,
	users userProfileReader,
) *ReconciliationService {
	return &ReconciliationService{ett: ett, cra: cra, prestations: prestations, users: users}
}

func (s *ReconciliationService) CompareMonth(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month string) (ports.ReconciliationReport, error) {
	report := ports.ReconciliationReport{UserID: userID, Month: month}
	parsed, err := cradomain.ParseMonth(month)
	if err != nil {
		return report, err
	}
	craDays := make(map[string]int)
	if ts, err := s.cra.TimesheetOf(ctx, tenant, userID, parsed); err == nil {
		for _, week := range ts.Weeks {
			for _, line := range week.Lines {
				if line.Duration.Minutes <= 0 {
					continue
				}
				report.CRAHours += float64(line.Duration.Minutes) / 60
				if line.Day.Weekday() != time.Saturday && line.Day.Weekday() != time.Sunday {
					key := line.Day.Format("2006-01-02")
					craDays[key] += line.Duration.Minutes
				}
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
	ettDays := make(map[string]bool)
	if err == nil {
		for _, rec := range records {
			key := rec.WorkDate.Format("2006-01-02")
			ettDays[key] = true
			if rec.ClockIn != nil && rec.ClockOut != nil {
				report.ETTHours += rec.ClockOut.Sub(*rec.ClockIn).Hours()
			} else if rec.EffectiveHours > 0 {
				report.ETTHours += rec.EffectiveHours
			}
		}
	}
	for day := range craDays {
		if !ettDays[day] {
			report.MissingETTDays++
		}
	}
	report.DeltaHours = report.CRAHours - report.ETTHours
	if abs(report.DeltaHours) > 1 {
		report.Alert = true
		report.AlertMessage = "Écart CRA / ETT supérieur à 1 h"
	}
	if report.MissingETTDays > 0 {
		report.Alert = true
		if report.AlertMessage != "" {
			report.AlertMessage += " · "
		}
		report.AlertMessage += fmt.Sprintf("%d jour(s) CRA sans relevé ETT", report.MissingETTDays)
	}
	return report, nil
}

func (s *ReconciliationService) CompareTenant(ctx context.Context, tenant kernel.TenantID, month string) ([]ports.ReconciliationReport, error) {
	if s.prestations == nil {
		return nil, nil
	}
	parsed, err := cradomain.ParseMonth(month)
	if err != nil {
		return nil, err
	}
	summaries, err := s.prestations.ListPrestations(ctx, tenant, parsed)
	if err != nil {
		return nil, err
	}
	out := make([]ports.ReconciliationReport, 0, len(summaries))
	for _, summary := range summaries {
		if s.users != nil {
			detail, err := s.users.FindUserDetailByID(ctx, tenant, summary.UserID)
			if err != nil || !detail.SalarieETT {
				continue
			}
		}
		report, err := s.CompareMonth(ctx, tenant, summary.UserID, month)
		if err != nil {
			continue
		}
		report.UserLogin = summary.UserLogin
		report.UserName = strings.TrimSpace(summary.UserPrenom + " " + summary.UserNom)
		out = append(out, report)
	}
	return out, nil
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
