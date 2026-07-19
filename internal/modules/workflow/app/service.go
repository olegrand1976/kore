package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
)

const definitionCacheTTL = 15 * time.Minute

type Service struct {
	repo      ports.WorkflowRepository
	cache     cache.Cache
	keys      cache.KeyBuilder
	publisher ports.TransitionPublisher
	effects   ports.SideEffectExecutor
	guards    ports.GuardEvaluator
	clock     func() time.Time
}

func NewService(
	repo ports.WorkflowRepository,
	appCache cache.Cache,
	keys cache.KeyBuilder,
	publisher ports.TransitionPublisher,
	effects ports.SideEffectExecutor,
) ports.WorkflowService {
	if publisher == nil {
		publisher = ports.NoopTransitionPublisher{}
	}
	if effects == nil {
		effects = ports.NoopSideEffectExecutor{}
	}
	return &Service{
		repo:      repo,
		cache:     appCache,
		keys:      keys,
		publisher: publisher,
		effects:   effects,
		guards:    ports.NoopGuardEvaluator{},
		clock:     time.Now,
	}
}

func (s *Service) WithGuardEvaluator(eval ports.GuardEvaluator) *Service {
	if eval != nil {
		s.guards = eval
	}
	return s
}

func (s *Service) DefineWorkflow(ctx context.Context, def domain.WorkflowDefinition) error {
	if err := def.Validate(); err != nil {
		return err
	}
	if def.ID == uuid.Nil {
		def.ID = uuid.New()
	}
	if def.Version == 0 {
		def.Version = 1
	}
	if err := s.repo.SaveDefinition(ctx, def); err != nil {
		return err
	}
	key := s.keys.Key(def.TenantID, "workflow", "def", def.Code)
	return s.cache.Delete(ctx, key)
}

func (s *Service) Start(ctx context.Context, cmd ports.StartInstanceCommand) (domain.WorkflowInstance, error) {
	def, err := s.loadDefinition(ctx, cmd.TenantID, cmd.DefinitionCode)
	if err != nil {
		return domain.WorkflowInstance{}, err
	}
	initial, err := def.InitialState()
	if err != nil {
		return domain.WorkflowInstance{}, err
	}
	if cmd.InitialState != nil {
		if !def.HasState(*cmd.InitialState) {
			return domain.WorkflowInstance{}, fmt.Errorf("%w: unknown initial state %q", domain.ErrInvalidDefinition, *cmd.InitialState)
		}
		initial = *cmd.InitialState
	}
	instanceID := uuid.New()
	if cmd.InstanceID != nil && *cmd.InstanceID != uuid.Nil {
		instanceID = *cmd.InstanceID
	}
	inst := domain.WorkflowInstance{
		ID:             instanceID,
		TenantID:       cmd.TenantID,
		DefinitionCode: cmd.DefinitionCode,
		EntityID:       cmd.EntityID,
		CurrentState:   initial,
	}
	if err := s.repo.SaveInstance(ctx, inst); err != nil {
		return domain.WorkflowInstance{}, err
	}
	if state, ok := def.FindState(initial); ok {
		s.runEffects(ctx, state.OnEnterEffects, ports.SideEffectContext{
			TenantID:       cmd.TenantID,
			InstanceID:     inst.ID,
			DefinitionCode: cmd.DefinitionCode,
			EntityID:       cmd.EntityID,
			ToState:        initial,
		})
	}
	return inst, nil
}

func (s *Service) Fire(ctx context.Context, cmd ports.FireTransitionCommand) (domain.WorkflowInstance, error) {
	inst, err := s.repo.GetInstance(ctx, cmd.TenantID, cmd.InstanceID)
	if err != nil {
		return domain.WorkflowInstance{}, err
	}
	def, err := s.loadDefinition(ctx, cmd.TenantID, inst.DefinitionCode)
	if err != nil {
		return domain.WorkflowInstance{}, err
	}
	transition, ok := def.FindTransition(inst.CurrentState, cmd.Action)
	if !ok {
		return domain.WorkflowInstance{}, domain.ErrTransitionNotAllowed
	}
	if !domain.TransitionAllowed(transition, cmd.Actor) {
		return domain.WorkflowInstance{}, domain.ErrActionNotPermitted
	}
	if transition.Guard != "" {
		ok, err := s.guards.Evaluate(ctx, transition.Guard, inst.EntityID)
		if err != nil {
			return domain.WorkflowInstance{}, err
		}
		if !ok {
			return domain.WorkflowInstance{}, domain.ErrGuardFailed
		}
	}
	from := inst.CurrentState
	inst.CurrentState = transition.To
	if err := s.repo.SaveInstance(ctx, inst); err != nil {
		return domain.WorkflowInstance{}, err
	}
	log := domain.TransitionLog{
		ID:         uuid.New(),
		TenantID:   cmd.TenantID,
		InstanceID: inst.ID,
		FromState:  from,
		ToState:    transition.To,
		Action:     cmd.Action,
		ActorID:    cmd.Actor.UserID,
		OccurredAt: s.clock().UTC().Format(time.RFC3339),
	}
	if err := s.repo.AppendLog(ctx, log); err != nil {
		return domain.WorkflowInstance{}, err
	}
	_ = s.publisher.Publish(ctx, domain.TransitionOccurred{
		TenantID:       cmd.TenantID,
		InstanceID:     inst.ID,
		DefinitionCode: inst.DefinitionCode,
		EntityID:       inst.EntityID,
		FromState:      from,
		ToState:        transition.To,
		Action:         cmd.Action,
		ActorID:        cmd.Actor.UserID,
	})
	effectCtx := ports.SideEffectContext{
		TenantID:       cmd.TenantID,
		InstanceID:     inst.ID,
		DefinitionCode: inst.DefinitionCode,
		EntityID:       inst.EntityID,
		FromState:      from,
		ToState:        transition.To,
		Action:         cmd.Action,
		ActorID:        cmd.Actor.UserID,
	}
	s.runEffects(ctx, transition.OnFireEffects, effectCtx)
	if state, ok := def.FindState(transition.To); ok {
		s.runEffects(ctx, state.OnEnterEffects, effectCtx)
	}
	return inst, nil
}

func (s *Service) runEffects(ctx context.Context, effects []domain.SideEffect, effectCtx ports.SideEffectContext) {
	if len(effects) == 0 {
		return
	}
	_ = s.effects.Execute(ctx, effects, effectCtx)
}

func (s *Service) AvailableActions(ctx context.Context, tenant kernel.TenantID, instanceID domain.InstanceID, actor authx.Identity) ([]domain.ActionCode, error) {
	inst, err := s.repo.GetInstance(ctx, tenant, instanceID)
	if err != nil {
		return nil, err
	}
	def, err := s.loadDefinition(ctx, tenant, inst.DefinitionCode)
	if err != nil {
		return nil, err
	}
	var actions []domain.ActionCode
	for _, t := range def.AvailableTransitions(inst.CurrentState) {
		if !domain.TransitionAllowed(t, actor) {
			continue
		}
		if t.Guard != "" {
			ok, err := s.guards.Evaluate(ctx, t.Guard, inst.EntityID)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}
		actions = append(actions, t.Action)
	}
	return actions, nil
}

func (s *Service) History(ctx context.Context, tenant kernel.TenantID, instanceID domain.InstanceID) ([]domain.TransitionLog, error) {
	return s.repo.ListLogs(ctx, tenant, instanceID)
}

func (s *Service) GetDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.WorkflowDefinition, error) {
	return s.loadDefinition(ctx, tenant, code)
}

func (s *Service) GetInstance(ctx context.Context, tenant kernel.TenantID, id domain.InstanceID) (domain.WorkflowInstance, error) {
	return s.repo.GetInstance(ctx, tenant, id)
}

func (s *Service) loadDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.WorkflowDefinition, error) {
	key := s.keys.Key(tenant, "workflow", "def", code)
	var def domain.WorkflowDefinition
	err := s.cache.GetOrLoad(ctx, key, definitionCacheTTL, func(ctx context.Context) (any, error) {
		d, err := s.repo.GetDefinition(ctx, tenant, code)
		if err != nil {
			return nil, err
		}
		return d, nil
	}, &def)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return domain.WorkflowDefinition{}, domain.ErrWorkflowNotFound
		}
		return domain.WorkflowDefinition{}, err
	}
	return def, nil
}
