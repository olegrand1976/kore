package org

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type UserIdentityResolver struct {
	repo orgports.OrganizationRepository
}

func NewUserIdentityResolver(repo orgports.OrganizationRepository) ports.UserIdentityResolver {
	return &UserIdentityResolver{repo: repo}
}

func (r *UserIdentityResolver) ResolveUserIdentity(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.UserIdentity, error) {
	u, err := r.repo.FindUserByID(ctx, tenant, userID)
	if err != nil {
		return ports.UserIdentity{}, err
	}
	return ports.UserIdentity{
		Login:  string(u.Login),
		Prenom: u.Prenom,
		Nom:    u.Nom,
	}, nil
}

var _ ports.UserIdentityResolver = (*UserIdentityResolver)(nil)
