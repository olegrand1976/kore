package domain_test

import (
	"testing"

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
	require.NoError(t, d.Resolve())
	d.ConsumptionActive = false
	require.NoError(t, d.Reopen("rework"))
	assert.True(t, d.ConsumptionActive)
}

func TestDemand_AssignRequiresVisible(t *testing.T) {
	d := testDemand(true)
	err := d.Assign(uuid.New())
	assert.ErrorIs(t, err, domain.ErrDemandNotVisible)
}

func TestToXmlExportRow_Has17Fields(t *testing.T) {
	d := testDemand(false)
	row := domain.ToXmlExportRow(d)
	assert.NotEqual(t, uuid.Nil, row.DemandID)
	assert.NotEmpty(t, row.Type)
	assert.NotEmpty(t, row.Subject)
}
