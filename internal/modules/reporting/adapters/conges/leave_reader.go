package conges

import (
	"context"
	"time"

	congesdomain "github.com/kore/kore/internal/modules/conges/domain"
	congesports "github.com/kore/kore/internal/modules/conges/ports"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type LeaveReader struct {
	leaves congesports.LeaveService
}

func NewLeaveReader(leaves congesports.LeaveService) reportports.LeavePlanningReader {
	return &LeaveReader{leaves: leaves}
}

func (r *LeaveReader) ListApprovedDays(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]reportports.PlanningActivityRow, error) {
	if r.leaves == nil {
		return nil, nil
	}
	status := congesdomain.LeaveStatusApproved
	items, err := r.leaves.List(ctx, tenant, nil, &status)
	if err != nil {
		return nil, err
	}
	var out []reportports.PlanningActivityRow
	for _, item := range items {
		for day := item.Period.From; !day.After(item.Period.To); day = day.AddDate(0, 0, 1) {
			if day.Before(period.Start) || day.After(period.End) {
				continue
			}
			out = append(out, reportports.PlanningActivityRow{
				UserID:       item.UserID,
				Day:          day,
				Minutes:      0,
				MissionLabel: "Congé",
			})
		}
	}
	return out, nil
}

var _ reportports.LeavePlanningReader = (*LeaveReader)(nil)
