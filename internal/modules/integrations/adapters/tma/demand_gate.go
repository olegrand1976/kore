package tma

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/ports"
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

func (g *DemandGate) CreateDemandFromTaiga(
	ctx context.Context,
	tenant kernel.TenantID,
	applicationID, authorID uuid.UUID,
	subject, description string,
) (uuid.UUID, error) {
	if g == nil || g.tma == nil {
		return uuid.Nil, tmadomain.ErrDemandNotFound
	}
	d, err := g.tma.CreateDemand(ctx, tmaports.CreateDemandCommand{
		TenantID:         tenant,
		ApplicationID:    applicationID,
		AuthorID:         authorID,
		Subject:          subject,
		Description:      description,
		Priority:         string(kernel.PriorityNormal),
		SkipOutboundSync: true, // pull Taiga→Kore must not trigger PushDemandToTaiga
	})
	if err != nil {
		return uuid.Nil, err
	}
	return d.ID, nil
}

func (g *DemandGate) GetDemandSummary(
	ctx context.Context,
	tenant kernel.TenantID,
	demandID uuid.UUID,
) (ports.TaigaDemandSummary, error) {
	if g == nil || g.tma == nil {
		return ports.TaigaDemandSummary{}, tmadomain.ErrDemandNotFound
	}
	d, err := g.tma.Get(ctx, tenant, demandID)
	if err != nil {
		return ports.TaigaDemandSummary{}, err
	}
	return ports.TaigaDemandSummary{
		ID:            d.ID,
		ApplicationID: d.ApplicationID,
		Subject:       d.Subject,
		Description:   d.Description,
	}, nil
}

func (g *DemandGate) ListDemandsByApplication(
	ctx context.Context,
	tenant kernel.TenantID,
	applicationID uuid.UUID,
) ([]ports.TaigaDemandSummary, error) {
	if g == nil || g.tma == nil {
		return nil, nil
	}
	appID := applicationID
	items, err := g.tma.List(ctx, tenant, tmaports.ExportFilter{
		TenantID:      tenant,
		ApplicationID: &appID,
		VisibleOnly:   false,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ports.TaigaDemandSummary, 0, len(items))
	for _, d := range items {
		out = append(out, ports.TaigaDemandSummary{
			ID:            d.ID,
			ApplicationID: d.ApplicationID,
			Subject:       d.Subject,
			Description:   d.Description,
		})
	}
	return out, nil
}

var _ ports.TaigaDemandGate = (*DemandGate)(nil)
