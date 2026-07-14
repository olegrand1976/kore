package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	maintenancepostgres "github.com/kore/kore/internal/modules/maintenance/adapters/postgres"
	maintenancedomain "github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	supportpostgres "github.com/kore/kore/internal/modules/support/adapters/postgres"
	supportdomain "github.com/kore/kore/internal/modules/support/domain"
	tmapostgres "github.com/kore/kore/internal/modules/tma/adapters/postgres"
	"github.com/kore/kore/pkg/kernel"
)

type attachmentResourceChecker struct {
	tma         *tmapostgres.Repository
	support     *supportpostgres.Repository
	maintenance *maintenancepostgres.Repository
}

func NewAttachmentResourceChecker(
	tma *tmapostgres.Repository,
	support *supportpostgres.Repository,
	maintenance *maintenancepostgres.Repository,
) ports.AttachmentResourceChecker {
	return &attachmentResourceChecker{tma: tma, support: support, maintenance: maintenance}
}

func (c *attachmentResourceChecker) Exists(ctx context.Context, tenant kernel.TenantID, resourceType string, resourceID uuid.UUID) (bool, error) {
	switch resourceType {
	case domain.ResourceTypeTmaDemand:
		_, err := c.tma.Get(ctx, tenant, resourceID)
		return mapExists(err, pgx.ErrNoRows)
	case domain.ResourceTypeSupportTicket:
		_, err := c.support.GetTicket(ctx, tenant, resourceID)
		return mapExists(err, supportdomain.ErrTicketNotFound)
	case domain.ResourceTypeMaintenanceWorkRequest:
		_, err := c.maintenance.GetWorkRequest(ctx, tenant, resourceID)
		return mapExists(err, maintenancedomain.ErrWorkRequestNotFound)
	default:
		return false, domain.ErrInvalidAttachmentTarget
	}
}

func mapExists(err error, notFound error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if errors.Is(err, notFound) {
		return false, nil
	}
	return false, err
}

var _ ports.AttachmentResourceChecker = (*attachmentResourceChecker)(nil)
