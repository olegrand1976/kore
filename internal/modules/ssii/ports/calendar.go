package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

type WorkCalendarGateway interface {
	IsHolidayOrLeave(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, day time.Time, countryCode string) (bool, error)
}
