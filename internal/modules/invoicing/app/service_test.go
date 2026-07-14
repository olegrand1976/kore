package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/adapters/pdp"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

type virtualRepo struct {
	invoices []domain.Invoice
	lines    []domain.InvoiceLine
}

func (r *virtualRepo) SaveInvoice(_ context.Context, inv domain.Invoice) error {
	for i, existing := range r.invoices {
		if existing.ID == inv.ID {
			r.invoices[i] = inv
			return nil
		}
	}
	r.invoices = append(r.invoices, inv)
	return nil
}

func (r *virtualRepo) GetInvoice(_ context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error) {
	for _, inv := range r.invoices {
		if inv.ID == id && inv.TenantID == tenant {
			return inv, nil
		}
	}
	return domain.Invoice{}, domain.ErrInvoiceNotFound
}

func (r *virtualRepo) ListInvoices(context.Context, kernel.TenantID) ([]domain.Invoice, error) {
	return r.invoices, nil
}

func (r *virtualRepo) SaveInvoiceLine(_ context.Context, line domain.InvoiceLine) error {
	r.lines = append(r.lines, line)
	return nil
}

func (r *virtualRepo) ListInvoiceLines(_ context.Context, _ kernel.TenantID, invoiceID uuid.UUID) ([]domain.InvoiceLine, error) {
	var out []domain.InvoiceLine
	for _, l := range r.lines {
		if l.InvoiceID == invoiceID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *virtualRepo) SavePDPQueueItem(context.Context, domain.PDPQueueItem) error { return nil }

func (r *virtualRepo) InvoiceExistsForTimesheet(context.Context, kernel.TenantID, uuid.UUID) (bool, error) {
	return false, nil
}

func (r *virtualRepo) SumNonVirtualInvoicesInPeriod(context.Context, kernel.TenantID, kernel.Period) (int64, int, string, error) {
	return 0, 0, "EUR", nil
}

func TestComputeVirtualCalculatesTotals(t *testing.T) {
	repo := &virtualRepo{}
	svc := NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	period, err := kernel.NewPeriod(time.Now().AddDate(0, -1, 0), time.Now())
	if err != nil {
		t.Fatalf("period: %v", err)
	}
	inv, err := svc.ComputeVirtual(context.Background(), ports.ComputeVirtualCommand{
		TenantID: tenant,
		ClientID: uuid.New(),
		Period:   period,
		Lines: []ports.InvoiceLineInput{{
			Description: "Test",
			Quantity:    2,
			UnitPrice:   5000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("ComputeVirtual: %v", err)
	}
	if inv.TotalAmount != 10000 {
		t.Fatalf("expected total 10000, got %d", inv.TotalAmount)
	}
	if inv.TaxAmount != 2000 {
		t.Fatalf("expected tax 2000, got %d", inv.TaxAmount)
	}
	if inv.Status != domain.InvoiceStatusVirtuelle {
		t.Fatalf("expected virtuelle status, got %s", inv.Status)
	}
}

func TestTransmitUsesPDPGateway(t *testing.T) {
	repo := &virtualRepo{}
	gw := pdp.NewStubGateway()
	svc := NewService(repo, WithPDPGateway(gw))
	tenant := kernel.NewTenantID(uuid.New())
	clientID := uuid.New()
	inv, err := svc.Create(context.Background(), ports.CreateInvoiceCommand{
		TenantID: tenant,
		ClientID: clientID,
		Type:     domain.InvoiceTypeStandard,
		Currency: "EUR",
		Lines: []ports.InvoiceLineInput{{
			Description: "Line",
			Quantity:    1,
			UnitPrice:   10000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	sent, err := svc.Transmit(context.Background(), tenant, inv.ID)
	if err != nil {
		t.Fatalf("Transmit: %v", err)
	}
	if sent.PDPReceiptID == "" {
		t.Fatal("expected pdp receipt id")
	}
	if sent.Status != domain.InvoiceStatusTransmise {
		t.Fatalf("expected transmise, got %s", sent.Status)
	}
}

func TestSyncPDPStatus(t *testing.T) {
	repo := &virtualRepo{}
	svc := NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	inv := domain.NewInvoice(tenant, uuid.New(), domain.InvoiceTypeStandard, "EUR")
	inv.Status = domain.InvoiceStatusTransmise
	inv.PDPReceiptID = "pdp_test"
	_ = repo.SaveInvoice(context.Background(), inv)
	if err := svc.SyncPDPStatus(context.Background(), ports.PDPStatusEvent{
		TenantID:  tenant,
		InvoiceID: inv.ID,
		ReceiptID: "pdp_test",
		Status:    domain.InvoiceStatusAcceptee,
	}); err != nil {
		t.Fatalf("SyncPDPStatus: %v", err)
	}
	updated, _ := repo.GetInvoice(context.Background(), tenant, inv.ID)
	if updated.Status != domain.InvoiceStatusAcceptee {
		t.Fatalf("expected acceptee, got %s", updated.Status)
	}
}
