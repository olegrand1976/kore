package pdp

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

// StubGateway simulates a PDP/PA provider for development and tests.
type StubGateway struct {
	mu       sync.Mutex
	receipts map[string]domain.InvoiceStatus
}

func NewStubGateway() *StubGateway {
	return &StubGateway{receipts: make(map[string]domain.InvoiceStatus)}
}

func (g *StubGateway) Transmit(_ context.Context, _ kernel.TenantID, doc ports.En16931Document) (ports.PDPReceipt, error) {
	if doc == nil {
		return ports.PDPReceipt{}, fmt.Errorf("empty en16931 document")
	}
	receiptID := "pdp_" + uuid.New().String()
	g.mu.Lock()
	g.receipts[receiptID] = domain.InvoiceStatusTransmise
	g.mu.Unlock()
	return ports.PDPReceipt{ID: receiptID}, nil
}

func (g *StubGateway) SyncStatus(_ context.Context, receiptID string) (domain.InvoiceStatus, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	status, ok := g.receipts[receiptID]
	if !ok {
		return "", fmt.Errorf("receipt not found: %s", receiptID)
	}
	return status, nil
}

func (g *StubGateway) SetStatus(receiptID string, status domain.InvoiceStatus) {
	g.mu.Lock()
	g.receipts[receiptID] = status
	g.mu.Unlock()
}

var _ ports.PDPGateway = (*StubGateway)(nil)
