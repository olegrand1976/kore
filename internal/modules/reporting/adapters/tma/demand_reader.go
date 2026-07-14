package tma

import (
	"context"
	"time"

	"github.com/kore/kore/internal/modules/tma/domain"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type DemandReader struct {
	tma tmaports.TMAService
}

func NewDemandReader(tma tmaports.TMAService) reportports.TMADemandReader {
	return &DemandReader{tma: tma}
}

func (r *DemandReader) SummaryStats(ctx context.Context, tenant kernel.TenantID, month time.Time) (reportports.TMASummaryStats, error) {
	demands, err := r.tma.List(ctx, tenant, tmaports.ExportFilter{TenantID: tenant, VisibleOnly: true})
	if err != nil {
		return reportports.TMASummaryStats{}, err
	}
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	stats := reportports.TMASummaryStats{}
	for _, d := range demands {
		if d.Status != domain.DemandStatusResolved {
			stats.OpenDemands++
		}
		if d.Status == domain.DemandStatusResolved && !d.CreatedAt.Before(monthStart) && d.CreatedAt.Before(monthEnd) {
			stats.ValidatedMonth++
		}
	}
	return stats, nil
}

var _ reportports.TMADemandReader = (*DemandReader)(nil)
