//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/adapters/postgres"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestConges_LeaveRequestRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	from := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewDateRange(from, to)
	require.NoError(t, err)

	req := domain.NewLeaveRequest(tenant, userID, domain.LeaveTypeCongesPayes, period, "été")
	require.NoError(t, repo.Save(ctx, req))

	got, err := repo.Get(ctx, tenant, req.ID)
	require.NoError(t, err)
	require.Equal(t, req.ID, got.ID)
	require.Equal(t, domain.LeaveStatusPending, got.Status)
	require.Equal(t, domain.LeaveTypeCongesPayes, got.Type)
}
