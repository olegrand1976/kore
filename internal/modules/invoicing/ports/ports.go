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
	TenantID kernel.TenantID
	ClientID uuid.UUID
	Period   kernel.Period
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
}
