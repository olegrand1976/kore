//go:build integration

// Package dbtest spins up an ephemeral Postgres via testcontainers and applies
// all module migrations, so integration tests run against a real database.
package dbtest

import (
	"context"
	"testing"
	"time"

	"github.com/kore/kore/internal/app"
	"github.com/kore/kore/internal/platform/db"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewPostgres starts a throwaway Postgres container, runs every module
// migration and returns a ready-to-use pool. The container and pool are
// torn down automatically at the end of the test.
func NewPostgres(t *testing.T) *db.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("kore_test"),
		postgres.WithUsername("kore"),
		postgres.WithPassword("kore"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := db.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	runner := db.NewMigrationRunner(pool, app.AllModuleMigrations())
	require.NoError(t, runner.Up(ctx))

	return pool
}
