package cra

import (
	"context"
	"time"

	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/modules/ssii/ports"
)

type FeederAdapter struct {
	feeder craports.CRAFeeder
}

func NewFeederAdapter(feeder craports.CRAFeeder) ports.CRAFeeder {
	return &FeederAdapter{feeder: feeder}
}

func (a *FeederAdapter) ProposeLines(ctx context.Context, lines []ports.ProposedMissionLine) error {
	if len(lines) == 0 {
		return nil
	}
	craLines := make([]craports.ProposedLine, len(lines))
	for i, line := range lines {
		craLines[i] = craports.ProposedLine{
			TenantID: line.TenantID,
			UserID:   line.UserID,
			Month:    line.Month,
			Source: cradomain.SourceRef{
				Type: "mission",
				ID:   line.MissionID.String(),
			},
			Day:      line.Day,
			Duration: line.Duration,
			Comment:  line.Comment,
		}
	}
	return a.feeder.ProposeLines(ctx, craLines)
}

type CleanerAdapter struct {
	cleaner craports.CRAFutureCleaner
}

func NewCleanerAdapter(cleaner craports.CRAFutureCleaner) ports.CRAFutureCleaner {
	return &CleanerAdapter{cleaner: cleaner}
}

func (a *CleanerAdapter) RemoveFutureLines(ctx context.Context, missionID string, from time.Time) error {
	return a.cleaner.RemoveFutureLines(ctx, cradomain.SourceRef{Type: "mission", ID: missionID}, from)
}
