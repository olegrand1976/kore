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

func TestOrg_SocieteLogoContentRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	societeID := uuid.New()
	require.NoError(t, repo.SaveTenant(ctx, domain.Tenant{ID: tenant.UUID(), Name: "LogoCo"}))
	require.NoError(t, repo.SaveSociete(ctx, domain.Societe{
		ID:            societeID,
		TenantID:      tenant,
		RaisonSociale: "LogoCo SAS",
		Devise:        "EUR",
		Logo:          "/api/v1/branding/logo/" + tenant.UUID().String(),
	}))

	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x01, 0x02}
	require.NoError(t, repo.SaveSocieteLogo(ctx, tenant, societeID, png, "image/png"))

	content, contentType, err := repo.GetTenantLogo(ctx, tenant)
	require.NoError(t, err)
	require.Equal(t, "image/png", contentType)
	require.Equal(t, png, content)

	other := kernel.NewTenantID(uuid.New())
	_, _, err = repo.GetTenantLogo(ctx, other)
	require.ErrorIs(t, err, domain.ErrLogoNotFound)
}
