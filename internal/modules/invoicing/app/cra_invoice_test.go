package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

type craInvoiceRepo struct {
	exists bool
	saved  []domain.Invoice
}

func (r *craInvoiceRepo) SaveInvoice(_ context.Context, inv domain.Invoice) error {
	r.saved = append(r.saved, inv)
	return nil
}

func (r *craInvoiceRepo) GetInvoice(context.Context, kernel.TenantID, uuid.UUID) (domain.Invoice, error) {
	return domain.Invoice{}, domain.ErrInvoiceNotFound
}

func (r *craInvoiceRepo) ListInvoices(context.Context, kernel.TenantID) ([]domain.Invoice, error) {
	return nil, nil
}

func (r *craInvoiceRepo) SaveInvoiceLine(context.Context, domain.InvoiceLine) error { return nil }

func (r *craInvoiceRepo) ListInvoiceLines(context.Context, kernel.TenantID, uuid.UUID) ([]domain.InvoiceLine, error) {
	return nil, nil
}

func (r *craInvoiceRepo) SavePDPQueueItem(context.Context, domain.PDPQueueItem) error { return nil }

func (r *craInvoiceRepo) InvoiceExistsForTimesheet(context.Context, kernel.TenantID, uuid.UUID) (bool, error) {
	return r.exists, nil
}

func (r *craInvoiceRepo) SumNonVirtualInvoicesInPeriod(context.Context, kernel.TenantID, kernel.Period) (int64, int, string, error) {
	return 0, 0, "EUR", nil
}

func TestCreateFromCRAValidation_Idempotent(t *testing.T) {
	repo := &craInvoiceRepo{exists: true}
	svc := NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	inv, err := svc.CreateFromCRAValidation(context.Background(), ports.CreateFromCRACommand{
		TenantID:       tenant,
		TimesheetID:    uuid.New(),
		ClientID:       uuid.New(),
		Month:          "2026-07",
		BillableHours:  8,
		MissionLabel:   "Mission",
		UserLabel:      "User",
		UnitPriceCents: 10000,
	})
	if err != nil {
		t.Fatalf("CreateFromCRAValidation: %v", err)
	}
	if inv.ID != uuid.Nil {
		t.Fatalf("expected empty invoice on duplicate, got %v", inv.ID)
	}
	if len(repo.saved) != 0 {
		t.Fatal("expected no save when invoice already exists")
	}
}

func TestCreateFromCRAValidation_CreatesDraft(t *testing.T) {
	repo := &craInvoiceRepo{}
	svc := NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	clientID := uuid.New()
	inv, err := svc.CreateFromCRAValidation(context.Background(), ports.CreateFromCRACommand{
		TenantID:       tenant,
		TimesheetID:    uuid.New(),
		ClientID:       clientID,
		Month:          "2026-07",
		BillableHours:  10,
		MissionLabel:   "Mission",
		UserLabel:      "User",
		Currency:       "EUR",
		UnitPriceCents: 10000,
	})
	if err != nil {
		t.Fatalf("CreateFromCRAValidation: %v", err)
	}
	if inv.ID == uuid.Nil {
		t.Fatal("expected invoice id")
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected one saved invoice, got %d", len(repo.saved))
	}
	if repo.saved[0].ClientID != clientID {
		t.Fatalf("unexpected client id: %v", repo.saved[0].ClientID)
	}
}
