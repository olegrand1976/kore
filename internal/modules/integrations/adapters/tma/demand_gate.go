package tma

import (
	"context"
	"errors"

	"github.com/google/uuid"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

type DemandGate struct {
	tma tmaports.TMAService
}

func NewDemandGate(tma tmaports.TMAService) *DemandGate {
	return &DemandGate{tma: tma}
}

func (g *DemandGate) KoreDemandExists(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (bool, error) {
	if g == nil || g.tma == nil {
		return true, nil
	}
	_, err := g.tma.Get(ctx, tenant, demandID)
	if err != nil {
		if errors.Is(err, tmadomain.ErrDemandNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
