package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDemand(requiresChefGate bool) domain.Demand {
	return domain.NewDemand(kernel.NewTenantID(uuid.New()), uuid.New(), uuid.New(), "bug", "", kernel.PriorityNormal, nil, requiresChefGate)
}

func TestNewDemand_ChefGateInvisible(t *testing.T) {
	d := testDemand(true)
	assert.Equal(t, domain.DemandStatusAwaitingCreation, d.Status)
	assert.False(t, d.Visible)
}

func TestDemand_ValidateCreation(t *testing.T) {
	d := testDemand(true)
	require.NoError(t, d.ValidateCreation())
	assert.True(t, d.Visible)
	assert.Equal(t, domain.DemandStatusOpen, d.Status)
}

func TestDemand_ReopenReactivatesConsumption(t *testing.T) {
	d := testDemand(false)
	require.NoError(t, d.Resolve(time.Now().UTC()))
	d.ConsumptionActive = false
	require.NoError(t, d.Reopen("rework"))
	assert.True(t, d.ConsumptionActive)
	assert.Equal(t, "rework", d.ReopenReason)
}

func TestDemand_ReopenRequiresReason(t *testing.T) {
	d := testDemand(false)
	require.NoError(t, d.Resolve(time.Now().UTC()))
	assert.ErrorIs(t, d.Reopen("  "), domain.ErrReopenReasonRequired)
}

func TestDemand_ResolveClearsReopenReason(t *testing.T) {
	d := testDemand(false)
	require.NoError(t, d.Resolve(time.Now().UTC()))
	require.NoError(t, d.Reopen("motif rework"))
	require.Equal(t, "motif rework", d.ReopenReason)
	require.NoError(t, d.Resolve(time.Now().UTC()))
	assert.Empty(t, d.ReopenReason)
}

func TestDemand_IsOpen(t *testing.T) {
	cases := []struct {
		status domain.DemandStatus
		want   bool
	}{
		{domain.DemandStatusAwaitingCreation, false},
		{domain.DemandStatusOpen, true},
		{domain.DemandStatusAssigned, true},
		{domain.DemandStatusInProgress, true},
		{domain.DemandStatusRework, true},
		{domain.DemandStatusResolved, false},
	}
	for _, tc := range cases {
		d := testDemand(false)
		d.Status = tc.status
		assert.Equal(t, tc.want, d.IsOpen(), "status=%s", tc.status)
		assert.Equal(t, tc.want, domain.IsOpenStatus(tc.status), "status=%s", tc.status)
	}
}

func TestDemand_AssignRequiresVisible(t *testing.T) {
	d := testDemand(true)
	err := d.Assign(uuid.New())
	assert.ErrorIs(t, err, domain.ErrDemandNotVisible)
}

func TestDemand_TakeOverKeepsAssignee(t *testing.T) {
	d := testDemand(false)
	assignee := uuid.New()
	worker := uuid.New()
	require.NoError(t, d.Assign(assignee))
	require.NoError(t, d.TakeOver(worker))
	require.NotNil(t, d.AssigneeID)
	assert.Equal(t, assignee, *d.AssigneeID)
	require.NotNil(t, d.TakenOverByID)
	assert.Equal(t, worker, *d.TakenOverByID)
	assert.Equal(t, domain.DemandStatusInProgress, d.Status)
	assert.Equal(t, worker, *d.WorkerID())
}

func TestDemand_ValidateCreationKeepsPreAssignee(t *testing.T) {
	d := testDemand(true)
	assignee := uuid.New()
	d.AssigneeID = &assignee
	require.NoError(t, d.ValidateCreation())
	assert.Equal(t, domain.DemandStatusAssigned, d.Status)
	require.NotNil(t, d.AssigneeID)
	assert.Equal(t, assignee, *d.AssigneeID)
}

func TestDemand_SoftDelete(t *testing.T) {
	d := testDemand(false)
	at := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	require.NoError(t, d.SoftDelete(at))
	require.NotNil(t, d.DeletedAt)
	assert.True(t, d.DeletedAt.Equal(at))
	assert.ErrorIs(t, d.SoftDelete(at), domain.ErrDemandNotFound)
}
