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
	require.Equal(t, 1, got.TicketNumber)
}

func TestTMA_DemandTicketNumbersSequentialAndEnsure(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	d1 := domain.NewDemand(tenant, uuid.New(), uuid.New(), "First", "", kernel.PriorityNormal, nil, false)
	d2 := domain.NewDemand(tenant, uuid.New(), uuid.New(), "Second", "", kernel.PriorityNormal, nil, false)
	require.NoError(t, repo.Save(ctx, d1))
	require.NoError(t, repo.Save(ctx, d2))

	got1, err := repo.Get(ctx, tenant, d1.ID)
	require.NoError(t, err)
	got2, err := repo.Get(ctx, tenant, d2.ID)
	require.NoError(t, err)
	require.Equal(t, 1, got1.TicketNumber)
	require.Equal(t, 2, got2.TicketNumber)

	// Simulate a legacy row without ticket_number (pre-migration edge / NULL).
	_, err = pool.Exec(ctx, `UPDATE tma.demands SET ticket_number = NULL WHERE id = $1`, d2.ID)
	require.NoError(t, err)

	require.NoError(t, repo.EnsureTicketNumbers(ctx, tenant))
	got2, err = repo.Get(ctx, tenant, d2.ID)
	require.NoError(t, err)
	require.Equal(t, 3, got2.TicketNumber)

	list, err := repo.List(ctx, tenant, ports.ExportFilter{})
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, 1, list[0].TicketNumber)
	require.Equal(t, 3, list[1].TicketNumber)
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
