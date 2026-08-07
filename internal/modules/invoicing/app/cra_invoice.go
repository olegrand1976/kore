package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

const defaultCRATaxRate = 20.0

func (s *service) CreateFromCRAValidation(ctx context.Context, cmd ports.CreateFromCRACommand) (domain.Invoice, error) {
	if err := s.requireEnabled(ctx, cmd.TenantID); err != nil {
		return domain.Invoice{}, err
	}
	if cmd.ClientID == uuid.Nil || cmd.BillableHours <= 0 {
		return domain.Invoice{}, domain.ErrNoBillableContent
	}
	if cmd.UnitPriceCents <= 0 {
		return domain.Invoice{}, domain.ErrZeroUnitPrice
	}
	exists, err := s.repo.InvoiceExistsForTimesheet(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return domain.Invoice{}, err
	}
	if exists {
		return domain.Invoice{}, domain.ErrAlreadyInvoiced
	}
	mission := cmd.MissionLabel
	if mission == "" {
		mission = "Prestation"
	}
	user := cmd.UserLabel
	if user != "" {
		mission = fmt.Sprintf("%s — %s", mission, user)
	}
	desc := cmd.Description
	if desc == "" {
		desc = fmt.Sprintf("CRA/%s/%s %s", cmd.TimesheetID, cmd.Month, mission)
	}
	taxRate := cmd.TaxRate
	if taxRate <= 0 {
		taxRate = defaultCRATaxRate
	}
	timesheetID := cmd.TimesheetID
	return s.Create(ctx, ports.CreateInvoiceCommand{
		TenantID:          cmd.TenantID,
		ClientID:          cmd.ClientID,
		Type:              domain.InvoiceTypeStandard,
		Currency:          cmd.Currency,
		SourceTimesheetID: &timesheetID,
		Lines: []ports.InvoiceLineInput{{
			Description: desc,
			Quantity:    cmd.BillableHours,
			UnitPrice:   cmd.UnitPriceCents,
			TaxRate:     taxRate,
		}},
	})
}

func (s *service) HasInvoiceForTimesheet(ctx context.Context, tenant kernel.TenantID, timesheetID uuid.UUID) (bool, error) {
	// Existence check must not depend on the org toggle — used by CRA preview dry-run.
	return s.repo.InvoiceExistsForTimesheet(ctx, tenant, timesheetID)
}
