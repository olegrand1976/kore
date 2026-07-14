package cra

import (
	"context"

	craports "github.com/kore/kore/internal/modules/cra/ports"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type BillableReader struct {
	cra craports.CRAService
}

func NewBillableReader(cra craports.CRAService) reportports.CRABillableReader {
	return &BillableReader{cra: cra}
}

func (r *BillableReader) BillableHoursForMonth(ctx context.Context, tenant kernel.TenantID, month string) (float64, error) {
	parsed, err := cradomain.ParseMonth(month)
	if err != nil {
		return 0, err
	}
	items, err := r.cra.BillableSummary(ctx, tenant, parsed)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, item := range items {
		total += item.BillableHours
	}
	return total, nil
}

var _ reportports.CRABillableReader = (*BillableReader)(nil)
