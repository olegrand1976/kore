package org

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

// MissionContactCleaner adapts SSIIRepository for org contact deletion cascade.
type MissionContactCleaner struct {
	Repo ports.SSIIRepository
}

func (c MissionContactCleaner) PurgeRemovedClientContacts(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID, removedIDs []uuid.UUID) error {
	if c.Repo == nil {
		return nil
	}
	return c.Repo.PurgeClientContactsFromMissions(ctx, tenant, clientID, removedIDs)
}
