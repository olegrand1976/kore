package orgadapter

import (
	"context"

	"github.com/google/uuid"
	congesports "github.com/kore/kore/internal/modules/conges/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type SocieteReader struct {
	repo orgports.OrganizationRepository
}

func NewSocieteReader(repo orgports.OrganizationRepository) *SocieteReader {
	return &SocieteReader{repo: repo}
}

func (r *SocieteReader) GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (orgdomain.Societe, error) {
	return r.repo.GetSociete(ctx, tenant, id)
}

func (r *SocieteReader) ResolveSocieteIDForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (uuid.UUID, error) {
	return r.repo.ResolveSocieteIDForUser(ctx, tenant, userID)
}

var _ congesports.OrgSocieteReader = (*SocieteReader)(nil)
