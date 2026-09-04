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

func (r *validationRepo) ListReminderCandidatesByMonth(context.Context, kernel.TenantID, domain.Month) ([]domain.ReminderCandidate, error) {
	return nil, nil
}

func (r *validationRepo) ListDailyActivityInPeriod(context.Context, kernel.TenantID, kernel.Period) ([]ports.DailyActivityRow, error) {
	return nil, nil
}

func (r *validationRepo) Delete(_ context.Context, _ kernel.TenantID, id ports.TimesheetID) error {
	if r.ts.ID != id {
		return domain.ErrTimesheetNotFound
	}
	r.ts = domain.Timesheet{}
	return nil
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

func TestResolveSellUnitPriceCents_FromMissionTJM(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	missionID := uuid.New()
	svc := NewService(nil, nil, nil).WithMissionRateReader(missionRateStub{
		rate: ports.MissionRate{TJMAmount: 80000, RateUnit: "tjm", Currency: "EUR"},
	})
	ts := domain.Timesheet{
		TenantID:       tenant,
		UserID:         uuid.New(),
		CommercialInfo: domain.CommercialInfo{MissionID: &missionID},
	}
	price, currency := svc.resolveSellUnitPriceCents(context.Background(), ts)
	if currency != "EUR" {
		t.Fatalf("expected EUR, got %s", currency)
	}
	// 800 EUR/day → 80000 cents / 480 min * 60 = 10000 cents/h
	if price != 10000 {
		t.Fatalf("expected 10000 cents/h, got %d", price)
	}
}

func TestResolveSellUnitPriceCents_FromMissionHourly(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	missionID := uuid.New()
	svc := NewService(nil, nil, nil).WithMissionRateReader(missionRateStub{
		rate: ports.MissionRate{TJMAmount: 12500, RateUnit: "hourly", Currency: "EUR"},
	})
	ts := domain.Timesheet{
		TenantID:       tenant,
		UserID:         uuid.New(),
		CommercialInfo: domain.CommercialInfo{MissionID: &missionID},
	}
	price, currency := svc.resolveSellUnitPriceCents(context.Background(), ts)
	if currency != "EUR" || price != 12500 {
		t.Fatalf("expected 12500 EUR/h cents, got %d %s", price, currency)
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

func TestSaveWeek_AllowsDuplicateActivityTypesOnSameDay(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	weekID := uuid.New()
	day := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	source := domain.SourceRef{Type: "manual", ID: "default"}
	repo := &validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    "2026-07",
		Status:   domain.StatusBrouillon,
		Weeks: []domain.WeekEntry{{
			ID:         weekID,
			WeekNumber: 1,
			Lines: []domain.TimeLine{{
				ID:          uuid.New(),
				TenantID:    tenant,
				WeekEntryID: weekID,
				Source:      source,
				Day:         day,
				Duration:    kernel.Duration{Minutes: 300},
				Origin:      domain.OriginManual,
			}},
		}},
	}}
	svc := NewService(repo, nil, nil)

	if _, err := svc.SaveWeek(context.Background(), ports.SaveWeekCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		WeekNumber:  1,
		Lines: []domain.TimeLine{
			{
				Source:   source,
				Day:      day,
				Duration: kernel.Duration{Minutes: 300},
				Comment:  "prestation 1",
			},
			{
				Source:   source,
				Day:      day,
				Duration: kernel.Duration{Minutes: 180},
				Comment:  "prestation 2",
			},
		},
	}); err != nil {
		t.Fatalf("SaveWeek: %v", err)
	}

	week, _ := repo.ts.Week(1)
	if week == nil {
		t.Fatal("week not found after save")
	}
	if len(week.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(week.Lines))
	}
	if week.Lines[0].Comment != "prestation 1" || week.Lines[1].Comment != "prestation 2" {
		t.Fatalf("unexpected comments: %q, %q", week.Lines[0].Comment, week.Lines[1].Comment)
	}
}

func TestSaveWeek_KeepsKnownLineIDsAndRejectsForeignOnes(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	weekID := uuid.New()
	day := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	source := domain.SourceRef{Type: "manual", ID: "default"}
	existingID := uuid.New()
	foreignID := uuid.New()
	repo := &validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   uuid.New(),
		Month:    "2026-07",
		Status:   domain.StatusBrouillon,
		Weeks: []domain.WeekEntry{{
			ID:         weekID,
			WeekNumber: 1,
			Lines: []domain.TimeLine{{
				ID:          existingID,
				TenantID:    tenant,
				WeekEntryID: weekID,
				Source:      source,
				Day:         day,
				Duration:    kernel.Duration{Minutes: 300},
				Origin:      domain.OriginPrefill,
			}},
		}},
	}}
	svc := NewService(repo, nil, nil)

	if _, err := svc.SaveWeek(context.Background(), ports.SaveWeekCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		WeekNumber:  1,
		Lines: []domain.TimeLine{
			// Known ID: kept, so the grid row is not remounted on save.
			{ID: existingID, Source: source, Day: day, Duration: kernel.Duration{Minutes: 375}, Origin: domain.OriginPrefill},
			// Unknown ID: must be replaced, the insert has no ON CONFLICT clause.
			{ID: foreignID, Source: source, Day: day, Duration: kernel.Duration{Minutes: 60}},
		},
	}); err != nil {
		t.Fatalf("SaveWeek: %v", err)
	}

	week, _ := repo.ts.Week(1)
	if week == nil || len(week.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %+v", week)
	}
	if week.Lines[0].ID != existingID {
		t.Fatalf("known line ID not preserved: got %s, want %s", week.Lines[0].ID, existingID)
	}
	if week.Lines[0].Duration.Minutes != 375 {
		t.Fatalf("expected updated duration 375, got %d", week.Lines[0].Duration.Minutes)
	}
	if week.Lines[0].Origin != domain.OriginPrefill {
		t.Fatalf("origin not preserved: got %q", week.Lines[0].Origin)
	}
	if week.Lines[1].ID == foreignID {
		t.Fatal("foreign line ID must not be reused")
	}
	if week.Lines[1].ID == uuid.Nil {
		t.Fatal("expected a generated ID for the new line")
	}
}

func TestSaveWeek_KeepsCommentOnlyLine(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	weekID := uuid.New()
	day := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	source := domain.SourceRef{Type: "manual", ID: "default"}
	repo := &validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    "2026-07",
		Status:   domain.StatusBrouillon,
		Weeks: []domain.WeekEntry{{
			ID:         weekID,
			WeekNumber: 1,
			Lines:      nil,
		}},
	}}
	svc := NewService(repo, nil, nil)

	if _, err := svc.SaveWeek(context.Background(), ports.SaveWeekCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		WeekNumber:  1,
		Lines: []domain.TimeLine{
			{
				Source:   source,
				Day:      day,
				Duration: kernel.Duration{Minutes: 0},
				Comment:  "  note sans heures  ",
			},
			{
				Source:   source,
				Day:      day.AddDate(0, 0, 1),
				Duration: kernel.Duration{Minutes: 0},
				Comment:  "   ",
			},
			{
				Source:   source,
				Day:      day.AddDate(0, 0, 2),
				Duration: kernel.Duration{Minutes: -30},
				Comment:  "negatif",
			},
			{
				Source:   source,
				Day:      day.AddDate(0, 0, 3),
				Duration: kernel.Duration{Minutes: 120},
				Comment:  "ok",
			},
		},
	}); err != nil {
		t.Fatalf("SaveWeek: %v", err)
	}

	week, _ := repo.ts.Week(1)
	if week == nil {
		t.Fatal("week not found after save")
	}
	if len(week.Lines) != 2 {
		t.Fatalf("expected 2 lines (comment-only + timed), got %d", len(week.Lines))
	}
	if week.Lines[0].Duration.Minutes != 0 || week.Lines[0].Comment != "  note sans heures  " {
		t.Fatalf("unexpected comment-only line: minutes=%d comment=%q", week.Lines[0].Duration.Minutes, week.Lines[0].Comment)
	}
	if week.Lines[1].Duration.Minutes != 120 || week.Lines[1].Comment != "ok" {
		t.Fatalf("unexpected timed line: minutes=%d comment=%q", week.Lines[1].Duration.Minutes, week.Lines[1].Comment)
	}
}
