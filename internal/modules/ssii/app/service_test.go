package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type fakeFeeder struct {
	lines []ports.ProposedMissionLine
}

func (f *fakeFeeder) ProposeLines(_ context.Context, lines []ports.ProposedMissionLine) error {
	f.lines = append(f.lines, lines...)
	return nil
}

type fakeCalendar struct {
	blocked map[string]bool
}

func (f *fakeCalendar) IsHolidayOrLeave(_ context.Context, _ kernel.TenantID, _ uuid.UUID, day time.Time, _ string) (bool, error) {
	return f.blocked[day.Format("2006-01-02")], nil
}

func TestPrefillMissionDays_SkipsBlockedDays(t *testing.T) {
	feeder := &fakeFeeder{}
	calendar := &fakeCalendar{blocked: map[string]bool{
		"2026-07-14": true,
	}}
	svc := &service{feeder: feeder, calendar: calendar}
	mission := domain.Mission{
		ID:        uuid.New(),
		TenantID:  kernel.NewTenantID(uuid.New()),
		StartDate: time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC),
	}
	userID := uuid.New()
	err := svc.prefillMissionDays(context.Background(), mission, []uuid.UUID{userID}, "FR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feeder.lines) == 0 {
		t.Fatal("expected prefill lines")
	}
	for _, line := range feeder.lines {
		if line.Day.Format("2006-01-02") == "2026-07-14" {
			t.Fatal("blocked holiday should not be prefilled")
		}
		_ = cradomain.Month(line.Month)
	}
}
