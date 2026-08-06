//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/adapters/postgres"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestOrg_TenantRequestSettingsInvoicingEnabledRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	require.NoError(t, repo.SaveTenant(ctx, domain.Tenant{ID: tenant.UUID(), Name: "Invoicing Flag"}))

	_, found, err := repo.GetTenantRequestSettings(ctx, tenant)
	require.NoError(t, err)
	require.False(t, found)

	require.NoError(t, repo.SaveTenantRequestSettings(ctx, domain.TenantRequestSettings{
		TenantID:         tenant,
		ChannelsEnabled:  domain.ChannelsEnabled{TMA: true},
		GuidesEnabled:    true,
		InvoicingEnabled: true,
		UpdatedAt:        time.Now().UTC(),
	}))

	got, found, err := repo.GetTenantRequestSettings(ctx, tenant)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, got.InvoicingEnabled)
	require.True(t, got.ChannelsEnabled.TMA)
	require.True(t, got.GuidesEnabled)

	require.NoError(t, repo.SaveTenantRequestSettings(ctx, domain.TenantRequestSettings{
		TenantID:         tenant,
		ChannelsEnabled:  domain.ChannelsEnabled{TMA: true, Support: true},
		GuidesEnabled:    false,
		InvoicingEnabled: false,
		UpdatedAt:        time.Now().UTC(),
	}))

	got, found, err = repo.GetTenantRequestSettings(ctx, tenant)
	require.NoError(t, err)
	require.True(t, found)
	require.False(t, got.InvoicingEnabled)
	require.True(t, got.ChannelsEnabled.Support)
	require.False(t, got.GuidesEnabled)
}
