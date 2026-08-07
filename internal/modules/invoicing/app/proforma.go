package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
)

func (s *service) EmitProforma(ctx context.Context, cmd ports.EmitProformaCommand) (domain.Invoice, error) {
	if err := s.requireEnabled(ctx, cmd.TenantID); err != nil {
		return domain.Invoice{}, err
	}
	inv, err := s.Get(ctx, cmd.TenantID, cmd.InvoiceID)
	if err != nil {
		return domain.Invoice{}, err
	}

	email := strings.TrimSpace(strings.ToLower(cmd.RecipientEmail))
	clientName := ""
	if email == "" {
		if s.clientReader == nil {
			return domain.Invoice{}, domain.ErrNoClientEmail
		}
		email, clientName, err = s.clientReader.PrimaryBillingContact(ctx, cmd.TenantID, inv.ClientID)
		if err != nil {
			return domain.Invoice{}, err
		}
	} else if s.clientReader != nil {
		_, clientName, _ = s.clientReader.PrimaryBillingContact(ctx, cmd.TenantID, inv.ClientID)
	}
	email = normalizeEmail(email)
	if email == "" {
		return domain.Invoice{}, domain.ErrNoClientEmail
	}

	token, tokenHash, err := newProformaToken()
	if err != nil {
		return domain.Invoice{}, err
	}
	now := s.now()
	if err := inv.EmitProforma(tokenHash, email, now); err != nil {
		return domain.Invoice{}, err
	}
	if err := s.repo.SaveInvoice(ctx, inv); err != nil {
		return domain.Invoice{}, err
	}

	link, err := proformaLink(cmd.PublicBaseURL, token)
	if err != nil {
		return domain.Invoice{}, err
	}
	subject := "Kore — Proforma à valider"
	if clientName != "" {
		subject = fmt.Sprintf("Kore — Proforma à valider (%s)", clientName)
	}
	body := fmt.Sprintf(
		"Bonjour,\n\nVeuillez consulter et valider la proforma ci-dessous.\n\nMontant HT : %.2f %s\nTVA : %.2f %s\nTotal TTC : %.2f %s\n\nValider la proforma :\n%s\n\nCe lien expire sous 14 jours.\n",
		float64(inv.TotalAmount)/100, inv.Currency,
		float64(inv.TaxAmount)/100, inv.Currency,
		float64(inv.TotalAmount+inv.TaxAmount)/100, inv.Currency,
		link,
	)
	if s.mailer != nil {
		// Best-effort: invoice is already in proforma state; caller can resend.
		_ = s.mailer.Send(ctx, email, subject, body)
	}
	s.fireCRAProformaTransition(ctx, inv, cmd.ActorID, "emit_proforma", uuid.Nil)
	return inv, nil
}

func (s *service) GetProformaByToken(ctx context.Context, token string) (ports.ProformaPreview, error) {
	inv, err := s.loadByProformaToken(ctx, token)
	if err != nil {
		return ports.ProformaPreview{}, err
	}
	if err := inv.CanValidateProforma(s.now()); err != nil {
		return ports.ProformaPreview{}, err
	}
	return s.toPreview(ctx, inv, false, false), nil
}

func (s *service) ValidateProformaByToken(ctx context.Context, cmd ports.ProformaDecisionCommand) (ports.ProformaPreview, error) {
	inv, err := s.loadByProformaToken(ctx, cmd.Token)
	if err != nil {
		return ports.ProformaPreview{}, err
	}
	tokenHash := inv.ProformaTokenHash
	now := s.now()
	if err := inv.ValidateProforma(now, cmd.Comment); err != nil {
		return ports.ProformaPreview{}, err
	}
	if err := s.repo.ApplyProformaDecision(ctx, tokenHash, inv); err != nil {
		return ports.ProformaPreview{}, err
	}

	emailSent := false
	recipient := normalizeEmail(inv.ProformaRecipientEmail)
	if recipient != "" && s.mailer != nil {
		// Best-effort: validation already persisted and token cleared; do not fail the client.
		if err := s.mailer.Send(ctx, recipient, "Kore — Facture", formatInvoiceEmail(inv)); err == nil {
			inv.MarkInvoiceSent(now)
			_ = s.repo.SaveInvoice(ctx, inv)
			emailSent = true
		}
	}
	preview := s.toPreview(ctx, inv, true, false)
	preview.InvoiceEmailSent = emailSent
	// Transition tracks client validation (email send remains best-effort).
	s.fireCRAProformaTransition(ctx, inv, uuid.Nil, "validate_client", uuid.Nil)
	return preview, nil
}

func (s *service) RejectProformaByToken(ctx context.Context, cmd ports.ProformaDecisionCommand) (ports.ProformaPreview, error) {
	inv, err := s.loadByProformaToken(ctx, cmd.Token)
	if err != nil {
		return ports.ProformaPreview{}, err
	}
	tokenHash := inv.ProformaTokenHash
	now := s.now()
	if err := inv.RejectProforma(now, cmd.Comment); err != nil {
		return ports.ProformaPreview{}, err
	}
	if err := s.repo.ApplyProformaDecision(ctx, tokenHash, inv); err != nil {
		return ports.ProformaPreview{}, err
	}
	s.fireCRAProformaTransition(ctx, inv, uuid.Nil, "reject_client", uuid.Nil)
	return s.toPreview(ctx, inv, false, true), nil
}

func (s *service) loadByProformaToken(ctx context.Context, token string) (domain.Invoice, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.Invoice{}, domain.ErrProformaTokenInvalid
	}
	inv, err := s.repo.GetInvoiceByProformaTokenHash(ctx, hashProformaToken(token))
	if err != nil {
		return domain.Invoice{}, domain.ErrProformaTokenInvalid
	}
	lines, err := s.repo.ListInvoiceLines(ctx, inv.TenantID, inv.ID)
	if err != nil {
		return domain.Invoice{}, err
	}
	inv.Lines = lines
	return inv, nil
}

func (s *service) toPreview(ctx context.Context, inv domain.Invoice, validated, rejected bool) ports.ProformaPreview {
	clientName := ""
	if s.clientReader != nil {
		_, clientName, _ = s.clientReader.PrimaryBillingContact(ctx, inv.TenantID, inv.ClientID)
	}
	lines := make([]ports.ProformaLinePreview, 0, len(inv.Lines))
	for _, line := range inv.Lines {
		lines = append(lines, ports.ProformaLinePreview{
			Description: line.Description,
			Quantity:    line.Quantity,
			UnitPrice:   line.UnitPrice,
			TaxRate:     line.TaxRate,
		})
	}
	return ports.ProformaPreview{
		InvoiceID:   inv.ID,
		ClientName:  clientName,
		Currency:    inv.Currency,
		TotalAmount: inv.TotalAmount,
		TaxAmount:   inv.TaxAmount,
		Status:      inv.Status,
		ExpiresAt:   inv.ProformaExpiresAt,
		Lines:       lines,
		Validated:   validated,
		Rejected:    rejected,
		Comment:     inv.ProformaClientComment,
	}
}

func formatInvoiceEmail(inv domain.Invoice) string {
	var b strings.Builder
	b.WriteString("Bonjour,\n\nVotre validation a bien été enregistrée. Voici votre facture.\n\n")
	for _, line := range inv.Lines {
		fmt.Fprintf(&b, "- %s : %.2f × %.2f %s (TVA %.1f%%)\n",
			line.Description, line.Quantity, float64(line.UnitPrice)/100, inv.Currency, line.TaxRate)
	}
	fmt.Fprintf(&b, "\nMontant HT : %.2f %s\nTVA : %.2f %s\nTotal TTC : %.2f %s\n\nCordialement,\nKore\n",
		float64(inv.TotalAmount)/100, inv.Currency,
		float64(inv.TaxAmount)/100, inv.Currency,
		float64(inv.TotalAmount+inv.TaxAmount)/100, inv.Currency,
	)
	return b.String()
}

func newProformaToken() (token, tokenHash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	return token, hashProformaToken(token), nil
}

func hashProformaToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return ""
	}
	return email
}

func proformaLink(publicBaseURL, token string) (string, error) {
	base := strings.TrimSpace(publicBaseURL)
	if base == "" {
		return "", fmt.Errorf("public base url required")
	}
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("public base url must be absolute")
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/public/proforma/" + url.PathEscape(token)
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
