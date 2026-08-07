package tma

import (
	"context"
	"sort"
	"time"

	reportdomain "github.com/kore/kore/internal/modules/reporting/domain"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/internal/modules/tma/domain"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
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
		if domain.IsOpenStatus(d.Status) {
			stats.OpenDemands++
		}
		if d.Status == domain.DemandStatusResolved && !d.CreatedAt.Before(monthStart) && d.CreatedAt.Before(monthEnd) {
			stats.ValidatedMonth++
		}
	}
	return stats, nil
}

func (r *DemandReader) HomeStats(ctx context.Context, tenant kernel.TenantID) (reportports.HomeTMAStats, error) {
	demands, err := r.tma.List(ctx, tenant, tmaports.ExportFilter{TenantID: tenant, VisibleOnly: true})
	if err != nil {
		return reportports.HomeTMAStats{}, err
	}
	counts := map[string]int{}
	stats := reportports.HomeTMAStats{Total: len(demands)}
	for _, d := range demands {
		key := string(d.Status)
		counts[key]++
		if domain.IsOpenStatus(d.Status) {
			stats.Open++
		}
	}
	stats.StatusCounts = make([]reportdomain.HomeStatusCount, 0, len(counts))
	for k, v := range counts {
		stats.StatusCounts = append(stats.StatusCounts, reportdomain.HomeStatusCount{Key: k, Value: v})
	}
	sort.Slice(stats.StatusCounts, func(i, j int) bool {
		if stats.StatusCounts[i].Value == stats.StatusCounts[j].Value {
			return stats.StatusCounts[i].Key < stats.StatusCounts[j].Key
		}
		return stats.StatusCounts[i].Value > stats.StatusCounts[j].Value
	})
	return stats, nil
}

var _ reportports.TMADemandReader = (*DemandReader)(nil)
