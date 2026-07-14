package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/kernel"
)

// SendMonthlyReminders returns users with cra_requis and an incomplete timesheet for the month
// on sociétés with cra_mail_auto enabled. Notification delivery is delegated to callers.
func (s *Service) SendMonthlyReminders(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]uuid.UUID, error) {
	if s.calendar == nil {
		return nil, nil
	}
	summaries, err := s.repo.ListSummariesByTenantMonth(ctx, tenant, month)
	if err != nil {
		return nil, err
	}
	var pending []uuid.UUID
	for _, summary := range summaries {
		userSettings := s.settingsForUser(ctx, tenant, summary.UserID)
		if !userSettings.CraMailAuto {
			continue
		}
		if summary.Status == domain.StatusDefinitif {
			continue
		}
		if summary.WeeksSubmitted < summary.WeeksTotal || summary.TotalMinutes <= 0 {
			pending = append(pending, summary.UserID)
		}
	}
	return pending, nil
}

// BillableMinutesForMonth sums billable minutes for a user/month.
func (s *Service) BillableMinutesForMonth(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month domain.Month) (int, error) {
	ts, err := s.repo.Get(ctx, tenant, userID, month)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, week := range ts.Weeks {
		for _, line := range week.Lines {
			if line.Billable && line.Duration.Minutes > 0 {
				total += line.Duration.Minutes
			}
		}
	}
	return total, nil
}

// LastMondayOfMonth returns the last Monday in the given month (RG-CRA-03).
func LastMondayOfMonth(year int, month time.Month) time.Time {
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	for last.Weekday() != time.Monday {
		last = last.AddDate(0, 0, -1)
	}
	return last
}
