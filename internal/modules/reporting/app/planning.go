package app

import (
	"context"

	reportingcra "github.com/kore/kore/internal/modules/reporting/adapters/cra"
	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
)

func (s *service) GetGantt(ctx context.Context, q ports.GanttQuery) (domain.GanttView, error) {
	if s.craPlanning == nil {
		return domain.GanttView{Period: q.Period, Items: []domain.GanttItem{}}, nil
	}
	rows, err := s.craPlanning.ListDailyActivity(ctx, q.TenantID, q.Period)
	if err != nil {
		return domain.GanttView{}, err
	}
	return reportingcra.BuildGanttView(q.Period, rows), nil
}

func (s *service) GetPlanning(ctx context.Context, q ports.PlanningQuery) (domain.PlanningView, error) {
	if s.craPlanning == nil {
		return domain.PlanningView{Period: q.Period, Rows: []domain.PlanningRow{}}, nil
	}
	rows, err := s.craPlanning.ListDailyActivity(ctx, q.TenantID, q.Period)
	if err != nil {
		return domain.PlanningView{}, err
	}
	if s.leavePlan != nil {
		leaveRows, err := s.leavePlan.ListApprovedDays(ctx, q.TenantID, q.Period)
		if err != nil {
			return domain.PlanningView{}, err
		}
		rows = append(rows, leaveRows...)
	}
	return reportingcra.BuildPlanningView(q.Period, rows), nil
}
