package en16931

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/pkg/kernel"
)

func TestMapInvoice(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	inv := domain.Invoice{
		ID:          uuid.New(),
		TenantID:    tenant,
		ClientID:    uuid.New(),
		Type:        domain.InvoiceTypeStandard,
		Status:      domain.InvoiceStatusPreparee,
		Currency:    "EUR",
		TotalAmount: 10000,
		TaxAmount:   2000,
		CreatedAt:   time.Now().UTC(),
		Lines: []domain.InvoiceLine{{
			ID:          uuid.New(),
			Description: "Prestation",
			Quantity:    10,
			UnitPrice:   1000,
			TaxRate:     20,
		}},
	}
	doc := MapInvoice(inv)
	if doc["specificationIdentifier"] != "urn:cen.eu:en16931:2017" {
		t.Fatalf("unexpected spec id: %v", doc["specificationIdentifier"])
	}
	lines, ok := doc["invoiceLines"].([]map[string]any)
	if !ok || len(lines) != 1 {
		t.Fatalf("expected one invoice line, got %v", doc["invoiceLines"])
	}
	total, ok := doc["legalMonetaryTotal"].(map[string]any)
	if !ok || total["taxExclusiveAmount"] != int64(10000) {
		t.Fatalf("unexpected total: %v", total)
	}
}
