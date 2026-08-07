package app

import (
	"context"
	"errors"
	"time"

	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
)

const dashboardCacheTTL = 5 * time.Minute

type service struct {
	repo          ports.ReportingRepository
	craBillable   ports.CRABillableReader
	craPlanning   ports.CRAPlanningReader
	leavePlan     ports.LeavePlanningReader
	invoicing     ports.InvoicingBillingReader
	tmaDemands    ports.TMADemandReader
	homeUser      ports.HomeUserReader
	homeCRAReader ports.HomeCRAReader
	homeLeave     ports.HomeLeaveReader
	homeBudget    ports.HomeBudgetReader
	cache         cache.Cache
	keys          cache.KeyBuilder
}

func NewService(
	repo ports.ReportingRepository,
	craBillable ports.CRABillableReader,
	craPlanning ports.CRAPlanningReader,
	invoicing ports.InvoicingBillingReader,
	leavePlan ports.LeavePlanningReader,
	tmaDemands ports.TMADemandReader,
	opts ...ServiceOption,
) ports.ReportingService {
	s := &service{
		repo:        repo,
		craBillable: craBillable,
		craPlanning: craPlanning,
		invoicing:   invoicing,
		leavePlan:   leavePlan,
		tmaDemands:  tmaDemands,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type ServiceOption func(*service)

func WithHomeReaders(user ports.HomeUserReader, cra ports.HomeCRAReader, leave ports.HomeLeaveReader, budget ports.HomeBudgetReader) ServiceOption {
	return func(s *service) {
		s.homeUser = user
		s.homeCRAReader = cra
		s.homeLeave = leave
		s.homeBudget = budget
	}
}

func WithCache(c cache.Cache, keys cache.KeyBuilder) ServiceOption {
	return func(s *service) {
		s.cache = c
		s.keys = keys
	}
}

func (s *service) GetDashboard(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error) {
	if s.cache != nil && s.keys != nil {
		key := s.keys.Key(tenant, "reporting", code, "latest")
		var cached domain.Dashboard
		found, err := s.cache.Get(ctx, key, &cached)
		if err == nil && found {
			return cached, nil
		}
	}
	dash, err := s.loadDashboard(ctx, tenant, code)
	if err != nil {
		return domain.Dashboard{}, err
	}
	if s.cache != nil && s.keys != nil {
		_ = s.cache.Set(ctx, s.keys.Key(tenant, "reporting", code, "latest"), dash, dashboardCacheTTL)
	}
	return dash, nil
}

func (s *service) loadDashboard(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error) {
	dash, err := s.repo.GetDashboardSnapshot(ctx, tenant, code)
	if err == nil {
		return dash, nil
	}
	if !errors.Is(err, domain.ErrDashboardNotFound) {
		return domain.Dashboard{}, err
	}
	if code != "cra" {
		return domain.Dashboard{}, err
	}
	return s.synthesizeCRADashboard(ctx, tenant)
}

func (s *service) synthesizeCRADashboard(ctx context.Context, tenant kernel.TenantID) (domain.Dashboard, error) {
	stats, err := s.GetBillingStats(ctx, ports.BillingStatsQuery{
		TenantID: tenant,
		Period:   defaultDashboardPeriod(),
	})
	if err != nil {
		return domain.Dashboard{}, err
	}
	now := time.Now().UTC()
	return domain.Dashboard{
		Code:       "cra",
		Period:     defaultDashboardPeriod(),
		ComputedAt: now,
		Payload: map[string]any{
			"billableHours": stats.BillableHours,
			"totalAmount":   stats.TotalAmount,
			"invoiceCount":  stats.InvoiceCount,
			"currency":      stats.Currency,
		},
	}, nil
}

func (s *service) RefreshDashboardSnapshot(ctx context.Context, tenant kernel.TenantID, code string) error {
	var dash domain.Dashboard
	var err error
	switch code {
	case "cra":
		dash, err = s.synthesizeCRADashboard(ctx, tenant)
	default:
		dash, err = s.repo.GetDashboardSnapshot(ctx, tenant, code)
	}
	if err != nil {
		return err
	}
	dash.ComputedAt = time.Now().UTC()
	if err := s.repo.UpsertDashboardSnapshot(ctx, tenant, dash); err != nil {
		return err
	}
	if s.cache != nil && s.keys != nil {
		_ = s.cache.Delete(ctx, s.keys.Key(tenant, "reporting", code, "latest"))
	}
	return nil
}

func defaultDashboardPeriod() kernel.Period {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	period, _ := kernel.NewPeriod(start, end)
	return period
}

func (s *service) RunReport(ctx context.Context, cmd ports.RunReportCommand) (domain.ReportResult, error) {
	def, err := s.repo.GetReportDefinition(ctx, cmd.TenantID, cmd.ReportCode)
	if err != nil {
		return domain.ReportResult{}, err
	}
	rows := []map[string]any{}
	if def.Code == "tma_summary" {
		openDemands := 0
		validatedMonth := 0
		if s.tmaDemands != nil {
			stats, statsErr := s.tmaDemands.SummaryStats(ctx, cmd.TenantID, time.Now().UTC())
			if statsErr == nil {
				openDemands = stats.OpenDemands
				validatedMonth = stats.ValidatedMonth
			}
		}
		rows = []map[string]any{
			{"metric": "open_demands", "value": openDemands},
			{"metric": "validated_month", "value": validatedMonth},
		}
	}
	return domain.ReportResult{
		Definition: def,
		Rows:       rows,
	}, nil
}

func (s *service) GetBillingStats(ctx context.Context, q ports.BillingStatsQuery) (domain.BillingStats, error) {
	stats := domain.BillingStats{
		Period:   q.Period,
		Currency: "EUR",
	}
	if s.craBillable != nil {
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
	}
	if s.invoicing != nil {
		total, count, currency, err := s.invoicing.SumRealInvoicesInPeriod(ctx, q.TenantID, q.Period)
		if err != nil {
			return domain.BillingStats{}, err
		}
		stats.TotalAmount = total
		stats.InvoiceCount = count
		if currency != "" {
			stats.Currency = currency
		}
	}
	return stats, nil
}

var _ ports.ReportingService = (*service)(nil)
