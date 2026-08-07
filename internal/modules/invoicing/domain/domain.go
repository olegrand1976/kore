package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrInvoiceNotFound          = errors.New("invoice not found")
	ErrInvalidInvoiceState      = errors.New("invalid invoice state transition")
	ErrInvoicingDisabled        = errors.New("invoicing disabled for organisation")
	ErrAlreadyInvoiced          = errors.New("timesheet already invoiced")
	ErrZeroUnitPrice            = errors.New("unit price is zero")
	ErrNoBillableContent        = errors.New("no billable content for invoice")
	ErrInvalidInvoiceLine       = errors.New("invalid invoice line")
	ErrNoClientEmail            = errors.New("no client email for proforma")
	ErrProformaTokenInvalid     = errors.New("proforma token invalid")
	ErrProformaTokenExpired     = errors.New("proforma token expired")
	ErrProformaAlreadyValidated = errors.New("proforma already validated")
	ErrProformaAlreadyRejected  = errors.New("proforma already rejected")
	ErrProformaCommentRequired  = errors.New("proforma rejection comment required")
	ErrProformaCommentTooLong   = errors.New("proforma comment too long")
	ErrProformaConflict         = errors.New("proforma decision conflict")
)

type InvoiceStatus string

const (
	InvoiceStatusVirtuelle       InvoiceStatus = "virtuelle"
	InvoiceStatusPreparee        InvoiceStatus = "preparee"
	InvoiceStatusProforma        InvoiceStatus = "proforma"
	InvoiceStatusProformaRefusee InvoiceStatus = "proforma_refusee"
	InvoiceStatusTransmise       InvoiceStatus = "transmise"
	InvoiceStatusAcceptee        InvoiceStatus = "acceptee"
	InvoiceStatusRefusee         InvoiceStatus = "refusee"
	InvoiceStatusEncaissee       InvoiceStatus = "encaissee"
	InvoiceStatusAnnulee         InvoiceStatus = "annulee"
)

// ProformaTokenTTL is the validity window of a client validation magic link.
const ProformaTokenTTL = 14 * 24 * time.Hour

// ProformaCommentMaxLen caps client comment size.
const ProformaCommentMaxLen = 2000

type InvoiceType string

const (
	InvoiceTypeStandard   InvoiceType = "standard"
	InvoiceTypeCreditNote InvoiceType = "credit_note"
)

type Invoice struct {
	ID                     uuid.UUID       `json:"id"`
	TenantID               kernel.TenantID `json:"tenantId"`
	ClientID               uuid.UUID       `json:"clientId"`
	ClientPays             string          `json:"clientPays,omitempty"` // enriched on read (org.clients.pays), not persisted
	Type                   InvoiceType     `json:"type"`
	Status                 InvoiceStatus   `json:"status"`
	Currency               string          `json:"currency"`
	TotalAmount            int64           `json:"totalAmount"`
	TaxAmount              int64           `json:"taxAmount"`
	PDPReceiptID           string          `json:"pdpReceiptId,omitempty"`
	TransmittedAt          *time.Time      `json:"transmittedAt,omitempty"`
	CreatedAt              time.Time       `json:"createdAt"`
	SourceTimesheetID      *uuid.UUID      `json:"sourceTimesheetId,omitempty"`
	ProformaTokenHash      string          `json:"-"`
	ProformaRecipientEmail string          `json:"proformaRecipientEmail,omitempty"`
	ProformaSentAt         *time.Time      `json:"proformaSentAt,omitempty"`
	ProformaExpiresAt      *time.Time      `json:"proformaExpiresAt,omitempty"`
	ProformaValidatedAt    *time.Time      `json:"proformaValidatedAt,omitempty"`
	ProformaRejectedAt     *time.Time      `json:"proformaRejectedAt,omitempty"`
	ProformaClientComment  string          `json:"proformaClientComment,omitempty"`
	InvoiceSentAt          *time.Time      `json:"invoiceSentAt,omitempty"`
	Lines                  []InvoiceLine   `json:"lines,omitempty"`
}

type InvoiceLine struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    kernel.TenantID `json:"tenantId"`
	InvoiceID   uuid.UUID       `json:"invoiceId"`
	Description string          `json:"description"`
	Quantity    float64         `json:"quantity"`
	UnitPrice   int64           `json:"unitPrice"`
	TaxRate     float64         `json:"taxRate"`
}

type PDPQueueItem struct {
	ID          uuid.UUID
	TenantID    kernel.TenantID
	InvoiceID   uuid.UUID
	Payload     map[string]any
	Status      string
	Attempts    int
	LastError   string
	CreatedAt   time.Time
	NextRetryAt *time.Time
}

func NewInvoice(tenant kernel.TenantID, clientID uuid.UUID, invType InvoiceType, currency string) Invoice {
	if currency == "" {
		currency = "EUR"
	}
	return Invoice{
		ID:        uuid.New(),
		TenantID:  tenant,
		ClientID:  clientID,
		Type:      invType,
		Status:    InvoiceStatusVirtuelle,
		Currency:  currency,
		CreatedAt: time.Now().UTC(),
	}
}

func (i *Invoice) CanTransmit() bool {
	return i.Status == InvoiceStatusPreparee
}

func (i *Invoice) Transmit() error {
	if !i.CanTransmit() {
		return ErrInvalidInvoiceState
	}
	now := time.Now().UTC()
	i.Status = InvoiceStatusTransmise
	i.TransmittedAt = &now
	return nil
}

func (i *Invoice) CanEmitProforma() bool {
	if i.Status == InvoiceStatusProforma || i.Status == InvoiceStatusProformaRefusee {
		return true
	}
	// After client validation, status is preparee with validatedAt set — do not re-emit.
	return i.Status == InvoiceStatusPreparee && i.ProformaValidatedAt == nil
}

// EmitProforma marks the invoice as a client-facing proforma awaiting validation.
func (i *Invoice) EmitProforma(tokenHash, recipientEmail string, now time.Time) error {
	if !i.CanEmitProforma() {
		return ErrInvalidInvoiceState
	}
	if tokenHash == "" || recipientEmail == "" {
		return ErrNoClientEmail
	}
	expires := now.UTC().Add(ProformaTokenTTL)
	i.Status = InvoiceStatusProforma
	i.ProformaTokenHash = tokenHash
	i.ProformaRecipientEmail = recipientEmail
	i.ProformaSentAt = &now
	i.ProformaExpiresAt = &expires
	i.ProformaValidatedAt = nil
	i.ProformaRejectedAt = nil
	i.ProformaClientComment = ""
	i.InvoiceSentAt = nil
	return nil
}

func (i *Invoice) CanRespondProforma(now time.Time) error {
	if i.Status != InvoiceStatusProforma {
		if i.ProformaValidatedAt != nil {
			return ErrProformaAlreadyValidated
		}
		if i.ProformaRejectedAt != nil || i.Status == InvoiceStatusProformaRefusee {
			return ErrProformaAlreadyRejected
		}
		return ErrInvalidInvoiceState
	}
	if i.ProformaExpiresAt != nil && !i.ProformaExpiresAt.After(now.UTC()) {
		return ErrProformaTokenExpired
	}
	return nil
}

// ValidateProforma converts a validated proforma into a prepared invoice (email send is handled by app).
func (i *Invoice) ValidateProforma(now time.Time, comment string) error {
	if err := i.CanRespondProforma(now); err != nil {
		return err
	}
	comment = strings.TrimSpace(comment)
	if len(comment) > ProformaCommentMaxLen {
		return ErrProformaCommentTooLong
	}
	ts := now.UTC()
	i.Status = InvoiceStatusPreparee
	i.ProformaTokenHash = ""
	i.ProformaExpiresAt = nil
	i.ProformaValidatedAt = &ts
	i.ProformaRejectedAt = nil
	i.ProformaClientComment = comment
	return nil
}

// RejectProforma records a client rejection; comment is mandatory.
func (i *Invoice) RejectProforma(now time.Time, comment string) error {
	if err := i.CanRespondProforma(now); err != nil {
		return err
	}
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return ErrProformaCommentRequired
	}
	if len(comment) > ProformaCommentMaxLen {
		return ErrProformaCommentTooLong
	}
	ts := now.UTC()
	i.Status = InvoiceStatusProformaRefusee
	i.ProformaTokenHash = ""
	i.ProformaExpiresAt = nil
	i.ProformaRejectedAt = &ts
	i.ProformaValidatedAt = nil
	i.ProformaClientComment = comment
	return nil
}

// MarkInvoiceSent records that the definitive invoice email was sent to the client.
func (i *Invoice) MarkInvoiceSent(now time.Time) {
	ts := now.UTC()
	i.InvoiceSentAt = &ts
}

func (i *Invoice) CanValidateProforma(now time.Time) error {
	return i.CanRespondProforma(now)
}
