package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

type captureWorkflow struct {
	starts      []ports.StartWorkflowCommand
	fires       []ports.FireTransitionCommand
	startErr    error
	fireErr     error
	hasInstance bool
}

func (c *captureWorkflow) Start(_ context.Context, cmd ports.StartWorkflowCommand) (ports.WorkflowInstance, error) {
	c.starts = append(c.starts, cmd)
	if c.startErr != nil {
		return ports.WorkflowInstance{}, c.startErr
	}
	c.hasInstance = true
	id := uuid.Nil
	if cmd.InstanceID != nil {
		id = *cmd.InstanceID
	}
	state := "preparee"
	if cmd.InitialState != nil {
		state = *cmd.InitialState
	}
	return ports.WorkflowInstance{ID: id, CurrentState: state}, nil
}

func (c *captureWorkflow) Fire(_ context.Context, cmd ports.FireTransitionCommand) (ports.WorkflowInstance, error) {
	if !c.hasInstance {
		return ports.WorkflowInstance{}, ports.ErrWorkflowInstanceNotFound
	}
	c.fires = append(c.fires, cmd)
	if c.fireErr != nil {
		return ports.WorkflowInstance{}, c.fireErr
	}
	return ports.WorkflowInstance{ID: cmd.InstanceID, CurrentState: "proforma"}, nil
}

func TestCreateFromCRAValidation_StartsWorkflow(t *testing.T) {
	repo := &craInvoiceRepo{}
	wf := &captureWorkflow{}
	svc := NewService(repo, WithWorkflow(wf))
	tenant := kernel.NewTenantID(uuid.New())
	timesheetID := uuid.New()
	userID := uuid.New()

	inv, err := svc.CreateFromCRAValidation(context.Background(), ports.CreateFromCRACommand{
		TenantID:        tenant,
		TimesheetID:     timesheetID,
		TimesheetUserID: userID,
		ClientID:        uuid.New(),
		Month:           "2026-08",
		BillableHours:   8,
		MissionLabel:    "Mission",
		UserLabel:       "User",
		UnitPriceCents:  10000,
		Currency:        "EUR",
	})
	if err != nil {
		t.Fatalf("CreateFromCRAValidation: %v", err)
	}
	if len(wf.starts) != 1 {
		t.Fatalf("expected 1 Start, got %d", len(wf.starts))
	}
	start := wf.starts[0]
	if start.DefinitionCode != ports.DefinitionCodeCRAProforma {
		t.Fatalf("code=%s", start.DefinitionCode)
	}
	if start.RequesterID != userID {
		t.Fatalf("requester=%s", start.RequesterID)
	}
	if start.InstanceID == nil || *start.InstanceID != inv.ID {
		t.Fatalf("instanceID mismatch")
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected invoice saved after Start, got %d", len(repo.saved))
	}
}

func TestCreateFromCRAValidation_StartFailureLeavesNoInvoice(t *testing.T) {
	repo := &craInvoiceRepo{}
	wf := &captureWorkflow{startErr: errors.New("definition missing")}
	svc := NewService(repo, WithWorkflow(wf))
	tenant := kernel.NewTenantID(uuid.New())

	_, err := svc.CreateFromCRAValidation(context.Background(), ports.CreateFromCRACommand{
		TenantID:        tenant,
		TimesheetID:     uuid.New(),
		TimesheetUserID: uuid.New(),
		ClientID:        uuid.New(),
		Month:           "2026-08",
		BillableHours:   8,
		UnitPriceCents:  10000,
		Currency:        "EUR",
	})
	if err == nil {
		t.Fatal("expected Start error")
	}
	if len(repo.saved) != 0 {
		t.Fatalf("expected no invoice persisted when Start fails, got %d", len(repo.saved))
	}
}

func TestProformaFlow_FiresWorkflowTransitions(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &captureMailer{}
	wf := &captureWorkflow{}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
		WithWorkflow(wf),
	)
	tenant := kernel.NewTenantID(uuid.New())
	inv, err := svc.Create(context.Background(), ports.CreateInvoiceCommand{
		TenantID: tenant,
		ClientID: uuid.New(),
		Type:     domain.InvoiceTypeStandard,
		Currency: "EUR",
		Lines: []ports.InvoiceLineInput{{
			Description: "Prestation",
			Quantity:    1,
			UnitPrice:   10000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Simulate CRA path: instance exists with same id.
	id := inv.ID
	if _, err := wf.Start(context.Background(), ports.StartWorkflowCommand{
		TenantID:       tenant,
		DefinitionCode: ports.DefinitionCodeCRAProforma,
		EntityID:       inv.ID.String(),
		InstanceID:     &id,
	}); err != nil {
		t.Fatalf("seed Start: %v", err)
	}

	actorID := uuid.New()
	if _, err := svc.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		ActorID:       actorID,
		PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("EmitProforma: %v", err)
	}
	if len(wf.fires) != 1 || wf.fires[0].Action != "emit_proforma" {
		t.Fatalf("fires after emit: %#v", wf.fires)
	}
	if wf.fires[0].ActorID != actorID {
		t.Fatalf("actor=%s", wf.fires[0].ActorID)
	}
	token := extractProformaToken(mailer.bodies[0])
	if _, err := svc.ValidateProformaByToken(context.Background(), ports.ProformaDecisionCommand{
		Token:   token,
		Comment: "OK",
	}); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if len(wf.fires) != 2 || wf.fires[1].Action != "validate_client" {
		t.Fatalf("fires after validate: %#v", wf.fires)
	}
}

func TestProformaReject_FiresWorkflow(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &captureMailer{}
	wf := &captureWorkflow{}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
		WithWorkflow(wf),
	)
	tenant := kernel.NewTenantID(uuid.New())
	inv, err := svc.Create(context.Background(), ports.CreateInvoiceCommand{
		TenantID: tenant,
		ClientID: uuid.New(),
		Type:     domain.InvoiceTypeStandard,
		Currency: "EUR",
		Lines: []ports.InvoiceLineInput{{
			Description: "Prestation",
			Quantity:    1,
			UnitPrice:   10000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	id := inv.ID
	_, _ = wf.Start(context.Background(), ports.StartWorkflowCommand{
		TenantID: tenant, DefinitionCode: ports.DefinitionCodeCRAProforma, EntityID: inv.ID.String(), InstanceID: &id,
	})
	if _, err := svc.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID: tenant, InvoiceID: inv.ID, PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("EmitProforma: %v", err)
	}
	token := extractProformaToken(mailer.bodies[0])
	if _, err := svc.RejectProformaByToken(context.Background(), ports.ProformaDecisionCommand{
		Token:   token,
		Comment: "Montant incorrect",
	}); err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if len(wf.fires) < 2 || wf.fires[len(wf.fires)-1].Action != "reject_client" {
		t.Fatalf("fires: %#v", wf.fires)
	}
}

func TestEmitProforma_EnsuresMissingInstance(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &captureMailer{}
	wf := &captureWorkflow{} // hasInstance=false until Start
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
		WithWorkflow(wf),
	)
	tenant := kernel.NewTenantID(uuid.New())
	inv, err := svc.Create(context.Background(), ports.CreateInvoiceCommand{
		TenantID: tenant,
		ClientID: uuid.New(),
		Type:     domain.InvoiceTypeStandard,
		Currency: "EUR",
		Lines: []ports.InvoiceLineInput{{
			Description: "Prestation",
			Quantity:    1,
			UnitPrice:   10000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID: tenant, InvoiceID: inv.ID, PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("EmitProforma: %v", err)
	}
	if len(wf.starts) != 1 {
		t.Fatalf("expected ensure Start, got %d", len(wf.starts))
	}
	if wf.starts[0].InitialState == nil || *wf.starts[0].InitialState != "preparee" {
		t.Fatalf("expected ensure initial preparee, got %#v", wf.starts[0].InitialState)
	}
	if len(wf.fires) != 1 || wf.fires[0].Action != "emit_proforma" {
		t.Fatalf("fires: %#v", wf.fires)
	}
}
