package app_test

import (
	"context"
	"testing"
	"time"

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

type memRepo struct {
	defs map[string]domain.WorkflowDefinition
	inst map[uuid.UUID]domain.WorkflowInstance
	logs []domain.TransitionLog
}

func newMemRepo() *memRepo {
	return &memRepo{
		defs: map[string]domain.WorkflowDefinition{},
		inst: map[uuid.UUID]domain.WorkflowInstance{},
	}
}

func (m *memRepo) SaveDefinition(_ context.Context, def domain.WorkflowDefinition) error {
	m.defs[def.Code] = def
	return nil
}

func (m *memRepo) GetDefinition(_ context.Context, _ kernel.TenantID, code string) (domain.WorkflowDefinition, error) {
	def, ok := m.defs[code]
	if !ok {
		return domain.WorkflowDefinition{}, domain.ErrWorkflowNotFound
	}
	return def, nil
}

func (m *memRepo) SaveInstance(_ context.Context, inst domain.WorkflowInstance) error {
	m.inst[inst.ID] = inst
	return nil
}

func (m *memRepo) GetInstance(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.WorkflowInstance, error) {
	inst, ok := m.inst[id]
	if !ok {
		return domain.WorkflowInstance{}, domain.ErrInstanceNotFound
	}
	return inst, nil
}

func (m *memRepo) AppendLog(_ context.Context, log domain.TransitionLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *memRepo) ListLogs(_ context.Context, _ kernel.TenantID, instanceID uuid.UUID) ([]domain.TransitionLog, error) {
	var out []domain.TransitionLog
	for _, l := range m.logs {
		if l.InstanceID == instanceID {
			out = append(out, l)
		}
	}
	return out, nil
}

type capturePublisher struct {
	events []domain.TransitionOccurred
}

func (p *capturePublisher) Publish(_ context.Context, evt domain.TransitionOccurred) error {
	p.events = append(p.events, evt)
	return nil
}

func TestService_FireTransition_AppendsLogAndPublishes(t *testing.T) {
	repo := newMemRepo()
	pub := &capturePublisher{}
	tenant := kernel.NewTenantID(uuid.New())
	def := domain.WorkflowDefinition{
		ID:         uuid.New(),
		TenantID:   tenant,
		Code:       "leave.request",
		EntityType: "leave",
		States: []domain.State{
			{Code: "pending", Label: "En attente", IsInitial: true},
			{Code: "approved", Label: "Validée", IsFinal: true},
		},
		Transitions: []domain.Transition{
			{From: "pending", To: "approved", Action: "approve"},
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

	svc := app.NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"), pub)
	actor := authx.Identity{UserID: uuid.New(), Profile: authx.Profile("Chef d'équipe")}

	updated, err := svc.Fire(context.Background(), ports.FireTransitionCommand{
		TenantID:   tenant,
		InstanceID: inst.ID,
		Action:     "approve",
		Actor:      actor,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StateCode("approved"), updated.CurrentState)
	require.Len(t, repo.logs, 1)
	assert.Equal(t, domain.ActionCode("approve"), repo.logs[0].Action)
	require.Len(t, pub.events, 1)
	assert.Equal(t, domain.StateCode("approved"), pub.events[0].ToState)
}

func TestService_AvailableActions_RespectsRoles(t *testing.T) {
	repo := newMemRepo()
	tenant := kernel.NewTenantID(uuid.New())
	def := domain.WorkflowDefinition{
		ID:         uuid.New(),
		TenantID:   tenant,
		Code:       "tma.incident",
		EntityType: "demand",
		States: []domain.State{
			{Code: "open", Label: "Ouverte", IsInitial: true},
			{Code: "assigned", Label: "Affectée", IsFinal: true},
		},
		Transitions: []domain.Transition{
			{From: "open", To: "assigned", Action: "assign", AllowedRoles: []string{"Administrateur"}},
		},
	}
	require.NoError(t, repo.SaveDefinition(context.Background(), def))
	inst := domain.WorkflowInstance{
		ID:             uuid.New(),
		TenantID:       tenant,
		DefinitionCode: "tma.incident",
		EntityID:       uuid.New().String(),
		CurrentState:   "open",
	}
	require.NoError(t, repo.SaveInstance(context.Background(), inst))

	svc := app.NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"), nil)
	actions, err := svc.AvailableActions(context.Background(), tenant, inst.ID, authx.Identity{
		UserID:  uuid.New(),
		Profile: authx.Profile("Collaborateur"),
	})
	require.NoError(t, err)
	assert.Empty(t, actions)

	actions, err = svc.AvailableActions(context.Background(), tenant, inst.ID, authx.Identity{
		UserID:  uuid.New(),
		Profile: authx.Profile("Administrateur"),
	})
	require.NoError(t, err)
	assert.Equal(t, []domain.ActionCode{"assign"}, actions)
}

func TestService_Start_UsesInitialState(t *testing.T) {
	repo := newMemRepo()
	tenant := kernel.NewTenantID(uuid.New())
	def := domain.WorkflowDefinition{
		ID:         uuid.New(),
		TenantID:   tenant,
		Code:       "demo",
		EntityType: "entity",
		States: []domain.State{
			{Code: "draft", Label: "Brouillon", IsInitial: true},
			{Code: "done", Label: "Terminé", IsFinal: true},
		},
	}
	require.NoError(t, repo.SaveDefinition(context.Background(), def))

	svc := app.NewService(repo, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"), nil)
	inst, err := svc.Start(context.Background(), ports.StartInstanceCommand{
		TenantID:       tenant,
		DefinitionCode: "demo",
		EntityID:       "entity-1",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StateCode("draft"), inst.CurrentState)
	assert.WithinDuration(t, time.Now(), time.Now(), time.Second)
}
