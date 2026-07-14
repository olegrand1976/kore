package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/adapters/pdf"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

type validationRepo struct {
	ts domain.Timesheet
}

func (r *validationRepo) Save(_ context.Context, ts domain.Timesheet) error {
	r.ts = ts
	return nil
}

func (r *validationRepo) GetByID(_ context.Context, _ kernel.TenantID, id ports.TimesheetID) (domain.Timesheet, error) {
	if r.ts.ID == id {
		return r.ts, nil
	}
	return domain.Timesheet{}, domain.ErrTimesheetNotFound
}

func (r *validationRepo) Get(context.Context, kernel.TenantID, ports.UserID, domain.Month) (domain.Timesheet, error) {
	return domain.Timesheet{}, domain.ErrTimesheetNotFound
}

func (r *validationRepo) FindConsumption(context.Context, kernel.TenantID, ports.ApplicationID, kernel.Period) ([]domain.Consumption, error) {
	return nil, nil
}

func (r *validationRepo) ListByUser(context.Context, kernel.TenantID, ports.UserID, int) ([]domain.Timesheet, error) {
	return nil, nil
}

func (r *validationRepo) ListByTenant(context.Context, kernel.TenantID, int) ([]domain.Timesheet, error) {
	return nil, nil
}

func (r *validationRepo) ListSummariesByUser(context.Context, kernel.TenantID, ports.UserID, int) ([]domain.TimesheetSummary, error) {
	return nil, nil
}

func (r *validationRepo) ListSummariesByTenant(context.Context, kernel.TenantID, int) ([]domain.TimesheetSummary, error) {
	return nil, nil
}

func (r *validationRepo) ListSummariesByTenantMonth(context.Context, kernel.TenantID, domain.Month) ([]domain.TimesheetSummary, error) {
	return nil, nil
}

func (r *validationRepo) ListDailyActivityInPeriod(context.Context, kernel.TenantID, kernel.Period) ([]ports.DailyActivityRow, error) {
	return nil, nil
}

func (r *validationRepo) DeleteFutureLines(context.Context, kernel.TenantID, domain.SourceRef, time.Time) error {
	return nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestSubmitWeekThenValidateFinal(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	managerID := uuid.New()
	weekID := uuid.New()
	day := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)
	repo := &validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    "2026-07",
		Status:   domain.StatusBrouillon,
		CommercialInfo: domain.CommercialInfo{
			Client:  "ACME",
			Mission: "Projet X",
		},
		Weeks: []domain.WeekEntry{{
			ID:         weekID,
			WeekNumber: 2,
			Lines: []domain.TimeLine{{
				ID:          uuid.New(),
				TenantID:    tenant,
				WeekEntryID: weekID,
				Source:      domain.SourceRef{Type: "manual", ID: "default"},
				Day:         day,
				Duration:    kernel.Duration{Minutes: 480},
				Origin:      domain.OriginManual,
			}},
		}},
	}}
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo, nil, nil).
		WithClock(fixedClock{now: now})

	if err := svc.SubmitWeek(context.Background(), ports.SubmitWeekCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		WeekNumber:  2,
		UserID:      userID,
	}); err != nil {
		t.Fatalf("SubmitWeek: %v", err)
	}
	if repo.ts.Status != domain.StatusValideSemaine {
		t.Fatalf("expected ValidéSemaine, got %s", repo.ts.Status)
	}

	if _, err := svc.ValidateFinal(context.Background(), ports.ManagerValidateCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		ManagerID:   managerID,
	}); err != nil {
		t.Fatalf("ValidateFinal: %v", err)
	}
	if repo.ts.Status != domain.StatusDefinitif {
		t.Fatalf("expected Définitif, got %s", repo.ts.Status)
	}
	if repo.ts.ValidatedBy == nil || *repo.ts.ValidatedBy != managerID {
		t.Fatal("expected validatedBy manager")
	}
}

func TestGeneratePDF_RequiresCommercialInfo(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   uuid.New(),
		Month:    "2026-07",
		Status:   domain.StatusBrouillon,
	}}
	svc := NewService(repo, nil, nil).WithPDFRenderer(pdf.NewStubRenderer())

	_, err := svc.GeneratePDF(context.Background(), tenant, repo.ts.ID)
	if err != domain.ErrCommercialInfoRequired {
		t.Fatalf("expected ErrCommercialInfoRequired, got %v", err)
	}

	repo.ts.CommercialInfo = domain.CommercialInfo{Client: "ACME", Mission: "Projet X"}
	_, err = svc.GeneratePDF(context.Background(), tenant, repo.ts.ID)
	if err != nil {
		t.Fatalf("expected PDF success, got %v", err)
	}
}

func TestResolveUnitPriceCents_FromMissionTJM(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	missionID := uuid.New()
	svc := NewService(nil, nil, nil).WithMissionRateReader(missionRateStub{
		rate: ports.MissionRate{TJMAmount: 80000, Currency: "EUR"},
	})
	ts := domain.Timesheet{TenantID: tenant, UserID: uuid.New()}
	price, currency := svc.resolveUnitPriceCents(context.Background(), ts, missionID)
	if currency != "EUR" {
		t.Fatalf("expected EUR, got %s", currency)
	}
	// 800 EUR/day → 80000 cents / 480 min * 60 = 10000 cents/h
	if price != 10000 {
		t.Fatalf("expected 10000 cents/h, got %d", price)
	}
}

type missionRateStub struct {
	rate ports.MissionRate
	err  error
}

func (m missionRateStub) GetMissionRate(context.Context, kernel.TenantID, uuid.UUID) (ports.MissionRate, error) {
	return m.rate, m.err
}

func TestValidateFinal_RequiresSubmittedStatus(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   uuid.New(),
		Month:    "2026-07",
		Status:   domain.StatusBrouillon,
		CommercialInfo: domain.CommercialInfo{
			Client:  "ACME",
			Mission: "Projet X",
		},
	}}
	svc := NewService(repo, nil, nil)

	_, err := svc.ValidateFinal(context.Background(), ports.ManagerValidateCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		ManagerID:   uuid.New(),
	})
	if err != domain.ErrWeekIncomplete {
		t.Fatalf("expected ErrWeekIncomplete, got %v", err)
	}
}

func TestValidateFinal_RequiresCommercialInfo(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   uuid.New(),
		Month:    "2026-07",
		Status:   domain.StatusValideSemaine,
	}}
	svc := NewService(repo, nil, nil)

	_, err := svc.ValidateFinal(context.Background(), ports.ManagerValidateCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		ManagerID:   uuid.New(),
	})
	if err != domain.ErrCommercialInfoRequired {
		t.Fatalf("expected ErrCommercialInfoRequired, got %v", err)
	}
}
