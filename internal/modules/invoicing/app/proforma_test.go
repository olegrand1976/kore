package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/pkg/kernel"
)

type stubClientReader struct {
	email string
	name  string
}

func (s stubClientReader) PrimaryBillingContact(context.Context, kernel.TenantID, uuid.UUID) (string, string, error) {
	return s.email, s.name, nil
}

type captureMailer struct {
	subjects []string
	bodies   []string
}

func (m *captureMailer) Send(_ context.Context, to, subject, body string) error {
	m.subjects = append(m.subjects, subject)
	m.bodies = append(m.bodies, body)
	return nil
}

func TestEmitAndValidateProformaFlow(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &captureMailer{}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
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
	sent, err := svc.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	})
	if err != nil {
		t.Fatalf("EmitProforma: %v", err)
	}
	if sent.Status != domain.InvoiceStatusProforma {
		t.Fatalf("status=%s", sent.Status)
	}
	token := extractProformaToken(mailer.bodies[0])
	if token == "" {
		t.Fatal("token missing")
	}
	if _, err := svc.GetProformaByToken(context.Background(), token); err != nil {
		t.Fatalf("GetProformaByToken: %v", err)
	}
	validated, err := svc.ValidateProformaByToken(context.Background(), ports.ProformaDecisionCommand{
		Token:   token,
		Comment: "OK",
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !validated.Validated || validated.Status != domain.InvoiceStatusPreparee {
		t.Fatalf("unexpected %#v", validated)
	}
	if len(mailer.subjects) != 2 {
		t.Fatalf("expected invoice email, subjects=%v", mailer.subjects)
	}
	if _, err := svc.GetProformaByToken(context.Background(), token); err == nil {
		t.Fatal("token should be single-use")
	}
}

func TestEmitProformaRequiresEmail(t *testing.T) {
	repo := &virtualRepo{}
	svc := NewService(repo, WithClientContactReader(stubClientReader{}))
	tenant := kernel.NewTenantID(uuid.New())
	inv, err := svc.Create(context.Background(), ports.CreateInvoiceCommand{
		TenantID: tenant,
		ClientID: uuid.New(),
		Type:     domain.InvoiceTypeStandard,
		Lines: []ports.InvoiceLineInput{{
			Description: "Line",
			Quantity:    1,
			UnitPrice:   1000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, err = svc.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	})
	if !errors.Is(err, domain.ErrNoClientEmail) {
		t.Fatalf("expected ErrNoClientEmail, got %v", err)
	}
}

func extractProformaToken(body string) string {
	const marker = "/public/proforma/"
	i := strings.Index(body, marker)
	if i < 0 {
		return ""
	}
	rest := body[i+len(marker):]
	end := strings.IndexAny(rest, " \n\r\t")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func TestValidateProformaSucceedsWhenInvoiceMailFails(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &failAfterNMailer{failFrom: 2}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
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
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	token := extractProformaToken(mailer.bodies[0])
	validated, err := svc.ValidateProformaByToken(context.Background(), ports.ProformaDecisionCommand{
		Token: token,
	})
	if err != nil {
		t.Fatalf("Validate should succeed despite mail failure: %v", err)
	}
	if !validated.Validated || validated.Status != domain.InvoiceStatusPreparee {
		t.Fatalf("unexpected %#v", validated)
	}
	got, err := svc.Get(context.Background(), tenant, inv.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != domain.InvoiceStatusPreparee {
		t.Fatalf("status=%s", got.Status)
	}
	if got.ProformaTokenHash != "" {
		t.Fatal("token hash should be cleared")
	}
}

func TestProformaPreviewOmitsInternalIDs(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &captureMailer{}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
	)
	tenant := kernel.NewTenantID(uuid.New())
	inv, err := svc.Create(context.Background(), ports.CreateInvoiceCommand{
		TenantID: tenant,
		ClientID: uuid.New(),
		Type:     domain.InvoiceTypeStandard,
		Currency: "EUR",
		Lines: []ports.InvoiceLineInput{{
			Description: "Prestation",
			Quantity:    2,
			UnitPrice:   5000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	token := extractProformaToken(mailer.bodies[0])
	preview, err := svc.GetProformaByToken(context.Background(), token)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(preview.Lines) != 1 {
		t.Fatalf("lines=%d", len(preview.Lines))
	}
	if preview.Lines[0].Description != "Prestation" || preview.Lines[0].Quantity != 2 {
		t.Fatalf("line=%#v", preview.Lines[0])
	}
}

type failAfterNMailer struct {
	n        int
	failFrom int
	bodies   []string
}

func (m *failAfterNMailer) Send(_ context.Context, _, _, body string) error {
	m.n++
	m.bodies = append(m.bodies, body)
	if m.n >= m.failFrom {
		return errors.New("smtp down")
	}
	return nil
}

func TestRejectProformaByToken(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &captureMailer{}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
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
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	token := extractProformaToken(mailer.bodies[0])
	_, err = svc.RejectProformaByToken(context.Background(), ports.ProformaDecisionCommand{Token: token})
	if !errors.Is(err, domain.ErrProformaCommentRequired) {
		t.Fatalf("expected comment required, got %v", err)
	}
	rejected, err := svc.RejectProformaByToken(context.Background(), ports.ProformaDecisionCommand{
		Token:   token,
		Comment: "Montant trop élevé",
	})
	if err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if !rejected.Rejected || rejected.Status != domain.InvoiceStatusProformaRefusee {
		t.Fatalf("unexpected %#v", rejected)
	}
	got, err := svc.Get(context.Background(), tenant, inv.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ProformaClientComment != "Montant trop élevé" {
		t.Fatalf("comment=%q", got.ProformaClientComment)
	}
}

type failMailer struct{}

func (failMailer) Send(context.Context, string, string, string) error {
	return errors.New("smtp down")
}

func TestValidateProformaSurvivesMailFailure(t *testing.T) {
	repo := &virtualRepo{}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	good := &captureMailer{}
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(good),
		WithClock(func() time.Time { return now }),
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
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	token := extractProformaToken(good.bodies[0])
	svcFail := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(failMailer{}),
		WithClock(func() time.Time { return now }),
	)
	preview, err := svcFail.ValidateProformaByToken(context.Background(), ports.ProformaDecisionCommand{Token: token})
	if err != nil {
		t.Fatalf("Validate must succeed despite mail failure: %v", err)
	}
	if !preview.Validated || preview.InvoiceEmailSent {
		t.Fatalf("expected validated without email, got %#v", preview)
	}
	got, err := svcFail.Get(context.Background(), tenant, inv.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != domain.InvoiceStatusPreparee || got.ProformaValidatedAt == nil {
		t.Fatalf("invoice not validated: %#v", got)
	}
	if _, err := svcFail.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	}); !errors.Is(err, domain.ErrInvalidInvoiceState) {
		t.Fatalf("re-emit after validation: %v", err)
	}
}

func TestGetProformaPreviewOmitsTenantIDs(t *testing.T) {
	repo := &virtualRepo{}
	mailer := &captureMailer{}
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	svc := NewService(repo,
		WithClientContactReader(stubClientReader{email: "marie.dupont@acme.test", name: "ACME"}),
		WithMailSender(mailer),
		WithClock(func() time.Time { return now }),
	)
	tenant := kernel.NewTenantID(uuid.New())
	inv, err := svc.Create(context.Background(), ports.CreateInvoiceCommand{
		TenantID: tenant,
		ClientID: uuid.New(),
		Type:     domain.InvoiceTypeStandard,
		Currency: "EUR",
		Lines: []ports.InvoiceLineInput{{
			Description: "Prestation",
			Quantity:    2,
			UnitPrice:   5000,
			TaxRate:     20,
		}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.EmitProforma(context.Background(), ports.EmitProformaCommand{
		TenantID:      tenant,
		InvoiceID:     inv.ID,
		PublicBaseURL: "http://localhost:3001",
	}); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	token := extractProformaToken(mailer.bodies[0])
	preview, err := svc.GetProformaByToken(context.Background(), token)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(preview.Lines) != 1 || preview.Lines[0].Description != "Prestation" {
		t.Fatalf("lines=%#v", preview.Lines)
	}
}
