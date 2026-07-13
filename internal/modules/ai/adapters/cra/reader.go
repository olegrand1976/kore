package cra

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ReaderAdapter struct {
	svc craports.CRAService
}

func NewReaderAdapter(svc craports.CRAService) *ReaderAdapter {
	return &ReaderAdapter{svc: svc}
}

func (a *ReaderAdapter) GetTimesheetByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Timesheet, error) {
	return a.svc.GetByID(ctx, tenant, id)
}

func (a *ReaderAdapter) ListRecentTimesheets(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, limit int) ([]domain.Timesheet, error) {
	return a.svc.ListTimesheets(ctx, tenant, userID, false, limit)
}

var _ ports.CRAReader = (*ReaderAdapter)(nil)
