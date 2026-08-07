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

type HomeDashboardQuery struct {
	TenantID         kernel.TenantID
	UserID           uuid.UUID
	IncludeCRA       bool
	IncludeLeave     bool
	IncludeTMA       bool
	IncludeBudget    bool
	IncludeBilling   bool
	CanValidateLeave bool
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

type TMASummaryStats struct {
	OpenDemands    int
	ValidatedMonth int
}

type TMADemandReader interface {
	SummaryStats(ctx context.Context, tenant kernel.TenantID, month time.Time) (TMASummaryStats, error)
	HomeStats(ctx context.Context, tenant kernel.TenantID) (HomeTMAStats, error)
}

type HomeTMAStats struct {
	Open         int
	Total        int
	StatusCounts []domain.HomeStatusCount
}

type HomeUserReader interface {
	GetCraRequis(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (bool, error)
}

type HomeCRAReader interface {
	ListRecentSummaries(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, limit int) ([]HomeCRATimesheet, error)
}

type HomeCRATimesheet struct {
	Month          string
	Status         string
	TotalMinutes   int
	WeeksSubmitted int
	WeeksTotal     int
	PrefillRatio   int
}

type HomeLeaveReader interface {
	ListStatuses(ctx context.Context, tenant kernel.TenantID, userID *uuid.UUID) ([]string, error)
}

type HomeBudgetReader interface {
	ListHomeBudgets(ctx context.Context, tenant kernel.TenantID) ([]HomeBudgetRow, error)
}

type HomeBudgetRow struct {
	ID           string
	Label        string
	PlannedDays  float64
	ConsumedDays float64
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
	GetHomeDashboard(ctx context.Context, q HomeDashboardQuery) (domain.HomeDashboard, error)
	RunReport(ctx context.Context, cmd RunReportCommand) (domain.ReportResult, error)
	GetBillingStats(ctx context.Context, q BillingStatsQuery) (domain.BillingStats, error)
	RefreshDashboardSnapshot(ctx context.Context, tenant kernel.TenantID, code string) error
}

type ReportingRepository interface {
	GetReportDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.ReportDefinition, error)
	ListReportDefinitions(ctx context.Context, tenant kernel.TenantID) ([]domain.ReportDefinition, error)
	GetDashboardSnapshot(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error)
	UpsertDashboardSnapshot(ctx context.Context, tenant kernel.TenantID, dash domain.Dashboard) error
	ListTenantIDsForSnapshotRefresh(ctx context.Context) ([]kernel.TenantID, error)
}
