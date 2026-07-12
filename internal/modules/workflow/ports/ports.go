package ports

import (
	"context"

	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type StartInstanceCommand struct {
	TenantID       kernel.TenantID
	DefinitionCode string
	EntityID       string
}

type FireTransitionCommand struct {
	TenantID   kernel.TenantID
	InstanceID domain.InstanceID
	Action     domain.ActionCode
	Actor      authx.Identity
}

type WorkflowService interface {
	DefineWorkflow(ctx context.Context, def domain.WorkflowDefinition) error
	Start(ctx context.Context, cmd StartInstanceCommand) (domain.WorkflowInstance, error)
	Fire(ctx context.Context, cmd FireTransitionCommand) (domain.WorkflowInstance, error)
	AvailableActions(ctx context.Context, tenant kernel.TenantID, instanceID domain.InstanceID, actor authx.Identity) ([]domain.ActionCode, error)
	History(ctx context.Context, tenant kernel.TenantID, instanceID domain.InstanceID) ([]domain.TransitionLog, error)
	GetDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.WorkflowDefinition, error)
	GetInstance(ctx context.Context, tenant kernel.TenantID, id domain.InstanceID) (domain.WorkflowInstance, error)
}

type WorkflowRepository interface {
	SaveDefinition(ctx context.Context, def domain.WorkflowDefinition) error
	GetDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.WorkflowDefinition, error)
	SaveInstance(ctx context.Context, inst domain.WorkflowInstance) error
	GetInstance(ctx context.Context, tenant kernel.TenantID, id domain.InstanceID) (domain.WorkflowInstance, error)
	AppendLog(ctx context.Context, log domain.TransitionLog) error
	ListLogs(ctx context.Context, tenant kernel.TenantID, instanceID domain.InstanceID) ([]domain.TransitionLog, error)
}

type TransitionPublisher interface {
	Publish(ctx context.Context, evt domain.TransitionOccurred) error
}

type GuardEvaluator interface {
	Evaluate(ctx context.Context, guard string, entityID string) (bool, error)
}

type NoopTransitionPublisher struct{}

func (NoopTransitionPublisher) Publish(_ context.Context, _ domain.TransitionOccurred) error {
	return nil
}

type NoopGuardEvaluator struct{}

func (NoopGuardEvaluator) Evaluate(_ context.Context, guard string, _ string) (bool, error) {
	return guard == "", nil
}
