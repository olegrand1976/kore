package seed

import (
	"context"

	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/pkg/kernel"
)

func (r *Runner) ensureWorkflows(ctx context.Context, tenant kernel.TenantID) error {
	defs := []domain.WorkflowDefinition{
		leaveRequestWorkflow(tenant),
		tmaIncidentWorkflow(tenant),
	}
	for _, def := range defs {
		if err := r.deps.Workflow.DefineWorkflow(ctx, def); err != nil {
			return err
		}
	}
	return nil
}

func leaveRequestWorkflow(tenant kernel.TenantID) domain.WorkflowDefinition {
	return domain.WorkflowDefinition{
		TenantID:   tenant,
		Code:       "leave.request",
		EntityType: "leave_request",
		States: []domain.State{
			{Code: "en_attente", Label: "En attente", IsInitial: true},
			{Code: "valide", Label: "Validé", IsFinal: true},
			{Code: "refuse", Label: "Refusé", IsFinal: true},
		},
		Transitions: []domain.Transition{
			{From: "en_attente", To: "valide", Action: "approve", AllowedRoles: []string{}},
			{From: "en_attente", To: "refuse", Action: "reject", AllowedRoles: []string{}},
		},
	}
}

func tmaIncidentWorkflow(tenant kernel.TenantID) domain.WorkflowDefinition {
	return domain.WorkflowDefinition{
		TenantID:   tenant,
		Code:       "tma.incident",
		EntityType: "tma_demand",
		States: []domain.State{
			{Code: "en_attente_creation", Label: "En attente création"},
			{Code: "ouverte", Label: "Ouverte", IsInitial: true},
			{Code: "affectee", Label: "Affectée"},
			{Code: "resolue", Label: "Résolue", IsFinal: true},
			{Code: "rework", Label: "Rework"},
		},
		Transitions: []domain.Transition{
			{From: "en_attente_creation", To: "ouverte", Action: "validate_creation", AllowedRoles: []string{}},
			{From: "ouverte", To: "affectee", Action: "assign", AllowedRoles: []string{}},
			{From: "affectee", To: "resolue", Action: "resolve", AllowedRoles: []string{}},
			{From: "resolue", To: "rework", Action: "reopen", AllowedRoles: []string{}},
			{From: "rework", To: "affectee", Action: "assign", AllowedRoles: []string{}},
		},
	}
}
