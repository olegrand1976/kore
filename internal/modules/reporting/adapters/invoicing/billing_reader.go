package invoicing

import (
	"context"

	invoicingports "github.com/kore/kore/internal/modules/invoicing/ports"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type BillingReader struct {
	repo invoicingports.InvoicingRepository
}

func NewBillingReader(repo invoicingports.InvoicingRepository) reportports.InvoicingBillingReader {
	return &BillingReader{repo: repo}
}

func (r *BillingReader) SumRealInvoicesInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) (int64, int, string, error) {
	return r.repo.SumNonVirtualInvoicesInPeriod(ctx, tenant, period)
}

var _ reportports.InvoicingBillingReader = (*BillingReader)(nil)
