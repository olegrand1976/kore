package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/internal/modules/support/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo   ports.SupportRepository
	feeder ports.CRAFeeder
}

func NewService(repo ports.SupportRepository, feeder ports.CRAFeeder) ports.SupportService {
	return &service{repo: repo, feeder: feeder}
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID) ([]domain.Ticket, error) {
	return s.repo.ListTickets(ctx, tenant)
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Ticket, error) {
	return s.repo.GetTicket(ctx, tenant, id)
}

func (s *service) Create(ctx context.Context, cmd ports.CreateTicketCommand) (domain.Ticket, error) {
	t := domain.NewTicket(cmd.TenantID, cmd.ApplicationID, cmd.Subject, cmd.Description, cmd.ReporterID)
	return t, s.repo.SaveTicket(ctx, t)
}

func (s *service) TakeOver(ctx context.Context, tenant kernel.TenantID, ticketID, assigneeID uuid.UUID) (domain.Ticket, error) {
	t, err := s.repo.GetTicket(ctx, tenant, ticketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	t.TakeOver(assigneeID)
	return t, s.repo.SaveTicket(ctx, t)
}

func (s *service) AddReply(ctx context.Context, cmd ports.AddReplyCommand) (domain.TicketReply, error) {
	if _, err := s.repo.GetTicket(ctx, cmd.TenantID, cmd.TicketID); err != nil {
		return domain.TicketReply{}, err
	}
	reply := domain.NewTicketReply(cmd.TenantID, cmd.TicketID, cmd.AuthorID, cmd.Content)
	return reply, s.repo.SaveReply(ctx, reply)
}

func (s *service) Resolve(ctx context.Context, tenant kernel.TenantID, ticketID uuid.UUID) (domain.Ticket, error) {
	t, err := s.repo.GetTicket(ctx, tenant, ticketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	if err := t.Resolve(); err != nil {
		return domain.Ticket{}, err
	}
	if err := s.repo.SaveTicket(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	if s.feeder != nil && t.AssigneeID != nil && t.ResolvedAt != nil {
		day := time.Date(t.ResolvedAt.Year(), t.ResolvedAt.Month(), t.ResolvedAt.Day(), 0, 0, 0, 0, time.UTC)
		_ = s.feeder.ProposeLines(ctx, []ports.ProposedLine{{
			TenantID:   tenant,
			UserID:     *t.AssigneeID,
			SourceType: "ticket",
			SourceID:   t.ID,
			Day:        day,
			Duration:   kernel.Duration{Minutes: 60},
			Comment:    t.Subject,
		}})
	}
	return t, nil
}

var _ ports.SupportService = (*service)(nil)
