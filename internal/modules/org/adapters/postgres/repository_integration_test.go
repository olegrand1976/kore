//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/adapters/postgres"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestOrg_TenantIsolationSocietes(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenantA := kernel.NewTenantID(uuid.New())
	tenantB := kernel.NewTenantID(uuid.New())
	require.NoError(t, repo.SaveTenant(ctx, domain.Tenant{ID: tenantA.UUID(), Name: "Alpha"}))
	require.NoError(t, repo.SaveTenant(ctx, domain.Tenant{ID: tenantB.UUID(), Name: "Beta"}))

	require.NoError(t, repo.SaveSociete(ctx, domain.Societe{
		ID:            uuid.New(),
		TenantID:      tenantA,
		RaisonSociale: "Alpha SAS",
		Devise:        "EUR",
		Adresse:       "1 rue de la Paix",
		Siret:         "12345678901234",
		URLTenant:     "alpha.kore.local",
	}))

	// Tenant B must not see tenant A's societes.
	listB, err := repo.ListSocietes(ctx, tenantB)
	require.NoError(t, err)
	require.Empty(t, listB, "tenant B should not read tenant A societes")

	listA, err := repo.ListSocietes(ctx, tenantA)
	require.NoError(t, err)
	require.Len(t, listA, 1)
	require.Equal(t, "Alpha SAS", listA[0].RaisonSociale)
}
