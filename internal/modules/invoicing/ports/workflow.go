package ports

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

const DefinitionCodeCRAProforma = "invoicing.cra_proforma"

var ErrWorkflowInstanceNotFound = errors.New("workflow instance not found")

type StartWorkflowCommand struct {
	TenantID       kernel.TenantID
	DefinitionCode string
	EntityID       string
	InstanceID     *uuid.UUID
	RequesterID    uuid.UUID
	// InitialState optionnel : aligner une instance lazy sur l'état facture courant.
	InitialState *string
}

type FireTransitionCommand struct {
	TenantID   kernel.TenantID
	InstanceID uuid.UUID
	Action     string
	ActorID    uuid.UUID
}

type WorkflowInstance struct {
	ID           uuid.UUID
	CurrentState string
}

type WorkflowService interface {
	Start(ctx context.Context, cmd StartWorkflowCommand) (WorkflowInstance, error)
	Fire(ctx context.Context, cmd FireTransitionCommand) (WorkflowInstance, error)
}
