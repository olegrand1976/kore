package org

import (
	"context"

	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/modules/project/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ApplicationReader struct {
	org orgports.OrganizationService
}

func NewApplicationReader(org orgports.OrganizationService) ports.ApplicationReader {
	return &ApplicationReader{org: org}
}

func (r *ApplicationReader) GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (orgdomain.Application, error) {
	return r.org.GetApplication(ctx, tenant, id)
}
