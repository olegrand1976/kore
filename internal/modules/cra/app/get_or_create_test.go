package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
)

type getOrCreateRepo struct {
	fakeCRARepo
	existing domain.Timesheet
	has      bool
}

func (f *getOrCreateRepo) Get(_ context.Context, _ kernel.TenantID, _ ports.UserID, month domain.Month) (domain.Timesheet, error) {
	if f.has && f.existing.Month == month {
		return f.existing, nil
	}
	return domain.Timesheet{}, domain.ErrTimesheetNotFound
}

func (f *getOrCreateRepo) Save(_ context.Context, ts domain.Timesheet) error {
	f.existing = ts
	f.has = true
	f.ts = ts
	return nil
}

func TestGetOrCreate_CreatesThenReturnsExisting(t *testing.T) {
	repo := &getOrCreateRepo{}
	svc := NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"))
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	month := domain.Month("2026-08")

	created, wasCreated, err := svc.GetOrCreate(context.Background(), tenant, userID, month)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !wasCreated {
		t.Fatal("expected created=true on first call")
	}
	if created.ID == uuid.Nil {
		t.Fatal("expected timesheet id")
	}

	again, wasCreated, err := svc.GetOrCreate(context.Background(), tenant, userID, month)
	if err != nil {
		t.Fatalf("get existing: %v", err)
	}
	if wasCreated {
		t.Fatal("expected created=false on second call")
	}
	if again.ID != created.ID {
		t.Fatalf("id mismatch: got %s want %s", again.ID, created.ID)
	}
}
