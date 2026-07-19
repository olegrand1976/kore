package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/workflow/app"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureEffects struct {
	calls []effectsCall
}

type effectsCall struct {
	count int
	ctx   ports.SideEffectContext
}

func (c *captureEffects) Execute(_ context.Context, effects []domain.SideEffect, effectCtx ports.SideEffectContext) error {
	c.calls = append(c.calls, effectsCall{count: len(effects), ctx: effectCtx})
	return nil
}

func TestService_FireRunsTransitionAndStateEffects(t *testing.T) {
	repo := newMemRepo()
	effects := &captureEffects{}
	tenant := kernel.NewTenantID(uuid.New())
	def := domain.WorkflowDefinition{
		ID:         uuid.New(),
		TenantID:   tenant,
		Code:       "leave.request",
		EntityType: "leave",
		States: []domain.State{
			{Code: "pending", Label: "En attente", IsInitial: true},
			{
				Code: "approved", Label: "Validée", IsFinal: true,
				OnEnterEffects: []domain.SideEffect{
					{Type: domain.SideEffectTypeEmail, Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeAll}, Subject: "Enter"},
				},
			},
		},
		Transitions: []domain.Transition{
			{
				From: "pending", To: "approved", Action: "approve",
				OnFireEffects: []domain.SideEffect{
					{Type: domain.SideEffectTypeEmail, Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeAll}, Subject: "Fire"},
				},
			},
		},
	}
	require.NoError(t, repo.SaveDefinition(context.Background(), def))

	inst := domain.WorkflowInstance{
		ID:             uuid.New(),
		TenantID:       tenant,
		DefinitionCode: "leave.request",
		EntityID:       uuid.New().String(),
		CurrentState:   "pending",
	}
	require.NoError(t, repo.SaveInstance(context.Background(), inst))

	svc := app.NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"), nil, effects)
	_, err := svc.Fire(context.Background(), ports.FireTransitionCommand{
		TenantID:   tenant,
		InstanceID: inst.ID,
		Action:     "approve",
		Actor:      authx.Identity{UserID: uuid.New()},
	})
	require.NoError(t, err)
	require.Len(t, effects.calls, 2)
	assert.Equal(t, 1, effects.calls[0].count)
	assert.Equal(t, domain.ActionCode("approve"), effects.calls[0].ctx.Action)
	assert.Equal(t, 1, effects.calls[1].count)
	assert.Equal(t, domain.StateCode("approved"), effects.calls[1].ctx.ToState)
}

func TestService_StartRunsInitialStateEnterEffects(t *testing.T) {
	repo := newMemRepo()
	effects := &captureEffects{}
	tenant := kernel.NewTenantID(uuid.New())
	def := domain.WorkflowDefinition{
		ID:         uuid.New(),
		TenantID:   tenant,
		Code:       "demo",
		EntityType: "entity",
		States: []domain.State{
			{
				Code: "draft", Label: "Brouillon", IsInitial: true,
				OnEnterEffects: []domain.SideEffect{
					{Type: domain.SideEffectTypeEmail, Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeAll}, Subject: "Start"},
				},
			},
			{Code: "done", Label: "Terminé", IsFinal: true},
		},
	}
	require.NoError(t, repo.SaveDefinition(context.Background(), def))

	svc := app.NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"), nil, effects)
	inst, err := svc.Start(context.Background(), ports.StartInstanceCommand{
		TenantID:       tenant,
		DefinitionCode: "demo",
		EntityID:       "entity-1",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StateCode("draft"), inst.CurrentState)
	require.Len(t, effects.calls, 1)
	assert.Equal(t, 1, effects.calls[0].count)
	assert.Equal(t, domain.StateCode("draft"), effects.calls[0].ctx.ToState)
}
