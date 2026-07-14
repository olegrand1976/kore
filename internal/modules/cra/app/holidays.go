package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/calendar"
	"github.com/kore/kore/pkg/kernel"
)

// PrefillPublicHolidays adds prefill lines for public holidays in the month.
func (s *Service) PrefillPublicHolidays(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month domain.Month, countryCode string) (int, error) {
	if countryCode == "" {
		countryCode = "FR"
	}
	ts, err := s.GetOrCreate(ctx, tenant, userID, month)
	if err != nil {
		return 0, err
	}
	if !ts.CanEdit() {
		return 0, domain.ErrCRAAlreadyValidated
	}
	year, mon, err := parseMonthParts(month)
	if err != nil {
		return 0, err
	}
	holidays := calendar.PublicHolidayDates(year, countryCode)
	weekStartDay := s.weekStartDayForUser(ctx, tenant, userID)

	var proposed []domain.TimeLine
	for day := range holidays {
		if int(day.Month()) != mon {
			continue
		}
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			continue
		}
		weekNum, err := domain.MonthWeekNumber(day, month, weekStartDay)
		if err != nil {
			continue
		}
		week := ts.EnsureWeek(weekNum)
		proposed = append(proposed, domain.TimeLine{
			ID:          uuid.New(),
			TenantID:    tenant,
			WeekEntryID: week.ID,
			Source:      domain.SourceRef{Type: "holiday", ID: day.Format("2006-01-02")},
			Day:         day,
			Duration:    kernel.Duration{Minutes: 0},
			Comment:     "Jour férié",
			Billable:    false,
			Origin:      domain.OriginPrefill,
		})
	}
	if len(proposed) == 0 {
		return 0, nil
	}
	capacity := s.settingsForUser(ctx, tenant, userID).DayCapacityMinutes
	byWeek := map[domain.WeekNumber][]domain.TimeLine{}
	for _, line := range proposed {
		weekNum, err := domain.MonthWeekNumber(line.Day.UTC(), month, weekStartDay)
		if err != nil {
			continue
		}
		byWeek[weekNum] = append(byWeek[weekNum], line)
	}
	added := 0
	for weekNum, lines := range byWeek {
		week := ts.EnsureWeek(weekNum)
		before := len(week.Lines)
		if err := domain.ApplyProposedLines(week, lines, capacity); err != nil {
			return added, err
		}
		added += len(week.Lines) - before
	}
	if err := s.repo.Save(ctx, ts); err != nil {
		return 0, err
	}
	s.invalidateConsumptionCache(ctx, tenant)
	return added, nil
}

func parseMonthParts(month domain.Month) (year int, mon int, err error) {
	t, err := time.Parse("2006-01", string(month))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid month: %w", err)
	}
	return t.Year(), int(t.Month()), nil
}
