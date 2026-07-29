package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestWorkRequest_Lifecycle(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	wr := domain.NewWorkRequest(tenant, uuid.New(), "Install patch", "desc", kernel.PriorityNormal, nil)
	require.Equal(t, domain.WorkStateCreated, wr.State)

	assignee := uuid.New()
	wr.Assign(assignee)
	require.Equal(t, domain.WorkStateAssigned, wr.State)
	require.Equal(t, assignee, *wr.AssigneeID)

	require.NoError(t, wr.Progress(1.5))
	require.Equal(t, domain.WorkStateInProgress, wr.State)
	require.Equal(t, 1.5, wr.ConsumptionDays)

	require.NoError(t, wr.Complete())
	require.Equal(t, domain.WorkStateCompleted, wr.State)
	require.NotNil(t, wr.CompletedAt)
}

func TestWorkRequest_InvalidTransitions(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	wr := domain.NewWorkRequest(tenant, uuid.New(), "x", "", kernel.PriorityNormal, nil)

	require.ErrorIs(t, wr.Progress(1), domain.ErrInvalidWorkState)
	require.ErrorIs(t, wr.Complete(), domain.ErrInvalidWorkState)

	wr.Assign(uuid.New())
	require.ErrorIs(t, wr.Complete(), domain.ErrInvalidWorkState)
}
