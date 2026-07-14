package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
)

type countingConsumptionRepo struct {
	validationRepo
	findCalls atomic.Int32
	data      []domain.Consumption
}

func (r *countingConsumptionRepo) FindConsumption(_ context.Context, _ kernel.TenantID, _ ports.ApplicationID, _ kernel.Period) ([]domain.Consumption, error) {
	r.findCalls.Add(1)
	return r.data, nil
}

func TestConsumedByApplication_UsesCache(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	day := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewPeriod(day, day)
	if err != nil {
		t.Fatalf("period: %v", err)
	}
	repo := &countingConsumptionRepo{
		data: []domain.Consumption{{
			UserID:   uuid.New(),
			Source:   domain.SourceRef{Type: "application", ID: appID.String()},
			Day:      day,
			Duration: kernel.Duration{Minutes: 240},
		}},
	}
	svc := NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"))

	first, err := svc.ConsumedByApplication(context.Background(), tenant, appID, period)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	second, err := svc.ConsumedByApplication(context.Background(), tenant, appID, period)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("expected one consumption row, got %d and %d", len(first), len(second))
	}
	if repo.findCalls.Load() != 1 {
		t.Fatalf("expected one repo call due to cache, got %d", repo.findCalls.Load())
	}
}

func TestConsumedByApplication_InvalidatesAfterSaveWeek(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	appID := uuid.New()
	weekID := uuid.New()
	day := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)
	repo := &countingConsumptionRepo{
		validationRepo: validationRepo{ts: domain.Timesheet{
			ID:       uuid.New(),
			TenantID: tenant,
			UserID:   userID,
			Month:    "2026-07",
			Status:   domain.StatusBrouillon,
			Weeks: []domain.WeekEntry{{
				ID:         weekID,
				WeekNumber: 2,
			}},
		}},
		data: []domain.Consumption{{
			UserID:   userID,
			Source:   domain.SourceRef{Type: "application", ID: appID.String()},
			Day:      day,
			Duration: kernel.Duration{Minutes: 120},
		}},
	}
	svc := NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"))
	period, _ := kernel.NewPeriod(day, day)

	if _, err := svc.ConsumedByApplication(context.Background(), tenant, appID, period); err != nil {
		t.Fatalf("warm cache: %v", err)
	}
	if _, err := svc.SaveWeek(context.Background(), ports.SaveWeekCommand{
		TenantID:    tenant,
		TimesheetID: repo.ts.ID,
		WeekNumber:  2,
		Lines: []domain.TimeLine{{
			Source:   domain.SourceRef{Type: "manual", ID: "default"},
			Day:      day,
			Duration: kernel.Duration{Minutes: 480},
		}},
	}); err != nil {
		t.Fatalf("SaveWeek: %v", err)
	}
	if _, err := svc.ConsumedByApplication(context.Background(), tenant, appID, period); err != nil {
		t.Fatalf("after invalidation: %v", err)
	}
	if repo.findCalls.Load() != 2 {
		t.Fatalf("expected cache invalidation to reload repo, got %d calls", repo.findCalls.Load())
	}
}

func TestTryPublishValidationInvoice_ClientUnresolved(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	svc := NewService(&validationRepo{ts: domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   uuid.New(),
		Month:    "2026-07",
		Status:   domain.StatusValideSemaine,
	}}, nil, nil).WithInvoicePublisher(invoicePublisherStub{})

	outcome := svc.tryPublishValidationInvoice(context.Background(), domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   uuid.New(),
		Month:    "2026-07",
	})
	if outcome.Status != ports.InvoiceDraftSkipped || outcome.Reason != "client_unresolved" {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
}

type invoicePublisherStub struct{}

func (invoicePublisherStub) PublishCRAValidationDraft(context.Context, ports.ValidationInvoiceCommand) (uuid.UUID, error) {
	return uuid.New(), nil
}
