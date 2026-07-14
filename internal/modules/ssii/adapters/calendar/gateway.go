package calendar

import (
	"context"
	"time"

	"github.com/google/uuid"
	congesdomain "github.com/kore/kore/internal/modules/conges/domain"
	congesports "github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/calendar"
	"github.com/kore/kore/pkg/kernel"
)

type Gateway struct {
	leaves congesports.LeaveRepository
}

func NewGateway(leaves congesports.LeaveRepository) ports.WorkCalendarGateway {
	return &Gateway{leaves: leaves}
}

func (g *Gateway) IsHolidayOrLeave(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, day time.Time, countryCode string) (bool, error) {
	if isPublicHoliday(day, countryCode) {
		return true, nil
	}
	requests, err := g.leaves.ListByUser(ctx, tenant, userID)
	if err != nil {
		return false, err
	}
	dayDate := truncateDay(day)
	for _, req := range requests {
		if req.Status != congesdomain.LeaveStatusApproved {
			continue
		}
		from := truncateDay(req.Period.From)
		to := truncateDay(req.Period.To)
		if !dayDate.Before(from) && !dayDate.After(to) {
			return true, nil
		}
	}
	return false, nil
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func isPublicHoliday(day time.Time, countryCode string) bool {
	return calendar.IsPublicHoliday(day, countryCode)
}
