//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/adapters/postgres"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestTMA_DemandRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	d := domain.NewDemand(tenant, uuid.New(), uuid.New(), "Incident prod", "details", kernel.PriorityHigh, nil, false)
	require.NoError(t, repo.Save(ctx, d))

	got, err := repo.Get(ctx, tenant, d.ID)
	require.NoError(t, err)
	require.Equal(t, d.ID, got.ID)
	require.Equal(t, domain.DemandStatusOpen, got.Status)
	require.True(t, got.Visible)
	require.Equal(t, "Incident prod", got.Subject)
}
