package app

import (
	"context"
	"errors"
	"time"

	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo        ports.ReportingRepository
	craBillable ports.CRABillableReader
	craPlanning ports.CRAPlanningReader
	leavePlan   ports.LeavePlanningReader
	invoicing   ports.InvoicingBillingReader
}

func NewService(repo ports.ReportingRepository, craBillable ports.CRABillableReader, craPlanning ports.CRAPlanningReader, invoicing ports.InvoicingBillingReader, leavePlan ports.LeavePlanningReader) ports.ReportingService {
	return &service{repo: repo, craBillable: craBillable, craPlanning: craPlanning, invoicing: invoicing, leavePlan: leavePlan}
}

func (s *service) GetDashboard(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error) {
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
	stats, err := s.GetBillingStats(ctx, ports.BillingStatsQuery{
		TenantID: tenant,
		Period:   defaultDashboardPeriod(),
	})
	if err != nil {
		return domain.Dashboard{}, err
	}
	now := time.Now().UTC()
	period, _ := kernel.NewPeriod(
		time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC),
	)
	return domain.Dashboard{
		Code:       code,
		Period:     period,
		ComputedAt: now,
		Payload: map[string]any{
			"billableHours": stats.BillableHours,
			"totalAmount":   stats.TotalAmount,
			"invoiceCount":  stats.InvoiceCount,
			"currency":      stats.Currency,
		},
	}, nil
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
		rows = []map[string]any{
			{"metric": "open_demands", "value": 0},
			{"metric": "validated_month", "value": 0},
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
