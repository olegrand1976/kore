package invoicing

import (
	"context"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	invoicingports "github.com/kore/kore/internal/modules/invoicing/ports"
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
	inv, err := p.invoicing.CreateFromCRAValidation(ctx, invoicingports.CreateFromCRACommand{
		TenantID:       cmd.TenantID,
		TimesheetID:    cmd.TimesheetID,
		ClientID:       cmd.ClientID,
		Month:          string(cmd.Month),
		BillableHours:  cmd.BillableHours,
		MissionLabel:   cmd.MissionLabel,
		UserLabel:      cmd.UserLabel,
		Currency:       currency,
		UnitPriceCents: cmd.UnitPriceCents,
		TaxRate:        cmd.TaxRate,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return inv.ID, nil
}

var _ craports.InvoiceDraftPublisher = (*DraftPublisher)(nil)
