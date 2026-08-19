package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type ArtifactValidator struct {
	pool *db.Pool
}

func NewArtifactValidator(pool *db.Pool) *ArtifactValidator {
	return &ArtifactValidator{pool: pool}
}

func (v *ArtifactValidator) EpicBelongsToApplication(ctx context.Context, tenant kernel.TenantID, appID, epicID uuid.UUID) (bool, error) {
	var exists bool
	err := v.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM project.epics
			WHERE tenant_id = $1 AND application_id = $2 AND id = $3
		)
	`, tenant.UUID(), appID, epicID).Scan(&exists)
	return exists, err
}

func (v *ArtifactValidator) SprintBelongsToApplication(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (bool, error) {
	var exists bool
	err := v.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM project.sprints
			WHERE tenant_id = $1 AND application_id = $2 AND id = $3
		)
	`, tenant.UUID(), appID, sprintID).Scan(&exists)
	return exists, err
}
