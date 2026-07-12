package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	billingdomain "github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/modules/publicsite/domain"
	"github.com/kore/kore/internal/modules/publicsite/ports"
	"github.com/kore/kore/internal/platform/cache"
)

const pricingCacheTTL = 5 * time.Minute

type Service struct {
	repo            ports.BookingRepository
	leads           ports.LeadRepository
	pricing         ports.PricingReader
	notifier        ports.TransactionalNotifier
	cache           cache.Cache
	keys            cache.KeyBuilder
	publishableKey  string
	clock           ports.Clock
}

type Repository interface {
	ports.LeadRepository
	ports.BookingRepository
}

func NewService(
	repo Repository,
	pricing ports.PricingReader,
	notifier ports.TransactionalNotifier,
	publishableKey string,
) *Service {
	return &Service{
		repo:           repo,
		leads:          repo,
		pricing:        pricing,
		notifier:       notifier,
		publishableKey: publishableKey,
		clock:          realClock{},
	}
}

func NewServiceWithCache(
	repo Repository,
	pricing ports.PricingReader,
	notifier ports.TransactionalNotifier,
	publishableKey string,
	appCache cache.Cache,
	keys cache.KeyBuilder,
) *Service {
	svc := NewService(repo, pricing, notifier, publishableKey)
	svc.cache = appCache
	svc.keys = keys
	return svc
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

func (s *Service) PublishableKey() string {
	return s.publishableKey
}

func (s *Service) GetPricing(ctx context.Context) (billingdomain.PricingCatalog, error) {
	if s.cache != nil && s.keys != nil {
		key := s.keys.PublicKey("publicsite", "pricing")
		var catalog billingdomain.PricingCatalog
		err := s.cache.GetOrLoad(ctx, key, pricingCacheTTL, func(ctx context.Context) (any, error) {
			return s.pricing.Catalog(ctx)
		}, &catalog)
		return catalog, err
	}
	return s.pricing.Catalog(ctx)
}

func (s *Service) ListModules(_ context.Context) ([]domain.ModulePresentation, error) {
	return []domain.ModulePresentation{
		{Code: "org", Name: "Organisation", Description: "Identité, tenant, RBAC et référentiel", Highlight: true},
		{Code: "cra", Name: "CRA", Description: "Compte-rendu d'activité et validation", Highlight: true},
		{Code: "conges", Name: "Congés", Description: "Gestion des absences et soldes", Highlight: false},
		{Code: "budget", Name: "Budget UO", Description: "Suivi budgétaire et consommation", Highlight: false},
		{Code: "tma", Name: "TMA", Description: "Maintenance applicative et incidents", Highlight: true},
		{Code: "workflow", Name: "Workflow", Description: "Moteur de validation configurable", Highlight: false},
	}, nil
}

func (s *Service) CaptureLead(ctx context.Context, cmd ports.CaptureLeadCommand) (domain.Lead, error) {
	lead, err := domain.NewLead(cmd.Email, cmd.Company, cmd.Size, cmd.Need, cmd.UTMSource, cmd.Consent, s.clock.Now())
	if err != nil {
		return domain.Lead{}, err
	}
	return lead, s.leads.Save(ctx, lead)
}

func (s *Service) DeleteLead(ctx context.Context, id uuid.UUID) error {
	return s.leads.Delete(ctx, id)
}

func (s *Service) AvailableSlots(ctx context.Context, filter ports.SlotFilter) ([]domain.BookingSlot, error) {
	now := s.clock.Now()
	if filter.From.IsZero() {
		filter.From = now
	}
	if filter.To.IsZero() {
		filter.To = now.Add(14 * 24 * time.Hour)
	}
	slots, err := s.repo.ListAvailableSlots(ctx, filter)
	if err != nil {
		return nil, err
	}
	out := make([]domain.BookingSlot, 0, len(slots))
	for _, slot := range slots {
		if slot.IsBookable(now) {
			out = append(out, slot)
		}
	}
	return out, nil
}

func (s *Service) BookAppointment(ctx context.Context, cmd ports.BookCommand) (domain.Appointment, error) {
	slot, err := s.repo.GetSlot(ctx, cmd.SlotID)
	if err != nil {
		return domain.Appointment{}, err
	}
	if !slot.IsBookable(s.clock.Now()) {
		return domain.Appointment{}, domain.ErrSlotExpired
	}
	if err := s.repo.ReserveSlot(ctx, cmd.SlotID); err != nil {
		return domain.Appointment{}, err
	}
	token, err := generateToken()
	if err != nil {
		_ = s.repo.ReleaseSlot(ctx, cmd.SlotID)
		return domain.Appointment{}, err
	}
	appt := domain.Appointment{
		ID:           uuid.New(),
		LeadID:       cmd.LeadID,
		CommercialID: cmd.CommercialID,
		SlotID:       cmd.SlotID,
		Channel:      cmd.Channel,
		Status:       domain.AppointmentStatusConfirmed,
		CancelToken:  token,
		CreatedAt:    s.clock.Now(),
	}
	if appt.CommercialID == uuid.Nil {
		appt.CommercialID = slot.CommercialID
	}
	if err := s.repo.SaveAppointment(ctx, appt); err != nil {
		_ = s.repo.ReleaseSlot(ctx, cmd.SlotID)
		return domain.Appointment{}, err
	}
	if s.notifier != nil && cmd.Email != "" {
		_ = s.sendConfirmation(ctx, cmd.Email, cmd.Name, slot, appt)
	}
	return appt, nil
}

func (s *Service) CancelAppointment(ctx context.Context, token string) error {
	appt, err := s.repo.GetAppointmentByToken(ctx, token)
	if err != nil {
		return err
	}
	appt.Status = domain.AppointmentStatusCanceled
	if err := s.repo.UpdateAppointment(ctx, appt); err != nil {
		return err
	}
	return s.repo.ReleaseSlot(ctx, appt.SlotID)
}

func (s *Service) Reschedule(ctx context.Context, cmd ports.RescheduleCommand) (domain.Appointment, error) {
	appt, err := s.repo.GetAppointmentByToken(ctx, cmd.Token)
	if err != nil {
		return domain.Appointment{}, err
	}
	newSlot, err := s.repo.GetSlot(ctx, cmd.NewSlotID)
	if err != nil {
		return domain.Appointment{}, err
	}
	if !newSlot.IsBookable(s.clock.Now()) {
		return domain.Appointment{}, domain.ErrSlotExpired
	}
	if err := s.repo.ReserveSlot(ctx, cmd.NewSlotID); err != nil {
		return domain.Appointment{}, err
	}
	oldSlotID := appt.SlotID
	appt.SlotID = cmd.NewSlotID
	appt.CommercialID = newSlot.CommercialID
	if err := s.repo.UpdateAppointment(ctx, appt); err != nil {
		_ = s.repo.ReleaseSlot(ctx, cmd.NewSlotID)
		return domain.Appointment{}, err
	}
	_ = s.repo.ReleaseSlot(ctx, oldSlotID)
	return appt, nil
}

func (s *Service) sendConfirmation(ctx context.Context, email, name string, slot domain.BookingSlot, appt domain.Appointment) error {
	ics := buildICS(name, slot, appt)
	subject := "Confirmation de votre entretien Kore"
	body := fmt.Sprintf("Bonjour %s,\n\nVotre entretien est confirmé le %s.\n\nLien d'annulation : token=%s\n",
		displayName(name, email),
		slot.SlotStart.Format("02/01/2006 15:04 MST"),
		appt.CancelToken,
	)
	return s.notifier.NotifyTransactional(ctx, ports.TransactionalMessage{
		To:      []string{email},
		Subject: subject,
		Body:    body,
		Attachments: []ports.TransactionalAttachment{{
			Filename:    "appointment.ics",
			ContentType: "text/calendar",
			Content:     []byte(ics),
		}},
	})
}

func displayName(name, email string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	return email
}

func buildICS(name string, slot domain.BookingSlot, appt domain.Appointment) string {
	return fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Kore//Public Booking//FR
BEGIN:VEVENT
UID:%s
DTSTAMP:%s
DTSTART:%s
DTEND:%s
SUMMARY:Entretien commercial Kore
DESCRIPTION:Canal: %s
END:VEVENT
END:VCALENDAR`,
		appt.ID.String(),
		slot.SlotStart.UTC().Format("20060102T150405Z"),
		slot.SlotStart.UTC().Format("20060102T150405Z"),
		slot.SlotEnd.UTC().Format("20060102T150405Z"),
		appt.Channel,
	)
}

func generateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

var _ ports.PublicSiteService = (*Service)(nil)
