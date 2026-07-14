package app

import (
	"context"

	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

func (s *Service) ListDailyActivityInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]ports.DailyActivityRow, error) {
	return s.repo.ListDailyActivityInPeriod(ctx, tenant, period)
}
