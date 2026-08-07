package budget

import (
	"context"
	"fmt"

	budgetports "github.com/kore/kore/internal/modules/budget/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type HomeReader struct {
	budgets budgetports.BudgetService
	org     orgports.OrganizationService
}

func NewHomeReader(budgets budgetports.BudgetService, org orgports.OrganizationService) reportports.HomeBudgetReader {
	return &HomeReader{budgets: budgets, org: org}
}

func (r *HomeReader) ListHomeBudgets(ctx context.Context, tenant kernel.TenantID) ([]reportports.HomeBudgetRow, error) {
	items, err := r.budgets.List(ctx, tenant)
	if err != nil {
		return nil, err
	}
	labels := map[string]string{}
	if r.org != nil {
		apps, appErr := r.org.ListApplications(ctx, tenant, orgports.ApplicationListFilter{})
		if appErr == nil {
			for _, app := range apps {
				labels[app.ID.String()] = app.Libelle
			}
		}
	}
	out := make([]reportports.HomeBudgetRow, 0, len(items))
	for _, b := range items {
		label := labels[b.ApplicationID.String()]
		if label == "" {
			id := b.ID.String()
			if len(id) > 8 {
				id = id[:8]
			}
			label = fmt.Sprintf("budget · %s", id)
		}
		out = append(out, reportports.HomeBudgetRow{
			ID:           b.ID.String(),
			Label:        label,
			PlannedDays:  b.Planned.Days,
			ConsumedDays: b.Consumed.Days,
		})
	}
	return out, nil
}

var _ reportports.HomeBudgetReader = (*HomeReader)(nil)
