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

	if err := ApplyProposedLines(week, proposed); err != nil {
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
	if err := ValidateDayCapacity(lines); err != ErrDayCapacityExceeded {
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

func TestCommercialInfo_Complete(t *testing.T) {
	if (CommercialInfo{}).Complete() {
		t.Fatal("empty commercial info should be incomplete")
	}
	info := CommercialInfo{Client: "ACME", Mission: "Support"}
	if !info.Complete() {
		t.Fatal("expected complete commercial info")
	}
}
