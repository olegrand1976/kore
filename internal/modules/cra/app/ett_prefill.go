package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/kernel"
)

func monthBounds(month domain.Month) (time.Time, time.Time, error) {
	year, mon, err := parseMonthParts(month)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	from := time.Date(year, time.Month(mon), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return from, to, nil
}

// PrefillFromETT adds prefill lines from ETT work records for days without CRA activity.
func (s *Service) PrefillFromETT(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month domain.Month) (int, error) {
	if s.ettRecords == nil {
		return 0, nil
	}
	ts, err := s.GetOrCreate(ctx, tenant, userID, month)
	if err != nil {
		return 0, err
	}
	if !ts.CanEdit() {
		return 0, domain.ErrCRAAlreadyValidated
	}
	from, to, err := monthBounds(month)
	if err != nil {
		return 0, err
	}
	records, err := s.ettRecords.ListUserDayHours(ctx, tenant, userID, from, to)
	if err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, nil
	}

	settings := s.settingsForUser(ctx, tenant, userID)
	weekStartDay := settings.WeekStartDay
	capacity := settings.DayCapacityMinutes

	existingMinutes := make(map[string]int)
	for _, week := range ts.Weeks {
		for _, line := range week.Lines {
			if line.Duration.Minutes <= 0 {
				continue
			}
			key := line.Day.Format("2006-01-02")
			existingMinutes[key] += line.Duration.Minutes
		}
	}

	var proposed []domain.TimeLine
	for _, rec := range records {
		day := rec.WorkDate.UTC()
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			continue
		}
		key := day.Format("2006-01-02")
		if existingMinutes[key] > 0 {
			continue
		}
		minutes := int(rec.Hours * 60)
		if minutes <= 0 {
			continue
		}
		if capacity > 0 && minutes > capacity {
			minutes = capacity
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
			Source:      domain.SourceRef{Type: "ett", ID: key},
			Day:         day,
			Duration:    kernel.Duration{Minutes: minutes},
			Comment:     fmt.Sprintf("Relevé ETT %.1fh", rec.Hours),
			Billable:    true,
			Origin:      domain.OriginPrefill,
		})
	}
	if len(proposed) == 0 {
		return 0, nil
	}

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
