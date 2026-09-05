package tma

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/ports"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

// WorkRefLabelReader names the TMA demand a CRA line is charged to.
type WorkRefLabelReader struct {
	repo tmaports.DemandRepository
}

func NewWorkRefLabelReader(repo tmaports.DemandRepository) ports.WorkRefLabelReader {
	return &WorkRefLabelReader{repo: repo}
}

func (r *WorkRefLabelReader) WorkRefLabel(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (string, error) {
	d, err := r.repo.Get(ctx, tenant, id)
	if err != nil {
		return "", err
	}
	return d.Subject, nil
}

var _ ports.WorkRefLabelReader = (*WorkRefLabelReader)(nil)
