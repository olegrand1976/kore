package cra

import (
	"context"

	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
)

type FeederAdapter struct {
	feeder craports.CRAFeeder
}

func NewFeederAdapter(feeder craports.CRAFeeder) ports.CRAFeeder {
	return &FeederAdapter{feeder: feeder}
}

func (a *FeederAdapter) ProposeLines(ctx context.Context, lines []ports.ProposedLine) error {
	craLines := make([]craports.ProposedLine, len(lines))
	for i, line := range lines {
		craLines[i] = craports.ProposedLine{
			TenantID: line.TenantID,
			UserID:   line.UserID,
			Source: domain.SourceRef{
				Type: line.SourceType,
				ID:   line.SourceID.String(),
			},
			Day:      line.Day,
			Duration: line.Duration,
			Comment:  line.Comment,
		}
	}
	return a.feeder.ProposeLines(ctx, craLines)
}
