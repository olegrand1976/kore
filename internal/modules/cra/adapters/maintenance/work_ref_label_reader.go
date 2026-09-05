package maintenance

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/ports"
	maintenanceports "github.com/kore/kore/internal/modules/maintenance/ports"
	"github.com/kore/kore/pkg/kernel"
)

// WorkRefLabelReader names the work request a CRA line is charged to.
type WorkRefLabelReader struct {
	repo maintenanceports.MaintenanceRepository
}

func NewWorkRefLabelReader(repo maintenanceports.MaintenanceRepository) ports.WorkRefLabelReader {
	return &WorkRefLabelReader{repo: repo}
}

func (r *WorkRefLabelReader) WorkRefLabel(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (string, error) {
	wr, err := r.repo.GetWorkRequest(ctx, tenant, id)
	if err != nil {
		return "", err
	}
	return wr.Subject, nil
}

var _ ports.WorkRefLabelReader = (*WorkRefLabelReader)(nil)
