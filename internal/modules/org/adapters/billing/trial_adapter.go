package billing

import (
	"context"

	billingdomain "github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

// TrialAdapter adapts billing EnsureTrial to org.ports.TrialProvisioner.
type TrialAdapter struct {
	Ensure func(ctx context.Context, tenantID kernel.TenantID, seats int, modules []billingdomain.ModuleCode) error
}

func (a TrialAdapter) EnsureTrial(ctx context.Context, tenantID kernel.TenantID, seats int, modules []string) error {
	codes := make([]billingdomain.ModuleCode, 0, len(modules))
	for _, m := range modules {
		codes = append(codes, billingdomain.ModuleCode(m))
	}
	return a.Ensure(ctx, tenantID, seats, codes)
}

var _ ports.TrialProvisioner = TrialAdapter{}
