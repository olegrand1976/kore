package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestTicket_Lifecycle(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	reporter := uuid.New()
	ticket := domain.NewTicket(tenant, uuid.New(), "Help", "desc", kernel.PriorityNormal, nil, &reporter)
	require.Equal(t, domain.TicketStateOpen, ticket.State)
	require.Equal(t, "web", ticket.Channel)

	assignee := uuid.New()
	ticket.Assign(assignee)
	require.Equal(t, domain.TicketStateInProgress, ticket.State)
	require.Equal(t, assignee, *ticket.AssigneeID)

	require.NoError(t, ticket.Resolve())
	require.Equal(t, domain.TicketStateResolved, ticket.State)
	require.NotNil(t, ticket.ResolvedAt)
	require.ErrorIs(t, ticket.Resolve(), domain.ErrInvalidTicketState)
}

func TestNewTicketReply(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	reply := domain.NewTicketReply(tenant, uuid.New(), uuid.New(), "ok")
	require.Equal(t, "ok", reply.Content)
	require.NotEqual(t, uuid.Nil, reply.ID)
}
