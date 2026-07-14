package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
)

func (s *service) CreateFromCRAValidation(ctx context.Context, cmd ports.CreateFromCRACommand) (domain.Invoice, error) {
	if cmd.ClientID == uuid.Nil || cmd.BillableHours <= 0 {
		return domain.Invoice{}, nil
	}
	exists, err := s.repo.InvoiceExistsForTimesheet(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return domain.Invoice{}, err
	}
	if exists {
		return domain.Invoice{}, nil
	}
	mission := cmd.MissionLabel
	if mission == "" {
		mission = "Prestation"
	}
	user := cmd.UserLabel
	if user != "" {
		mission = fmt.Sprintf("%s — %s", mission, user)
	}
	desc := fmt.Sprintf("CRA/%s/%s %s", cmd.TimesheetID, cmd.Month, mission)
	return s.Create(ctx, ports.CreateInvoiceCommand{
		TenantID: cmd.TenantID,
		ClientID: cmd.ClientID,
		Type:     domain.InvoiceTypeStandard,
		Currency: cmd.Currency,
		Lines: []ports.InvoiceLineInput{{
			Description: desc,
			Quantity:    cmd.BillableHours,
			UnitPrice:   cmd.UnitPriceCents,
			TaxRate:     cmd.TaxRate,
		}},
	})
}
