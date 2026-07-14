package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

type ProposedLine struct {
	TenantID   kernel.TenantID
	UserID     uuid.UUID
	SourceType string
	SourceID   uuid.UUID
	Day        time.Time
	Duration   kernel.Duration
	Comment    string
}

type CRAFeeder interface {
	ProposeLines(ctx context.Context, lines []ProposedLine) error
}
