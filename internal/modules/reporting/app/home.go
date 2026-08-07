package app

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

func (s *service) GetHomeDashboard(ctx context.Context, q ports.HomeDashboardQuery) (domain.HomeDashboard, error) {
	out := domain.HomeDashboard{}
	if q.IncludeCRA {
		block, err := s.homeCRA(ctx, q)
		if err != nil {
			out.Errors.CRA = true
		} else {
			out.CRA = block
		}
	}
	if q.IncludeLeave {
		block, err := s.buildHomeLeave(ctx, q)
		if err != nil {
			out.Errors.Conges = true
		} else {
			out.Leave = block
		}
	}
	if q.IncludeTMA {
		block, err := s.buildHomeTMA(ctx, q)
		if err != nil {
			out.Errors.TMA = true
		} else {
			out.TMA = block
		}
	}
	if q.IncludeBudget {
		block, err := s.buildHomeBudget(ctx, q)
		if err != nil {
			out.Errors.Budget = true
		} else {
			out.Budget = block
		}
	}
	if q.IncludeBilling {
		block, err := s.buildHomeBilling(ctx, q)
		if err != nil {
			out.Errors.Billing = true
		} else {
			out.Billing = block
		}
	}
	return out, nil
}

func (s *service) homeCRA(ctx context.Context, q ports.HomeDashboardQuery) (*domain.HomeCRABlock, error) {
	if s.homeCRAReader == nil || s.homeUser == nil {
		return &domain.HomeCRABlock{}, nil
	}
	required, err := s.homeUser.GetCraRequis(ctx, q.TenantID, q.UserID)
	if err != nil {
		return nil, err
	}
	items, err := s.homeCRAReader.ListRecentSummaries(ctx, q.TenantID, q.UserID, 12)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	monthKey := now.Format("2006-01")
	block := &domain.HomeCRABlock{Required: required, Months: make([]domain.HomeCraMonth, 0, 6)}
	byMonth := map[string]ports.HomeCRATimesheet{}
	for _, it := range items {
		byMonth[it.Month] = it
	}
	if cur, ok := byMonth[monthKey]; ok {
		st := cur.Status
		block.CurrentStatus = &st
		ratio := cur.PrefillRatio
		block.PrefillRatio = &ratio
		block.PrefillLow = ratio < 70
		if required {
			block.Alert = isCRAIncomplete(cur)
		}
	} else if required {
		block.Alert = true
	}
	for i := 5; i >= 0; i-- {
		d := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		key := d.Format("2006-01")
		m := domain.HomeCraMonth{Key: key}
		if cur, ok := byMonth[key]; ok {
			st := cur.Status
			m.Status = &st
		}
		block.Months = append(block.Months, m)
	}
	return block, nil
}

func isCRAIncomplete(cur ports.HomeCRATimesheet) bool {
	if cur.Status == "Définitif" {
		return false
	}
	if cur.WeeksTotal > 0 && cur.WeeksSubmitted < cur.WeeksTotal {
		return true
	}
	return cur.TotalMinutes <= 0
}

func (s *service) buildHomeLeave(ctx context.Context, q ports.HomeDashboardQuery) (*domain.HomeLeaveBlock, error) {
	if s.homeLeave == nil {
		return &domain.HomeLeaveBlock{}, nil
	}
	var userID *uuid.UUID
	if !q.CanValidateLeave {
		uid := q.UserID
		userID = &uid
	}
	statuses, err := s.homeLeave.ListStatuses(ctx, q.TenantID, userID)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	pending := 0
	for _, st := range statuses {
		counts[st]++
		if st == "en_attente" {
			pending++
		}
	}
	block := &domain.HomeLeaveBlock{
		Pending:      pending,
		StatusCounts: statusCountsFromMap(counts),
	}
	if q.CanValidateLeave {
		block.PendingValidations = pending
	}
	return block, nil
}

func (s *service) buildHomeTMA(ctx context.Context, q ports.HomeDashboardQuery) (*domain.HomeTMABlock, error) {
	if s.tmaDemands == nil {
		return &domain.HomeTMABlock{}, nil
	}
	stats, err := s.tmaDemands.HomeStats(ctx, q.TenantID)
	if err != nil {
		return nil, err
	}
	return &domain.HomeTMABlock{
		Open:         stats.Open,
		Total:        stats.Total,
		StatusCounts: stats.StatusCounts,
	}, nil
}

func (s *service) buildHomeBudget(ctx context.Context, q ports.HomeDashboardQuery) (*domain.HomeBudgetBlock, error) {
	if s.homeBudget == nil {
		return &domain.HomeBudgetBlock{}, nil
	}
	rows, err := s.homeBudget.ListHomeBudgets(ctx, q.TenantID)
	if err != nil {
		return nil, err
	}
	var planned, consumed float64
	overrun := 0
	bars := make([]domain.HomeBudgetBar, 0, len(rows))
	for _, row := range rows {
		planned += row.PlannedDays
		consumed += row.ConsumedDays
		if row.PlannedDays > 0 && row.ConsumedDays > row.PlannedDays {
			overrun++
		}
		bars = append(bars, domain.HomeBudgetBar{
			Key:   row.ID,
			Label: row.Label,
			Value: float64(consumptionPctUncapped(row.ConsumedDays, row.PlannedDays)),
		})
	}
	sort.Slice(bars, func(i, j int) bool { return bars[i].Value > bars[j].Value })
	if len(bars) > 6 {
		bars = bars[:6]
	}
	return &domain.HomeBudgetBlock{
		Overrun:        overrun,
		ConsumptionPct: consumptionPctUncapped(consumed, planned),
		Bars:           bars,
	}, nil
}

func consumptionPctUncapped(consumed, planned float64) int {
	if planned <= 0 {
		if consumed > 0 {
			return 100
		}
		return 0
	}
	return int(math.Round((consumed / planned) * 100))
}

func (s *service) buildHomeBilling(ctx context.Context, q ports.HomeDashboardQuery) (*domain.HomeBillingBlock, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 59)
	period, err := kernel.NewPeriod(start, end)
	if err != nil {
		return nil, err
	}
	stats, err := s.GetBillingStats(ctx, ports.BillingStatsQuery{TenantID: q.TenantID, Period: period})
	if err != nil {
		return nil, err
	}
	return &domain.HomeBillingBlock{
		AmountCents:   stats.TotalAmount,
		InvoiceCount:  stats.InvoiceCount,
		BillableHours: stats.BillableHours,
		Currency:      stats.Currency,
	}, nil
}

func statusCountsFromMap(counts map[string]int) []domain.HomeStatusCount {
	out := make([]domain.HomeStatusCount, 0, len(counts))
	for k, v := range counts {
		out = append(out, domain.HomeStatusCount{Key: k, Value: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value > out[j].Value })
	return out
}
