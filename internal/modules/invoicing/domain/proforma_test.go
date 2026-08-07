package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

func TestEmitAndValidateProforma(t *testing.T) {
	inv := NewInvoice(kernel.NewTenantID(uuid.New()), uuid.New(), InvoiceTypeStandard, "EUR")
	inv.Status = InvoiceStatusPreparee
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)

	if err := inv.EmitProforma("hash", "client@acme.test", now); err != nil {
		t.Fatalf("EmitProforma: %v", err)
	}
	if inv.Status != InvoiceStatusProforma {
		t.Fatalf("status=%s", inv.Status)
	}
	if err := inv.ValidateProforma(now.Add(time.Hour)); err != nil {
		t.Fatalf("ValidateProforma: %v", err)
	}
	if inv.Status != InvoiceStatusPreparee || inv.ProformaTokenHash != "" || inv.ProformaValidatedAt == nil {
		t.Fatalf("unexpected after validate: %+v", inv)
	}
	inv.MarkInvoiceSent(now.Add(2 * time.Hour))
	if inv.InvoiceSentAt == nil {
		t.Fatal("expected invoiceSentAt")
	}
}

func TestValidateProformaExpired(t *testing.T) {
	inv := NewInvoice(kernel.NewTenantID(uuid.New()), uuid.New(), InvoiceTypeStandard, "EUR")
	inv.Status = InvoiceStatusPreparee
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	_ = inv.EmitProforma("hash", "client@acme.test", now)
	if err := inv.ValidateProforma(now.Add(ProformaTokenTTL + time.Minute)); err != ErrProformaTokenExpired {
		t.Fatalf("expected expired, got %v", err)
	}
}
