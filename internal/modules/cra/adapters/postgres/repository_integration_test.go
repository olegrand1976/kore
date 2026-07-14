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

func TestCRA_FindConsumption_BillableApplicationLines(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	appID := uuid.New()
	month := domain.Month("2026-07")
	weekID := uuid.New()
	day := time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)

	ts := domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    month,
		Status:   domain.StatusBrouillon,
		Weeks: []domain.WeekEntry{{
			ID:         weekID,
			WeekNumber: 2,
			Lines: []domain.TimeLine{{
				ID:          uuid.New(),
				TenantID:    tenant,
				WeekEntryID: weekID,
				Source:      domain.SourceRef{Type: "application", ID: appID.String()},
				Day:         day,
				Duration:    kernel.Duration{Minutes: 240},
				Billable:    true,
				Origin:      domain.OriginManual,
			}, {
				ID:          uuid.New(),
				TenantID:    tenant,
				WeekEntryID: weekID,
				Source:      domain.SourceRef{Type: "application", ID: appID.String()},
				Day:         day,
				Duration:    kernel.Duration{Minutes: 120},
				Billable:    false,
				Origin:      domain.OriginManual,
			}},
		}},
	}
	require.NoError(t, repo.Save(ctx, ts))

	period, err := kernel.NewPeriod(day, day)
	require.NoError(t, err)
	items, err := repo.FindConsumption(ctx, tenant, appID, period)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, 240, items[0].Duration.Minutes)
}

func TestCRA_ListSummariesWithoutSSIISchema(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	ctx := context.Background()
	repo := postgres.NewRepository(pool)

	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	month := domain.Month("2026-07")
	require.NoError(t, repo.Save(ctx, domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    month,
		Status:   domain.StatusBrouillon,
	}))

	_, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS ssii CASCADE`)
	require.NoError(t, err)

	repo = postgres.NewRepository(pool)
	_, err = repo.ListSummariesByTenant(ctx, tenant, 10)
	require.NoError(t, err)
}

func TestCRA_SaveAndGetWithoutRejectReasonColumn(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	ctx := context.Background()
	repo := postgres.NewRepository(pool)

	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	month := domain.Month("2026-08")
	tsID := uuid.New()

	require.NoError(t, repo.Save(ctx, domain.Timesheet{
		ID:       tsID,
		TenantID: tenant,
		UserID:   userID,
		Month:    month,
		Status:   domain.StatusBrouillon,
	}))

	_, err := pool.Exec(ctx, `
		ALTER TABLE cra.timesheets
			DROP COLUMN IF EXISTS reject_reason,
			DROP COLUMN IF EXISTS rejected_at,
			DROP COLUMN IF EXISTS rejected_by
	`)
	require.NoError(t, err)

	repo = postgres.NewRepository(pool)
	newID := uuid.New()
	require.NoError(t, repo.Save(ctx, domain.Timesheet{
		ID:       newID,
		TenantID: tenant,
		UserID:   userID,
		Month:    month,
		Status:   domain.StatusValideSemaine,
	}))

	got, err := repo.Get(ctx, tenant, userID, month)
	require.NoError(t, err)
	require.Equal(t, domain.StatusValideSemaine, got.Status)
}
