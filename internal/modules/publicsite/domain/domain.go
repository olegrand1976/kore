package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrConsentRequired     = errors.New("consent required")
	ErrSlotAlreadyBooked   = errors.New("slot already booked")
	ErrSlotExpired         = errors.New("slot expired")
	ErrSlotNotFound        = errors.New("slot not found")
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrLeadNotFound        = errors.New("lead not found")
)

type LeadStatus string

const (
	LeadStatusNew       LeadStatus = "new"
	LeadStatusContacted LeadStatus = "contacted"
	LeadStatusQualified LeadStatus = "qualified"
	LeadStatusConverted LeadStatus = "converted"
	LeadStatusLost      LeadStatus = "lost"
)

type SlotStatus string

const (
	SlotStatusFree     SlotStatus = "free"
	SlotStatusReserved SlotStatus = "reserved"
	SlotStatusCanceled SlotStatus = "canceled"
)

type AppointmentStatus string

const (
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCanceled  AppointmentStatus = "canceled"
)

type MeetingChannel string

const (
	ChannelVideo MeetingChannel = "video"
	ChannelPhone MeetingChannel = "phone"
)

type Lead struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Company   string     `json:"company"`
	Size      string     `json:"size"`
	Need      string     `json:"need"`
	UTMSource string     `json:"utmSource"`
	ConsentAt time.Time  `json:"consentAt"`
	Status    LeadStatus `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
}

type BookingSlot struct {
	ID              uuid.UUID  `json:"id"`
	CommercialID    uuid.UUID  `json:"commercialId"`
	SlotStart       time.Time  `json:"slotStart"`
	SlotEnd         time.Time  `json:"slotEnd"`
	Status          SlotStatus `json:"status"`
	ExternalEventID string     `json:"externalEventId,omitempty"`
}

type Appointment struct {
	ID           uuid.UUID         `json:"id"`
	LeadID       uuid.UUID         `json:"leadId"`
	CommercialID uuid.UUID         `json:"commercialId"`
	SlotID       uuid.UUID         `json:"slotId"`
	Channel      MeetingChannel    `json:"channel"`
	Status       AppointmentStatus `json:"status"`
	CancelToken  string            `json:"cancelToken"`
	CreatedAt    time.Time         `json:"createdAt"`
}

type ModulePresentation struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Highlight   bool   `json:"highlight"`
}

func NewLead(email, company, size, need, utm string, consent bool, now time.Time) (Lead, error) {
	if !consent {
		return Lead{}, ErrConsentRequired
	}
	if email == "" {
		return Lead{}, errors.New("email required")
	}
	return Lead{
		ID:        uuid.New(),
		Email:     email,
		Company:   company,
		Size:      size,
		Need:      need,
		UTMSource: utm,
		ConsentAt: now.UTC(),
		Status:    LeadStatusNew,
		CreatedAt: now.UTC(),
	}, nil
}

func (s BookingSlot) IsBookable(now time.Time) bool {
	if s.Status != SlotStatusFree {
		return false
	}
	return s.SlotStart.After(now)
}
