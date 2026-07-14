package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/modules/ett/domain"
	"github.com/kore/kore/internal/modules/ett/ports"
	"github.com/kore/kore/pkg/kernel"
)

type fakeCRAReader struct {
	minutes int
}

func (f *fakeCRAReader) ConsumedByApplication(context.Context, kernel.TenantID, craports.ApplicationID, kernel.Period) ([]cradomain.Consumption, error) {
	return nil, nil
}

func (f *fakeCRAReader) TimesheetOf(context.Context, kernel.TenantID, uuid.UUID, cradomain.Month) (cradomain.Timesheet, error) {
	day := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)
	return cradomain.Timesheet{
		Weeks: []cradomain.WeekEntry{{
			Lines: []cradomain.TimeLine{{Day: day, Duration: kernel.Duration{Minutes: f.minutes}}},
		}},
	}, nil
}

type fakeETTRepo struct {
	hours float64
}

func (f *fakeETTRepo) SaveRecord(context.Context, domain.WorkTimeRecord) error { return nil }
func (f *fakeETTRepo) GetRecord(context.Context, kernel.TenantID, uuid.UUID) (domain.WorkTimeRecord, error) {
	return domain.WorkTimeRecord{}, nil
}
func (f *fakeETTRepo) FindRecordByUserDate(context.Context, kernel.TenantID, uuid.UUID, time.Time) (domain.WorkTimeRecord, error) {
	return domain.WorkTimeRecord{}, nil
}
func (f *fakeETTRepo) ListRecords(_ context.Context, _ ports.RecordsQuery) ([]domain.WorkTimeRecord, error) {
	in := time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC)
	out := in.Add(time.Duration(f.hours * float64(time.Hour)))
	return []domain.WorkTimeRecord{{ClockIn: &in, ClockOut: &out}}, nil
}
func (f *fakeETTRepo) AppendAuditEntry(context.Context, domain.AuditEntry) error { return nil }
func (f *fakeETTRepo) ListAuditEntries(context.Context, kernel.TenantID, uuid.UUID) ([]domain.AuditEntry, error) {
	return nil, nil
}
func (f *fakeETTRepo) GetCountryRule(context.Context, kernel.TenantID, string) (domain.CountryWorkRule, error) {
	return domain.CountryWorkRule{}, nil
}

func TestReconciliation_AlertOnDelta(t *testing.T) {
	svc := NewReconciliationService(&fakeETTRepo{hours: 1}, &fakeCRAReader{minutes: 480})
	report, err := svc.CompareMonth(context.Background(), kernel.NewTenantID(uuid.New()), uuid.New(), "2026-07")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Alert {
		t.Fatalf("expected alert for large delta, got %+v", report)
	}
}
