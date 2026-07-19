package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSideEffectValidateEmail(t *testing.T) {
	userID := uuid.New()
	effect := domain.SideEffect{
		Type: domain.SideEffectTypeEmail,
		Recipients: domain.EffectRecipients{
			Scope:   domain.RecipientScopeUser,
			UserIDs: []uuid.UUID{userID},
		},
		Subject: "Test {{entityId}}",
	}
	require.NoError(t, effect.Validate())
}

func TestSideEffectValidateRequiresSubjectOrBody(t *testing.T) {
	effect := domain.SideEffect{
		Type:       domain.SideEffectTypeEmail,
		Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeAll},
	}
	err := effect.Validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidDefinition)
}

func TestSideEffectValidateUserScopeRequiresIDs(t *testing.T) {
	effect := domain.SideEffect{
		Type:       domain.SideEffectTypeEmail,
		Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeUser},
		Subject:    "Hi",
	}
	err := effect.Validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidDefinition)
}

func TestValidateSideEffectsMaxLimit(t *testing.T) {
	effects := make([]domain.SideEffect, domain.MaxSideEffectsPerHook+1)
	for i := range effects {
		effects[i] = domain.SideEffect{
			Type:       domain.SideEffectTypeEmail,
			Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeAll},
			Subject:    "x",
		}
	}
	err := domain.ValidateSideEffects(effects)
	require.Error(t, err)
}

func TestWorkflowDefinitionValidateSideEffects(t *testing.T) {
	def := domain.WorkflowDefinition{
		Code:       "demo",
		EntityType: "entity",
		States: []domain.State{
			{Code: "draft", Label: "Draft", IsInitial: true},
			{Code: "done", Label: "Done", IsFinal: true, OnEnterEffects: []domain.SideEffect{
				{
					Type:       domain.SideEffectTypeEmail,
					Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeAll},
					BodyTemplate: "Done",
				},
			}},
		},
		Transitions: []domain.Transition{
			{
				From: "draft", To: "done", Action: "finish",
				OnFireEffects: []domain.SideEffect{
					{
						Type:       domain.SideEffectTypeEmail,
						Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeAll},
						Subject:    "Finished",
					},
				},
			},
		},
	}
	require.NoError(t, def.Validate())
}
