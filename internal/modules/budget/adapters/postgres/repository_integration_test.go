//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/budget/adapters/postgres"
	"github.com/kore/kore/internal/modules/budget/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestBudget_SaveGetRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	b := domain.NewBudget(tenant, uuid.New(), domain.BudgetTypeDefault, domain.ConsumptionTriple{
		Days: 10, UO: 5, Amount: 1000,
	}, "EUR")
	require.NoError(t, repo.Save(ctx, b))

	got, err := repo.Get(ctx, tenant, b.ID)
	require.NoError(t, err)
	require.Equal(t, b.ID, got.ID)
	require.Equal(t, domain.BudgetTypeDefault, got.Type)
	require.Equal(t, 10.0, got.Planned.Days)
	require.Equal(t, "EUR", got.Currency)
}
