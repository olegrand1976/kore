package org

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ApplicationBillingReader struct {
	repo orgports.OrganizationRepository
}

func NewApplicationBillingReader(repo orgports.OrganizationRepository) ports.ApplicationBillingReader {
	return &ApplicationBillingReader{repo: repo}
}

func (r *ApplicationBillingReader) GetApplicationBilling(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) (ports.ApplicationBillingInfo, error) {
	app, err := r.repo.GetApplication(ctx, tenant, applicationID)
	if err != nil {
		return ports.ApplicationBillingInfo{}, err
	}
	return ports.ApplicationBillingInfo{
		ModeFacturation: app.ModeFacturation,
		DefaultTJMCents: app.DefaultTJMCents,
	}, nil
}
