package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

// PR-08.2 / RG-CRA-01: prefill must not overwrite manual entries.
func TestApplyProposedLines_PreservesManual(t *testing.T) {
	day := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	source := SourceRef{Type: "mission", ID: "app-1"}
	week := &WeekEntry{
		ID:         uuid.New(),
		WeekNumber: 1,
		Lines: []TimeLine{
			{
				ID:       uuid.New(),
				Source:   source,
				Day:      day,
				Duration: kernel.Duration{Minutes: 480},
				Origin:   OriginManual,
			},
		},
	}

	proposed := []TimeLine{{
		Source:   source,
		Day:      day,
		Duration: kernel.Duration{Minutes: 240},
	}}

	if err := ApplyProposedLines(week, proposed, DefaultDayCapacityMinutes); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if week.Lines[0].Duration.Minutes != 480 {
		t.Fatalf("manual line overwritten: got %d minutes", week.Lines[0].Duration.Minutes)
	}
}

// RG-CRA-02: day capacity exceeded.
func TestValidateDayCapacity_Exceeded(t *testing.T) {
	day := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)
	lines := []TimeLine{
		{Day: day, Duration: kernel.Duration{Minutes: 300}, Source: SourceRef{Type: "manual", ID: "a"}},
		{Day: day, Duration: kernel.Duration{Minutes: 200}, Source: SourceRef{Type: "manual", ID: "b"}},
	}
	if err := ValidateDayCapacity(lines, DefaultDayCapacityMinutes); err != ErrDayCapacityExceeded {
		t.Fatalf("expected ErrDayCapacityExceeded, got %v", err)
	}
}

func TestDetectAbsenceConflict(t *testing.T) {
	day := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	lines := []TimeLine{
		{Day: day, Duration: kernel.Duration{Minutes: 60}, Source: SourceRef{Type: "conge", ID: "1"}},
		{Day: day, Duration: kernel.Duration{Minutes: 60}, Source: SourceRef{Type: "mission", ID: "2"}},
	}
	if err := DetectAbsenceConflict(lines); err == nil {
		t.Fatal("expected absence conflict")
	}
}

func TestDetectAbsenceConflict_Leave(t *testing.T) {
	day := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	lines := []TimeLine{
		{Day: day, Duration: kernel.Duration{Minutes: 60}, Source: SourceRef{Type: "leave", ID: "1"}},
		{Day: day, Duration: kernel.Duration{Minutes: 60}, Source: SourceRef{Type: "mission", ID: "2"}},
	}
	if err := DetectAbsenceConflict(lines); err == nil {
		t.Fatal("expected leave absence conflict")
	}
}

func TestCommercialInfo_Complete(t *testing.T) {
	if (CommercialInfo{}).Complete() {
		t.Fatal("empty commercial info should be incomplete")
	}
	info := CommercialInfo{Client: "ACME", Mission: "Support"}
	if !info.Complete() {
		t.Fatal("expected complete commercial info")
	}
}

func TestReject_ClearsSubmittedWeeks(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	submitted := now.Add(-time.Hour)
	managerID := uuid.New()
	ts := Timesheet{
		Status: StatusValideSemaine,
		Weeks: []WeekEntry{
			{WeekNumber: 1, SubmittedAt: &submitted},
			{WeekNumber: 2, SubmittedAt: &submitted},
		},
	}

	if err := ts.Reject(now, managerID, "Incomplet"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Status != StatusBrouillon {
		t.Fatalf("expected Brouillon, got %s", ts.Status)
	}
	if ts.RejectReason != "Incomplet" {
		t.Fatalf("expected reject reason, got %q", ts.RejectReason)
	}
	for _, week := range ts.Weeks {
		if week.SubmittedAt != nil {
			t.Fatalf("week %d should have submittedAt cleared", week.WeekNumber)
		}
	}
}

func TestIncompleteDaysInWeek(t *testing.T) {
	month := Month("2026-07")
	day := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	lines := []TimeLine{
		{Day: day, Duration: kernel.Duration{Minutes: 480}, Source: SourceRef{Type: "manual", ID: "a"}},
	}
	missing, err := IncompleteDaysInWeek(month, 1, DefaultWeekStartDay, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missing) == 0 {
		t.Fatal("expected incomplete days in week")
	}
}

func TestReject_FinalTimesheetFails(t *testing.T) {
	ts := Timesheet{Status: StatusDefinitif}
	if err := ts.Reject(time.Now(), uuid.New(), "too late"); err != ErrCRAAlreadyValidated {
		t.Fatalf("expected ErrCRAAlreadyValidated, got %v", err)
	}
}
