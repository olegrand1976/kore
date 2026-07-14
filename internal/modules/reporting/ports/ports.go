package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
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

type CRABillableReader interface {
	BillableHoursForMonth(ctx context.Context, tenant kernel.TenantID, month string) (float64, error)
}

type InvoicingBillingReader interface {
	SumRealInvoicesInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) (totalAmount int64, invoiceCount int, currency string, err error)
}

type CRAPlanningReader interface {
	ListDailyActivity(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]PlanningActivityRow, error)
}

type LeavePlanningReader interface {
	ListApprovedDays(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]PlanningActivityRow, error)
}

type PlanningActivityRow struct {
	UserID       uuid.UUID
	UserPrenom   string
	UserNom      string
	Day          time.Time
	Minutes      int
	MissionID    string
	MissionLabel string
	ClientLabel  string
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
