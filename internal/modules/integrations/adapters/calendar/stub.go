package calendar

import (
	"context"
	"log/slog"

	"github.com/kore/kore/pkg/kernel"
)

type StubGateway struct {
	logger *slog.Logger
}

func NewStubGateway(logger *slog.Logger) *StubGateway {
	if logger == nil {
		logger = slog.Default()
	}
	return &StubGateway{logger: logger}
}

func (g *StubGateway) Sync(ctx context.Context, tenant kernel.TenantID, provider string) (int, error) {
	g.logger.InfoContext(ctx, "calendar sync stub",
		"tenant", tenant.String(),
		"provider", provider,
	)
	return 0, nil
}
