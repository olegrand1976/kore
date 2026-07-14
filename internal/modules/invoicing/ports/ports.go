package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateInvoiceCommand struct {
	TenantID kernel.TenantID
	ClientID uuid.UUID
	Type     domain.InvoiceType
	Currency string
	Lines    []InvoiceLineInput
}

type InvoiceLineInput struct {
	Description string
	Quantity    float64
	UnitPrice   int64
	TaxRate     float64
}

type ComputeVirtualCommand struct {
	TenantID  kernel.TenantID
	ClientID  uuid.UUID
	MissionID *uuid.UUID
	Period    kernel.Period
	Lines     []InvoiceLineInput
}

type En16931Document map[string]any

type PDPReceipt struct {
	ID string
}

type PDPStatusEvent struct {
	TenantID  kernel.TenantID
	InvoiceID uuid.UUID
	ReceiptID string
	Status    domain.InvoiceStatus
}

type PDPGateway interface {
	Transmit(ctx context.Context, tenant kernel.TenantID, doc En16931Document) (PDPReceipt, error)
	SyncStatus(ctx context.Context, receiptID string) (domain.InvoiceStatus, error)
}

type CreateFromCRACommand struct {
	TenantID       kernel.TenantID
	TimesheetID    uuid.UUID
	ClientID       uuid.UUID
	Month          string
	BillableHours  float64
	MissionLabel   string
	UserLabel      string
	Currency       string
	UnitPriceCents int64
	TaxRate        float64
}

type InvoicingService interface {
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.Invoice, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error)
	Create(ctx context.Context, cmd CreateInvoiceCommand) (domain.Invoice, error)
	CreateFromCRAValidation(ctx context.Context, cmd CreateFromCRACommand) (domain.Invoice, error)
	ComputeVirtual(ctx context.Context, cmd ComputeVirtualCommand) (domain.Invoice, error)
	Transmit(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error)
	SyncPDPStatus(ctx context.Context, evt PDPStatusEvent) error
	CreateCreditNote(ctx context.Context, tenant kernel.TenantID, invoiceID uuid.UUID) (domain.Invoice, error)
}

type InvoicingRepository interface {
	SaveInvoice(ctx context.Context, inv domain.Invoice) error
	GetInvoice(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error)
	ListInvoices(ctx context.Context, tenant kernel.TenantID) ([]domain.Invoice, error)
	SaveInvoiceLine(ctx context.Context, line domain.InvoiceLine) error
	ListInvoiceLines(ctx context.Context, tenant kernel.TenantID, invoiceID uuid.UUID) ([]domain.InvoiceLine, error)
	SavePDPQueueItem(ctx context.Context, item domain.PDPQueueItem) error
	InvoiceExistsForTimesheet(ctx context.Context, tenant kernel.TenantID, timesheetID uuid.UUID) (bool, error)
	SumNonVirtualInvoicesInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) (totalAmount int64, invoiceCount int, currency string, err error)
}
