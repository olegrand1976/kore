//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/adapters/postgres"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func seedSecondApplication(
	t *testing.T,
	repo *postgres.Repository,
	pool *db.Pool,
	tenant kernel.TenantID,
	serviceID uuid.UUID,
	libelle string,
) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	appID := uuid.New()
	require.NoError(t, repo.SaveApplication(ctx, domain.Application{
		ID:         appID,
		TenantID:   tenant,
		Libelle:    libelle,
		Active:     true,
		ServiceIDs: []uuid.UUID{serviceID},
	}))
	return appID
}

func TestMergeApplications_movesDemandsAndDeactivatesSource(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	_, serviceID, refID := seedStructure(t, repo, tenant)
	absorbedID := seedSecondApplication(t, repo, pool, tenant, serviceID, "Lesson studio")

	authorID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO tma.demands (
			id, tenant_id, application_id, type, subject, description, priority,
			author_id, status, visible, consumption_active, requires_chef_gate, created_at
		) VALUES ($1, $2, $3, 'incident', 'Merge me', '', 'normal', $4, 'nouvelle', TRUE, TRUE, FALSE, NOW())
	`, uuid.New(), tenant.UUID(), absorbedID, authorID)
	require.NoError(t, err)

	merged, err := repo.MergeApplications(ctx, tenant, absorbedID, refID)
	require.NoError(t, err)
	require.Equal(t, refID, merged.ID)

	var demandCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tma.demands WHERE tenant_id = $1 AND application_id = $2
	`, tenant.UUID(), refID).Scan(&demandCount))
	require.Equal(t, 1, demandCount)

	absorbed, err := repo.GetApplication(ctx, tenant, absorbedID)
	require.NoError(t, err)
	require.False(t, absorbed.Active)
}

func TestMergeApplications_rejectsDuplicateDefaultBudgets(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	_, serviceID, refID := seedStructure(t, repo, tenant)
	absorbedID := seedSecondApplication(t, repo, pool, tenant, serviceID, "With budget")

	for _, appID := range []uuid.UUID{refID, absorbedID} {
		_, err := pool.Exec(ctx, `
			INSERT INTO budget.budgets (
				id, tenant_id, application_id, type,
				planned_days, planned_uo, planned_amount,
				consumed_days, consumed_uo, consumed_amount, currency
			) VALUES ($1, $2, $3, 'defaut', 0, 0, 0, 0, 0, 0, 'EUR')
		`, uuid.New(), tenant.UUID(), appID)
		require.NoError(t, err)
	}

	_, err := repo.MergeApplications(ctx, tenant, absorbedID, refID)
	require.ErrorIs(t, err, domain.ErrApplicationsMergeDuplicateDefaultBudget)
}

func TestMergeApplications_cleansAbsorbedShareRows(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	_, serviceID, refID := seedStructure(t, repo, tenant)
	absorbedID := seedSecondApplication(t, repo, pool, tenant, serviceID, "To clean")

	_, err := repo.MergeApplications(ctx, tenant, absorbedID, refID)
	require.NoError(t, err)

	var shareCount int
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM org.application_services WHERE tenant_id = $1 AND application_id = $2
	`, tenant.UUID(), absorbedID).Scan(&shareCount))
	require.Equal(t, 0, shareCount)
}

func TestMergeApplications_rejectsTwoActiveSprints(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	_, serviceID, refID := seedStructure(t, repo, tenant)
	absorbedID := seedSecondApplication(t, repo, pool, tenant, serviceID, "Duplicate sprint")

	now := time.Now().UTC()
	for _, appID := range []uuid.UUID{refID, absorbedID} {
		_, err := pool.Exec(ctx, `
			INSERT INTO project.sprints (
				id, tenant_id, application_id, name, goal, start_date, end_date, status, created_at
			) VALUES ($1, $2, $3, 'Sprint actif', '', CURRENT_DATE, CURRENT_DATE + 7, 'active', $4)
		`, uuid.New(), tenant.UUID(), appID, now)
		require.NoError(t, err)
	}

	_, err := repo.MergeApplications(ctx, tenant, absorbedID, refID)
	require.ErrorIs(t, err, domain.ErrApplicationsMergeActiveSprintConflict)
}
