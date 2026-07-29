//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/billing/adapters/postgres"
	"github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestBilling_SubscriptionRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	sub := domain.Subscription{
		ID:               uuid.New(),
		TenantID:         tenant,
		StripeCustomerID: "cus_test_e2e",
		Status:           domain.StatusTrial,
		Seats:            5,
	}
	require.NoError(t, repo.Save(ctx, sub))
	require.NoError(t, repo.SaveEntitlements(ctx, tenant, []domain.ModuleEntitlement{
		{TenantID: tenant, ModuleCode: domain.ModuleOrg, Enabled: true},
		{TenantID: tenant, ModuleCode: domain.ModuleCRA, Enabled: true},
	}))

	got, err := repo.GetByTenant(ctx, tenant)
	require.NoError(t, err)
	require.Equal(t, domain.StatusTrial, got.Status)
	require.Equal(t, 5, got.Seats)
	require.Equal(t, "cus_test_e2e", got.StripeCustomerID)
	require.True(t, got.IsModuleEnabled(domain.ModuleCRA))
}
