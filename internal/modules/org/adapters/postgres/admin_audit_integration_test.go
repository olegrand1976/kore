//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/adapters/postgres"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestRecordAdminAuditEvent_persistsMergePayload(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	_, _, _ = seedStructure(t, repo, tenant)
	actorID := uuid.New()
	absorbedID := uuid.New()
	referenceID := uuid.New()

	require.NoError(t, repo.RecordAdminAuditEvent(ctx, tenant, actorID, "applications.merge", map[string]interface{}{
		"absorbedApplicationId":  absorbedID.String(),
		"referenceApplicationId": referenceID.String(),
	}))

	var action string
	var payloadRaw []byte
	err := pool.QueryRow(ctx, `
		SELECT action, payload FROM org.admin_audit_events
		WHERE tenant_id = $1 AND actor_user_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, tenant.UUID(), actorID).Scan(&action, &payloadRaw)
	require.NoError(t, err)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(payloadRaw, &payload))
	require.Equal(t, "applications.merge", action)
	require.Equal(t, absorbedID.String(), payload["absorbedApplicationId"])
	require.Equal(t, referenceID.String(), payload["referenceApplicationId"])
}
