package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/kernel"
)

// SendMonthlyReminders returns cra_requis users with an incomplete (or missing) timesheet for the month.
func (s *Service) SendMonthlyReminders(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]uuid.UUID, error) {
	candidates, err := s.repo.ListReminderCandidatesByMonth(ctx, tenant, month)
	if err != nil {
		return nil, err
	}
	var pending []uuid.UUID
	for _, c := range candidates {
		if c.IsIncomplete() {
			pending = append(pending, c.UserID)
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
