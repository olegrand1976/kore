package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/kernel"
)

type ProposedMissionLine struct {
	TenantID  kernel.TenantID
	UserID    uuid.UUID
	MissionID uuid.UUID
	Month     domain.Month
	Day       time.Time
	Duration  kernel.Duration
	Comment   string
}

type CRAFeeder interface {
	ProposeLines(ctx context.Context, lines []ProposedMissionLine) error
}

type CRAFutureCleaner interface {
	RemoveFutureLines(ctx context.Context, missionID string, from time.Time) error
}
