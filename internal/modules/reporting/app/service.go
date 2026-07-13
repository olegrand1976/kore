package app

import (
	"context"

	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo ports.ReportingRepository
}

func NewService(repo ports.ReportingRepository) ports.ReportingService {
	return &service{repo: repo}
}

func (s *service) GetGantt(ctx context.Context, q ports.GanttQuery) (domain.GanttView, error) {
	return domain.GanttView{Period: q.Period, Items: []domain.GanttItem{}}, nil
}

func (s *service) GetPlanning(ctx context.Context, q ports.PlanningQuery) (domain.PlanningView, error) {
	return domain.PlanningView{Period: q.Period, Rows: []domain.PlanningRow{}}, nil
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
	return domain.BillingStats{
		Period:       q.Period,
		TotalAmount:  0,
		InvoiceCount: 0,
		Currency:     "EUR",
	}, nil
}

var _ ports.ReportingService = (*service)(nil)
