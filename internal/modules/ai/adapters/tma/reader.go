package tma

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/internal/modules/tma/domain"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ReaderAdapter struct {
	svc tmaports.TMAService
}

func NewReaderAdapter(svc tmaports.TMAService) *ReaderAdapter {
	return &ReaderAdapter{svc: svc}
}

func (a *ReaderAdapter) GetDemand(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Demand, error) {
	return a.svc.Get(ctx, tenant, id)
}

func (a *ReaderAdapter) ListDemands(ctx context.Context, tenant kernel.TenantID, visibleOnly bool) ([]domain.Demand, error) {
	return a.svc.List(ctx, tenant, tmaports.ExportFilter{TenantID: tenant, VisibleOnly: visibleOnly})
}

func (a *ReaderAdapter) GetAnalysis(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.AnalysisDossier, error) {
	return a.svc.GetAnalysis(ctx, tenant, demandID)
}

var _ ports.TMAReader = (*ReaderAdapter)(nil)
