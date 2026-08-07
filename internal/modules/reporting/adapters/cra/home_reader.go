package cra

import (
	"context"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type HomeReader struct {
	cra craports.CRAService
}

func NewHomeReader(cra craports.CRAService) reportports.HomeCRAReader {
	return &HomeReader{cra: cra}
}

func (r *HomeReader) ListRecentSummaries(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, limit int) ([]reportports.HomeCRATimesheet, error) {
	items, err := r.cra.ListTimesheetSummaries(ctx, tenant, userID, false, limit)
	if err != nil {
		return nil, err
	}
	out := make([]reportports.HomeCRATimesheet, 0, len(items))
	for _, it := range items {
		out = append(out, reportports.HomeCRATimesheet{
			Month:          string(it.Month),
			Status:         string(it.Status),
			TotalMinutes:   it.TotalMinutes,
			WeeksSubmitted: it.WeeksSubmitted,
			WeeksTotal:     it.WeeksTotal,
			PrefillRatio:   it.PrefillRatio,
		})
	}
	return out, nil
}

var _ reportports.HomeCRAReader = (*HomeReader)(nil)
