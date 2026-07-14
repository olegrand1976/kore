package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateTicketCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	Subject       string
	Description   string
	Priority      string
	DueAt         *time.Time
	ReporterID    *uuid.UUID
}

type AddReplyCommand struct {
	TenantID kernel.TenantID
	TicketID uuid.UUID
	AuthorID uuid.UUID
	Content  string
}

type SupportService interface {
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.Ticket, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Ticket, error)
	Create(ctx context.Context, cmd CreateTicketCommand) (domain.Ticket, error)
	Assign(ctx context.Context, tenant kernel.TenantID, ticketID, assigneeID uuid.UUID) (domain.Ticket, error)
	TakeOver(ctx context.Context, tenant kernel.TenantID, ticketID, assigneeID uuid.UUID) (domain.Ticket, error)
	AddReply(ctx context.Context, cmd AddReplyCommand) (domain.TicketReply, error)
	Resolve(ctx context.Context, tenant kernel.TenantID, ticketID uuid.UUID) (domain.Ticket, error)
	IngestInboundEmails(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) (int, error)
}

type SupportRepository interface {
	SaveTicket(ctx context.Context, t domain.Ticket) error
	GetTicket(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Ticket, error)
	ListTickets(ctx context.Context, tenant kernel.TenantID) ([]domain.Ticket, error)
	SaveReply(ctx context.Context, reply domain.TicketReply) error
	ListReplies(ctx context.Context, tenant kernel.TenantID, ticketID uuid.UUID) ([]domain.TicketReply, error)
}
