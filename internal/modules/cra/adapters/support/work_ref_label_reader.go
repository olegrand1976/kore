package support

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/ports"
	supportports "github.com/kore/kore/internal/modules/support/ports"
	"github.com/kore/kore/pkg/kernel"
)

// WorkRefLabelReader names the support ticket a CRA line is charged to.
type WorkRefLabelReader struct {
	repo supportports.SupportRepository
}

func NewWorkRefLabelReader(repo supportports.SupportRepository) ports.WorkRefLabelReader {
	return &WorkRefLabelReader{repo: repo}
}

func (r *WorkRefLabelReader) WorkRefLabel(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (string, error) {
	t, err := r.repo.GetTicket(ctx, tenant, id)
	if err != nil {
		return "", err
	}
	return t.Subject, nil
}

var _ ports.WorkRefLabelReader = (*WorkRefLabelReader)(nil)
