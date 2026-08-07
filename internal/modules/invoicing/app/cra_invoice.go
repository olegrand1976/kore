package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	inv, lines, err := s.buildPrepareeInvoice(ports.CreateInvoiceCommand{
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
	if err != nil {
		return domain.Invoice{}, err
	}
	// Start before persist so a workflow failure does not leave an orphan invoice
	// that would permanently block CRA retries via ErrAlreadyInvoiced.
	if err := s.startCRAProformaWorkflow(ctx, cmd.TenantID, inv.ID, cmd.TimesheetUserID, nil); err != nil {
		return domain.Invoice{}, err
	}
	if err := s.repo.SaveInvoiceWithLines(ctx, inv, lines); err != nil {
		return domain.Invoice{}, err
	}
	return inv, nil
}

func (s *service) HasInvoiceForTimesheet(ctx context.Context, tenant kernel.TenantID, timesheetID uuid.UUID) (bool, error) {
	// Existence check must not depend on the org toggle — used by CRA preview dry-run.
	return s.repo.InvoiceExistsForTimesheet(ctx, tenant, timesheetID)
}

func (s *service) startCRAProformaWorkflow(
	ctx context.Context,
	tenant kernel.TenantID,
	invoiceID, requesterID uuid.UUID,
	initialState *string,
) error {
	if s.workflow == nil {
		return nil
	}
	id := invoiceID
	_, err := s.workflow.Start(ctx, ports.StartWorkflowCommand{
		TenantID:       tenant,
		DefinitionCode: ports.DefinitionCodeCRAProforma,
		EntityID:       invoiceID.String(),
		InstanceID:     &id,
		RequesterID:    requesterID,
		InitialState:   initialState,
	})
	return err
}

// fireCRAProformaTransition is best-effort for legacy invoices; ensures an instance then fires.
func (s *service) fireCRAProformaTransition(
	ctx context.Context,
	inv domain.Invoice,
	actorID uuid.UUID,
	action string,
	requesterID uuid.UUID,
) {
	if s.workflow == nil {
		return
	}
	_, err := s.workflow.Fire(ctx, ports.FireTransitionCommand{
		TenantID:   inv.TenantID,
		InstanceID: inv.ID,
		Action:     action,
		ActorID:    actorID,
	})
	if errors.Is(err, ports.ErrWorkflowInstanceNotFound) {
		initial := ensureWorkflowStateForAction(action, inv)
		if startErr := s.startCRAProformaWorkflow(ctx, inv.TenantID, inv.ID, requesterID, &initial); startErr != nil {
			slog.Default().WarnContext(ctx, "invoicing workflow ensure start failed",
				"invoiceId", inv.ID, "action", action, "err", startErr)
			return
		}
		_, err = s.workflow.Fire(ctx, ports.FireTransitionCommand{
			TenantID:   inv.TenantID,
			InstanceID: inv.ID,
			Action:     action,
			ActorID:    actorID,
		})
	}
	if err != nil {
		slog.Default().WarnContext(ctx, "invoicing workflow fire failed",
			"invoiceId", inv.ID, "action", action, "err", err)
	}
}

// ensureWorkflowStateForAction picks the WF state before the transition, not the post-decision invoice status.
func ensureWorkflowStateForAction(action string, inv domain.Invoice) string {
	switch action {
	case "emit_proforma":
		if inv.Status == domain.InvoiceStatusProformaRefusee {
			return "proforma_refusee"
		}
		// First emit (invoice already flipped to proforma) or legacy: start at preparee then Fire → proforma.
		// Resend with existing instance uses proforma→proforma; missing instance recovers via preparee→proforma.
		return "preparee"
	case "validate_client", "reject_client":
		return "proforma"
	default:
		return "preparee"
	}
}
