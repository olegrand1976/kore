package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/support/app"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/internal/modules/support/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type fakeSupportRepo struct {
	tickets map[uuid.UUID]domain.Ticket
	replies []domain.TicketReply
}

func newFakeSupportRepo() *fakeSupportRepo {
	return &fakeSupportRepo{tickets: map[uuid.UUID]domain.Ticket{}}
}

func (f *fakeSupportRepo) ListTickets(_ context.Context, _ kernel.TenantID) ([]domain.Ticket, error) {
	out := make([]domain.Ticket, 0, len(f.tickets))
	for _, t := range f.tickets {
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeSupportRepo) GetTicket(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.Ticket, error) {
	t, ok := f.tickets[id]
	if !ok {
		return domain.Ticket{}, domain.ErrTicketNotFound
	}
	return t, nil
}

func (f *fakeSupportRepo) SaveTicket(_ context.Context, t domain.Ticket) error {
	f.tickets[t.ID] = t
	return nil
}

func (f *fakeSupportRepo) SaveReply(_ context.Context, r domain.TicketReply) error {
	f.replies = append(f.replies, r)
	return nil
}

func (f *fakeSupportRepo) ListReplies(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]domain.TicketReply, error) {
	return f.replies, nil
}

type noopSupportFeeder struct{}

func (noopSupportFeeder) ProposeLines(context.Context, []ports.ProposedLine) error { return nil }

func TestSupportService_CreateAssignReplyResolve(t *testing.T) {
	repo := newFakeSupportRepo()
	svc := app.NewService(repo, noopSupportFeeder{}, nil)
	tenant := kernel.NewTenantID(uuid.New())
	ctx := context.Background()
	reporter := uuid.New()

	ticket, err := svc.Create(ctx, ports.CreateTicketCommand{
		TenantID:      tenant,
		ApplicationID: uuid.New(),
		Subject:       "Bug",
		Description:   "details",
		Priority:      "high",
		ReporterID:    &reporter,
	})
	require.NoError(t, err)
	require.Equal(t, domain.TicketStateOpen, ticket.State)

	assignee := uuid.New()
	ticket, err = svc.Assign(ctx, tenant, ticket.ID, assignee)
	require.NoError(t, err)
	require.Equal(t, domain.TicketStateInProgress, ticket.State)

	reply, err := svc.AddReply(ctx, ports.AddReplyCommand{
		TenantID: tenant,
		TicketID: ticket.ID,
		AuthorID: assignee,
		Content:  "looking into it",
	})
	require.NoError(t, err)
	require.Equal(t, "looking into it", reply.Content)

	ticket, err = svc.Resolve(ctx, tenant, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, domain.TicketStateResolved, ticket.State)
}
