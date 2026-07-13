package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleDefinition() domain.WorkflowDefinition {
	return domain.WorkflowDefinition{
		ID:         uuid.New(),
		TenantID:   kernel.NewTenantID(uuid.New()),
		Code:       "tma.incident",
		EntityType: "demand",
		States: []domain.State{
			{Code: "open", Label: "Ouverte", IsInitial: true},
			{Code: "resolved", Label: "Résolue", IsFinal: true},
		},
		Transitions: []domain.Transition{
			{From: "open", To: "resolved", Action: "resolve"},
		},
	}
}

func TestWorkflowDefinition_Validate(t *testing.T) {
	def := sampleDefinition()
	require.NoError(t, def.Validate())

	def.States = nil
	assert.ErrorIs(t, def.Validate(), domain.ErrInvalidDefinition)
}

func TestWorkflowDefinition_InitialState(t *testing.T) {
	def := sampleDefinition()
	state, err := def.InitialState()
	require.NoError(t, err)
	assert.Equal(t, domain.StateCode("open"), state)
}

func TestWorkflowDefinition_FindTransition(t *testing.T) {
	def := sampleDefinition()
	tr, ok := def.FindTransition("open", "resolve")
	require.True(t, ok)
	assert.Equal(t, domain.StateCode("resolved"), tr.To)

	_, ok = def.FindTransition("open", "missing")
	assert.False(t, ok)
}

func TestTransitionAllowed_Roles(t *testing.T) {
	tr := domain.Transition{AllowedRoles: []string{"Chef d'équipe"}}
	allowed := domain.TransitionAllowed(tr, authx.Identity{Profile: authx.Profile("Chef d'équipe")})
	assert.True(t, allowed)

	denied := domain.TransitionAllowed(tr, authx.Identity{Profile: authx.Profile("Collaborateur")})
	assert.False(t, denied)
}

func TestTransitionAllowed_EmptyRoles(t *testing.T) {
	tr := domain.Transition{}
	assert.True(t, domain.TransitionAllowed(tr, authx.Identity{Profile: authx.Profile("Collaborateur")}))
}
