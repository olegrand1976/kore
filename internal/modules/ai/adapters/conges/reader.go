package conges

import (
	"context"
	"errors"

	"github.com/google/uuid"
	congesports "github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/pkg/kernel"
)

type ReaderAdapter struct {
	svc congesports.LeaveService
}

func NewReaderAdapter(svc congesports.LeaveService) *ReaderAdapter {
	return &ReaderAdapter{svc: svc}
}

func (a *ReaderAdapter) GetLeave(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.LeaveRequest, error) {
	items, err := a.svc.List(ctx, tenant, nil, nil)
	if err != nil {
		return domain.LeaveRequest{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.LeaveRequest{}, errors.New("leave request not found")
}

func (a *ReaderAdapter) ListLeaves(ctx context.Context, tenant kernel.TenantID, status *domain.LeaveStatus) ([]domain.LeaveRequest, error) {
	return a.svc.List(ctx, tenant, nil, status)
}

func (a *ReaderAdapter) ListBalances(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveBalance, error) {
	return a.svc.Balance(ctx, tenant, userID)
}

var _ ports.LeaveReader = (*ReaderAdapter)(nil)
