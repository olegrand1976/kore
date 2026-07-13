package ports

import (
	"context"

	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/pkg/kernel"
)

type GanttQuery struct {
	TenantID kernel.TenantID
	Period   kernel.Period
}

type PlanningQuery struct {
	TenantID kernel.TenantID
	Period   kernel.Period
}

type RunReportCommand struct {
	TenantID   kernel.TenantID
	ReportCode string
	Params     map[string]any
}

type BillingStatsQuery struct {
	TenantID kernel.TenantID
	Period   kernel.Period
}

type ReportingService interface {
	GetGantt(ctx context.Context, q GanttQuery) (domain.GanttView, error)
	GetPlanning(ctx context.Context, q PlanningQuery) (domain.PlanningView, error)
	GetDashboard(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error)
	RunReport(ctx context.Context, cmd RunReportCommand) (domain.ReportResult, error)
	GetBillingStats(ctx context.Context, q BillingStatsQuery) (domain.BillingStats, error)
}

type ReportingRepository interface {
	GetReportDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.ReportDefinition, error)
	ListReportDefinitions(ctx context.Context, tenant kernel.TenantID) ([]domain.ReportDefinition, error)
	GetDashboardSnapshot(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error)
}
