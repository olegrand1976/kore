package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

type fakeCRARepo struct {
	ts domain.Timesheet
}

func (f *fakeCRARepo) Save(_ context.Context, ts domain.Timesheet) error {
	f.ts = ts
	return nil
}

func (f *fakeCRARepo) GetByID(_ context.Context, _ kernel.TenantID, id ports.TimesheetID) (domain.Timesheet, error) {
	if f.ts.ID == id {
		return f.ts, nil
	}
	return domain.Timesheet{}, domain.ErrTimesheetNotFound
}

func (f *fakeCRARepo) Get(context.Context, kernel.TenantID, ports.UserID, domain.Month) (domain.Timesheet, error) {
	return domain.Timesheet{}, domain.ErrTimesheetNotFound
}

func (f *fakeCRARepo) FindConsumption(context.Context, kernel.TenantID, ports.ApplicationID, kernel.Period) ([]domain.Consumption, error) {
	return nil, nil
}

func (f *fakeCRARepo) ListByUser(context.Context, kernel.TenantID, ports.UserID, int) ([]domain.Timesheet, error) {
	return nil, nil
}

func (f *fakeCRARepo) ListByTenant(context.Context, kernel.TenantID, int) ([]domain.Timesheet, error) {
	return nil, nil
}

func (f *fakeCRARepo) ListSummariesByUser(context.Context, kernel.TenantID, ports.UserID, int) ([]domain.TimesheetSummary, error) {
	return nil, nil
}

func (f *fakeCRARepo) ListSummariesByTenant(context.Context, kernel.TenantID, int) ([]domain.TimesheetSummary, error) {
	return nil, nil
}

func (f *fakeCRARepo) ListSummariesByTenantMonth(_ context.Context, _ kernel.TenantID, month domain.Month) ([]domain.TimesheetSummary, error) {
	return []domain.TimesheetSummary{{
		ID:     f.ts.ID,
		UserID: f.ts.UserID,
		Month:  month,
		Status: f.ts.Status,
	}}, nil
}

func (f *fakeCRARepo) ListDailyActivityInPeriod(context.Context, kernel.TenantID, kernel.Period) ([]ports.DailyActivityRow, error) {
	return nil, nil
}

func (f *fakeCRARepo) DeleteFutureLines(context.Context, kernel.TenantID, domain.SourceRef, time.Time) error {
	return nil
}

func TestValidateAll_SkipsNonSubmitted(t *testing.T) {
	repo := &fakeCRARepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Month:    "2026-07",
		Status:   domain.StatusBrouillon,
		TenantID: kernel.NewTenantID(uuid.New()),
	}}
	svc := &Service{repo: repo, clock: ports.RealClock{}}
	result, err := svc.ValidateAll(context.Background(), ports.ValidateAllCommand{
		TenantID:  repo.ts.TenantID,
		ManagerID: uuid.New(),
		Month:     "2026-07",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Validated != 0 {
		t.Fatalf("expected 0 validated, got %d", result.Validated)
	}
	if len(result.Failed) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(result.Failed))
	}
}
