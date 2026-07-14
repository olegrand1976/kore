package app

import (
	"context"
	"time"

	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo        ports.ReportingRepository
	craBillable ports.CRABillableReader
	craPlanning ports.CRAPlanningReader
}

func NewService(repo ports.ReportingRepository, craBillable ports.CRABillableReader, craPlanning ports.CRAPlanningReader) ports.ReportingService {
	return &service{repo: repo, craBillable: craBillable, craPlanning: craPlanning}
}

func (s *service) GetDashboard(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error) {
	return s.repo.GetDashboardSnapshot(ctx, tenant, code)
}

func (s *service) RunReport(ctx context.Context, cmd ports.RunReportCommand) (domain.ReportResult, error) {
	def, err := s.repo.GetReportDefinition(ctx, cmd.TenantID, cmd.ReportCode)
	if err != nil {
		return domain.ReportResult{}, err
	}
	return domain.ReportResult{
		Definition: def,
		Rows:       []map[string]any{},
	}, nil
}

func (s *service) GetBillingStats(ctx context.Context, q ports.BillingStatsQuery) (domain.BillingStats, error) {
	stats := domain.BillingStats{
		Period:   q.Period,
		Currency: "EUR",
	}
	if s.craBillable == nil {
		return stats, nil
	}
	cur := time.Date(q.Period.Start.Year(), q.Period.Start.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(q.Period.End.Year(), q.Period.End.Month(), 1, 0, 0, 0, 0, time.UTC)
	for !cur.After(endMonth) {
		hours, err := s.craBillable.BillableHoursForMonth(ctx, q.TenantID, cur.Format("2006-01"))
		if err != nil {
			return domain.BillingStats{}, err
		}
		stats.BillableHours += hours
		cur = cur.AddDate(0, 1, 0)
	}
	return stats, nil
}

var _ ports.ReportingService = (*service)(nil)
