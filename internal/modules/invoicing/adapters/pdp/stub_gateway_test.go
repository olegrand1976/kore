package pdp

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

func TestStubGatewayTransmitAndSync(t *testing.T) {
	gw := NewStubGateway()
	tenant := kernel.NewTenantID(uuid.New())
	receipt, err := gw.Transmit(context.Background(), tenant, ports.En16931Document{"invoiceNumber": "1"})
	if err != nil {
		t.Fatalf("Transmit: %v", err)
	}
	if receipt.ID == "" {
		t.Fatal("expected receipt id")
	}
	status, err := gw.SyncStatus(context.Background(), receipt.ID)
	if err != nil {
		t.Fatalf("SyncStatus: %v", err)
	}
	if status != domain.InvoiceStatusTransmise {
		t.Fatalf("unexpected status: %s", status)
	}
	gw.SetStatus(receipt.ID, domain.InvoiceStatusAcceptee)
	status, _ = gw.SyncStatus(context.Background(), receipt.ID)
	if status != domain.InvoiceStatusAcceptee {
		t.Fatalf("expected acceptee, got %s", status)
	}
}
