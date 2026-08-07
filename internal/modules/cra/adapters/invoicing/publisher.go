package invoicing

import (
	"context"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	invoicingdomain "github.com/kore/kore/internal/modules/invoicing/domain"
	invoicingports "github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

type DraftPublisher struct {
	invoicing invoicingports.InvoicingService
}

func NewDraftPublisher(invoicing invoicingports.InvoicingService) craports.InvoiceDraftPublisher {
	return &DraftPublisher{invoicing: invoicing}
}

func (p *DraftPublisher) PublishCRAValidationDraft(ctx context.Context, cmd craports.ValidationInvoiceCommand) (uuid.UUID, error) {
	if p.invoicing == nil {
		return uuid.Nil, nil
	}
	currency := cmd.Currency
	if currency == "" {
		currency = "EUR"
	}
	taxRate := cmd.TaxRate
	if taxRate <= 0 {
		taxRate = 20
	}
	inv, err := p.invoicing.CreateFromCRAValidation(ctx, invoicingports.CreateFromCRACommand{
		TenantID:        cmd.TenantID,
		TimesheetID:     cmd.TimesheetID,
		TimesheetUserID: cmd.TimesheetUserID,
		ClientID:        cmd.ClientID,
		Month:           string(cmd.Month),
		BillableHours:   cmd.BillableHours,
		MissionLabel:    cmd.MissionLabel,
		UserLabel:       cmd.UserLabel,
		Currency:        currency,
		UnitPriceCents:  cmd.UnitPriceCents,
		TaxRate:         taxRate,
		Description:     cmd.Description,
	})
	if err != nil {
		return uuid.Nil, err
	}
	if inv.ID == uuid.Nil {
		return uuid.Nil, invoicingdomain.ErrNoBillableContent
	}
	return inv.ID, nil
}

func (p *DraftPublisher) TimesheetAlreadyInvoiced(ctx context.Context, tenant kernel.TenantID, timesheetID uuid.UUID) (bool, error) {
	if p.invoicing == nil {
		return false, nil
	}
	return p.invoicing.HasInvoiceForTimesheet(ctx, tenant, timesheetID)
}

var _ craports.InvoiceDraftPublisher = (*DraftPublisher)(nil)
