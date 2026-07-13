package workflow

import (
	"context"

	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/internal/modules/workflow/domain"
	wfports "github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type ReaderAdapter struct {
	svc wfports.WorkflowService
}

func NewReaderAdapter(svc wfports.WorkflowService) *ReaderAdapter {
	return &ReaderAdapter{svc: svc}
}

func (a *ReaderAdapter) GetInstance(ctx context.Context, tenant kernel.TenantID, id domain.InstanceID) (domain.WorkflowInstance, error) {
	return a.svc.GetInstance(ctx, tenant, id)
}

func (a *ReaderAdapter) AvailableActions(ctx context.Context, tenant kernel.TenantID, instanceID domain.InstanceID, actor authx.Identity) ([]domain.ActionCode, error) {
	return a.svc.AvailableActions(ctx, tenant, instanceID, actor)
}

var _ ports.WorkflowReader = (*ReaderAdapter)(nil)
