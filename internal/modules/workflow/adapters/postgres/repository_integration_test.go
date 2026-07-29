//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/workflow/adapters/postgres"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestWorkflow_DefinitionRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	def := domain.WorkflowDefinition{
		ID:         uuid.New(),
		TenantID:   tenant,
		Code:       "tma_default",
		EntityType: "demand",
		Version:    1,
		States: []domain.State{
			{Code: "open", Label: "Open", IsInitial: true},
			{Code: "done", Label: "Done", IsFinal: true},
		},
		Transitions: []domain.Transition{
			{From: "open", To: "done", Action: "resolve", AllowedRoles: []string{"Administrateur"}},
		},
	}
	require.NoError(t, def.Validate())
	require.NoError(t, repo.SaveDefinition(ctx, def))

	got, err := repo.GetDefinition(ctx, tenant, "tma_default")
	require.NoError(t, err)
	require.Equal(t, "tma_default", got.Code)
	require.Equal(t, "demand", got.EntityType)
	require.Len(t, got.States, 2)
	require.Len(t, got.Transitions, 1)
}
