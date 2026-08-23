package integrations

import (
	"context"
	"errors"

	"github.com/google/uuid"
	integrationdomain "github.com/kore/kore/internal/modules/integrations/domain"
	integrationports "github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type TaigaLinkReader struct {
	repo integrationports.TaigaRepository
}

func NewTaigaLinkReader(repo integrationports.TaigaRepository) ports.TaigaLinkReader {
	return &TaigaLinkReader{repo: repo}
}

func (r *TaigaLinkReader) IsApplicationTaigaLinked(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) (bool, error) {
	if r.repo == nil {
		return false, nil
	}
	_, err := r.repo.FindExternalLinkByKore(ctx, tenant, "application", applicationID)
	if err != nil {
		if errors.Is(err, integrationdomain.ErrExternalLinkNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
