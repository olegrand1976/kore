//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/adapters/postgres"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/modules/tma/ports"
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

func TestTMA_DemandReopenReasonRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	d := domain.NewDemand(tenant, uuid.New(), uuid.New(), "Incident prod", "details", kernel.PriorityHigh, nil, false)
	require.NoError(t, d.Resolve(time.Now().UTC()))
	require.NoError(t, d.Reopen("Correction incomplète"))
	require.NoError(t, repo.Save(ctx, d))

	got, err := repo.Get(ctx, tenant, d.ID)
	require.NoError(t, err)
	require.Equal(t, domain.DemandStatusRework, got.Status)
	require.Equal(t, "Correction incomplète", got.ReopenReason)
}

func TestTMA_DemandTakenOverByRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	d := domain.NewDemand(tenant, uuid.New(), uuid.New(), "Incident prod", "details", kernel.PriorityHigh, nil, false)
	assignee := uuid.New()
	worker := uuid.New()
	require.NoError(t, d.Assign(assignee))
	require.NoError(t, d.TakeOver(worker))
	require.NoError(t, repo.Save(ctx, d))

	got, err := repo.Get(ctx, tenant, d.ID)
	require.NoError(t, err)
	require.NotNil(t, got.AssigneeID)
	require.Equal(t, assignee, *got.AssigneeID)
	require.NotNil(t, got.TakenOverByID)
	require.Equal(t, worker, *got.TakenOverByID)
	require.Equal(t, domain.DemandStatusInProgress, got.Status)
}

func TestTMA_DemandSoftDelete(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	d := domain.NewDemand(tenant, uuid.New(), uuid.New(), "To remove", "details", kernel.PriorityNormal, nil, false)
	require.NoError(t, repo.Save(ctx, d))

	at := time.Now().UTC()
	require.NoError(t, repo.SoftDelete(ctx, tenant, d.ID, at))

	_, err := repo.Get(ctx, tenant, d.ID)
	require.ErrorIs(t, err, domain.ErrDemandNotFound)

	list, err := repo.List(ctx, tenant, ports.ExportFilter{})
	require.NoError(t, err)
	require.Empty(t, list)

	err = repo.SoftDelete(ctx, tenant, d.ID, at)
	require.ErrorIs(t, err, domain.ErrDemandNotFound)
}
