package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

type MissionBilling struct {
	MissionID   uuid.UUID
	ClientID    uuid.UUID
	Title       string
	RateUnit    string // tjm | hourly
	Days        float64
	Hours       float64
	Quantity    float64 // Days (tjm) or Hours (hourly)
	UnitPrice   int64   // cents — TJM or hourly rate
	Currency    string
	TotalAmount int64
}

type MissionReader interface {
	ActiveMissionDays(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID, period kernel.Period) (MissionBilling, error)
}
