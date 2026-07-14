package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

type MissionBilling struct {
	MissionID   uuid.UUID
	ClientID    uuid.UUID
	Days        float64
	TJMAmount   int64
	Currency    string
	TotalAmount int64
}

type MissionReader interface {
	ActiveMissionDays(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID, period kernel.Period) (MissionBilling, error)
}
