package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/billing/domain"
	publicdomain "github.com/kore/kore/internal/modules/publicsite/domain"
)

type CaptureLeadCommand struct {
	Email     string
	Company   string
	Size      string
	Need      string
	UTMSource string
	Consent   bool
}

type SlotFilter struct {
	CommercialID *uuid.UUID
	From         time.Time
	To           time.Time
}

type BookCommand struct {
	LeadID       uuid.UUID
	CommercialID uuid.UUID
	SlotID       uuid.UUID
	Channel      publicdomain.MeetingChannel
	Email        string
	Name         string
}

type RescheduleCommand struct {
	Token     string
	NewSlotID uuid.UUID
}

type TransactionalAttachment struct {
	Filename    string
	ContentType string
	Content     []byte
}

type TransactionalMessage struct {
	To          []string
	Subject     string
	Body        string
	Attachments []TransactionalAttachment
}

type PricingReader interface {
	Catalog(ctx context.Context) (domain.PricingCatalog, error)
}

type TransactionalNotifier interface {
	NotifyTransactional(ctx context.Context, msg TransactionalMessage) error
}

type LeadRepository interface {
	Save(ctx context.Context, l publicdomain.Lead) error
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (publicdomain.Lead, error)
}

type BookingRepository interface {
	ListAvailableSlots(ctx context.Context, filter SlotFilter) ([]publicdomain.BookingSlot, error)
	GetSlot(ctx context.Context, id uuid.UUID) (publicdomain.BookingSlot, error)
	ReserveSlot(ctx context.Context, slotID uuid.UUID) error
	SaveAppointment(ctx context.Context, a publicdomain.Appointment) error
	GetAppointmentByToken(ctx context.Context, token string) (publicdomain.Appointment, error)
	UpdateAppointment(ctx context.Context, a publicdomain.Appointment) error
	ReleaseSlot(ctx context.Context, slotID uuid.UUID) error
}

type Clock interface {
	Now() time.Time
}

type PublicSiteService interface {
	GetPricing(ctx context.Context) (domain.PricingCatalog, error)
	ListModules(ctx context.Context) ([]publicdomain.ModulePresentation, error)
	CaptureLead(ctx context.Context, cmd CaptureLeadCommand) (publicdomain.Lead, error)
	DeleteLead(ctx context.Context, id uuid.UUID) error
	AvailableSlots(ctx context.Context, filter SlotFilter) ([]publicdomain.BookingSlot, error)
	BookAppointment(ctx context.Context, cmd BookCommand) (publicdomain.Appointment, error)
	CancelAppointment(ctx context.Context, token string) error
	Reschedule(ctx context.Context, cmd RescheduleCommand) (publicdomain.Appointment, error)
	PublishableKey() string
}
