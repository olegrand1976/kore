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
	ID            uuid.UUID
	TenantID      kernel.TenantID
	ClientID      uuid.UUID
	Type          InvoiceType
	Status        InvoiceStatus
	Currency      string
	TotalAmount   int64
	TaxAmount     int64
	PDPReceiptID  string
	TransmittedAt *time.Time
	CreatedAt     time.Time
	Lines         []InvoiceLine
}

type InvoiceLine struct {
	ID          uuid.UUID
	TenantID    kernel.TenantID
	InvoiceID   uuid.UUID
	Description string
	Quantity    float64
	UnitPrice   int64
	TaxRate     float64
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
