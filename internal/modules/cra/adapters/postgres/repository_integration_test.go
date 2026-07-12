//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/adapters/postgres"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestCRA_UnicityTenantUserMonth(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	month, err := domain.ParseMonth(time.Now().Format("2006-01"))
	require.NoError(t, err)

	ts1 := domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    month,
		Status:   domain.StatusBrouillon,
	}
	require.NoError(t, repo.Save(ctx, ts1))

	// Same (tenant, user, month) with a different ID must not create a second
	// row: the unique constraint drives an upsert on the existing timesheet.
	ts2 := domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    month,
		Status:   domain.StatusValideSemaine,
	}
	require.NoError(t, repo.Save(ctx, ts2))

	all, err := repo.ListByTenant(ctx, tenant, 10)
	require.NoError(t, err)
	require.Len(t, all, 1, "unicity (tenant,user,month) must keep a single timesheet")
	require.Equal(t, ts1.ID, all[0].ID, "existing row id must be preserved by the upsert")
	require.Equal(t, domain.StatusValideSemaine, all[0].Status, "upsert must update the status")
}
