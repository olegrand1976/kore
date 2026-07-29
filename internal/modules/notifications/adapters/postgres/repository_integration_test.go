//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/adapters/postgres"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestNotifications_RuleRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	rule := domain.NotificationRule{
		ID:        uuid.New(),
		TenantID:  tenant,
		Code:      "cra_submitted",
		Trigger:   "cra.submitted",
		Frequency: domain.FrequencyImmediate,
		RecipientsPolicy: domain.RecipientPolicy{
			UserIDs: []uuid.UUID{uuid.New()},
		},
		Template:  "CRA soumis {{user}}",
		AttachPDF: false,
	}
	require.NoError(t, repo.SaveRule(ctx, rule))

	got, err := repo.GetRuleByCode(ctx, tenant, "cra_submitted")
	require.NoError(t, err)
	require.Equal(t, rule.ID, got.ID)
	require.Equal(t, domain.FrequencyImmediate, got.Frequency)
	require.Equal(t, "cra.submitted", got.Trigger)
}
