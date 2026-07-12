package workflow

import (
	"context"

	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/internal/modules/workflow/domain"
	wfports "github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/internal/platform/authx"
)

type Adapter struct {
	svc wfports.WorkflowService
}

func NewAdapter(svc wfports.WorkflowService) ports.WorkflowService {
	return &Adapter{svc: svc}
}

func (a *Adapter) Start(ctx context.Context, cmd ports.StartWorkflowCommand) (ports.WorkflowInstance, error) {
	inst, err := a.svc.Start(ctx, wfports.StartInstanceCommand{
		TenantID:       cmd.TenantID,
		DefinitionCode: cmd.DefinitionCode,
		EntityID:       cmd.EntityID,
	})
	if err != nil {
		return ports.WorkflowInstance{}, err
	}
	return ports.WorkflowInstance{ID: inst.ID, CurrentState: string(inst.CurrentState)}, nil
}

func (a *Adapter) Fire(ctx context.Context, cmd ports.FireTransitionCommand) (ports.WorkflowInstance, error) {
	inst, err := a.svc.Fire(ctx, wfports.FireTransitionCommand{
		TenantID:   cmd.TenantID,
		InstanceID: cmd.InstanceID,
		Action:     domain.ActionCode(cmd.Action),
		Actor: authx.Identity{
			UserID: cmd.ActorID,
		},
	})
	if err != nil {
		return ports.WorkflowInstance{}, err
	}
	return ports.WorkflowInstance{ID: inst.ID, CurrentState: string(inst.CurrentState)}, nil
}
