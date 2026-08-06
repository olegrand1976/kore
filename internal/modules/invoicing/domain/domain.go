package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrInvalidInvoiceState = errors.New("invalid invoice state transition")
	ErrInvoicingDisabled   = errors.New("invoicing disabled for organisation")
	ErrAlreadyInvoiced     = errors.New("timesheet already invoiced")
	ErrZeroUnitPrice       = errors.New("unit price is zero")
	ErrNoBillableContent   = errors.New("no billable content for invoice")
	ErrInvalidInvoiceLine  = errors.New("invalid invoice line")
)

type InvoiceStatus string

const (
	InvoiceStatusVirtuelle InvoiceStatus = "virtuelle"
	InvoiceStatusPreparee  InvoiceStatus = "preparee"
	InvoiceStatusTransmise InvoiceStatus = "transmise"
	InvoiceStatusAcceptee  InvoiceStatus = "acceptee"
	InvoiceStatusRefusee   InvoiceStatus = "refusee"
	InvoiceStatusEncaissee InvoiceStatus = "encaissee"
	InvoiceStatusAnnulee   InvoiceStatus = "annulee"
)

type InvoiceType string

const (
	InvoiceTypeStandard   InvoiceType = "standard"
	InvoiceTypeCreditNote InvoiceType = "credit_note"
)

type Invoice struct {
	ID                uuid.UUID       `json:"id"`
	TenantID          kernel.TenantID `json:"tenantId"`
	ClientID          uuid.UUID       `json:"clientId"`
	Type              InvoiceType     `json:"type"`
	Status            InvoiceStatus   `json:"status"`
	Currency          string          `json:"currency"`
	TotalAmount       int64           `json:"totalAmount"`
	TaxAmount         int64           `json:"taxAmount"`
	PDPReceiptID      string          `json:"pdpReceiptId,omitempty"`
	TransmittedAt     *time.Time      `json:"transmittedAt,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
	SourceTimesheetID *uuid.UUID      `json:"sourceTimesheetId,omitempty"`
	Lines             []InvoiceLine   `json:"lines,omitempty"`
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
