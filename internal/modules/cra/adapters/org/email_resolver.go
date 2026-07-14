package org

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type EmailResolver struct {
	repo orgports.OrganizationRepository
}

func NewEmailResolver(repo orgports.OrganizationRepository) ports.UserEmailResolver {
	return &EmailResolver{repo: repo}
}

func (r *EmailResolver) ResolveUserEmails(ctx context.Context, tenant kernel.TenantID, userIDs []uuid.UUID) ([]string, error) {
	return r.repo.ResolveUserEmails(ctx, tenant, userIDs)
}

var _ ports.UserEmailResolver = (*EmailResolver)(nil)
