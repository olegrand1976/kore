package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrInvalidTicketState = errors.New("invalid ticket state transition")
)

type TicketState string

const (
	TicketStateOpen       TicketState = "open"
	TicketStateInProgress TicketState = "in_progress"
	TicketStateResolved   TicketState = "resolved"
)

type Ticket struct {
	ID            uuid.UUID
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	Subject       string
	Description   string
	State         TicketState
	Channel       string
	ReporterID    *uuid.UUID
	AssigneeID    *uuid.UUID
	AnalysisNote  string
	CreatedAt     time.Time
	ResolvedAt    *time.Time
}

type TicketReply struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	TicketID  uuid.UUID
	AuthorID  uuid.UUID
	Content   string
	CreatedAt time.Time
}

func NewTicket(tenant kernel.TenantID, appID uuid.UUID, subject, description string, reporterID *uuid.UUID) Ticket {
	return Ticket{
		ID:            uuid.New(),
		TenantID:      tenant,
		ApplicationID: appID,
		Subject:       subject,
		Description:   description,
		State:         TicketStateOpen,
		Channel:       "web",
		ReporterID:    reporterID,
		CreatedAt:     time.Now().UTC(),
	}
}

func (t *Ticket) TakeOver(assigneeID uuid.UUID) {
	t.AssigneeID = &assigneeID
	t.State = TicketStateInProgress
}

func (t *Ticket) Resolve() error {
	if t.State == TicketStateResolved {
		return ErrInvalidTicketState
	}
	now := time.Now().UTC()
	t.State = TicketStateResolved
	t.ResolvedAt = &now
	return nil
}

func NewTicketReply(tenant kernel.TenantID, ticketID, authorID uuid.UUID, content string) TicketReply {
	return TicketReply{
		ID:        uuid.New(),
		TenantID:  tenant,
		TicketID:  ticketID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
}
