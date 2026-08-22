//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/adapters/postgres"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

// Couvre la migration 0002 : integrations.external_links / user_mappings.
func TestTaigaLinks_RoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()
	now := time.Now().UTC()

	link := domain.ExternalLink{
		TenantID:       tenant,
		Provider:       "taiga",
		ExternalType:   "userstory",
		ExternalID:     "42",
		ExternalRef:    intPtr(7),
		KoreEntityType: "demand",
		KoreEntityID:   demandID,
		Metadata:       map[string]any{"action": "create"},
		LastSyncAt:     &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, repo.UpsertExternalLink(ctx, link))

	got, err := repo.FindExternalLinkByKore(ctx, tenant, "demand", demandID)
	require.NoError(t, err)
	require.Equal(t, "42", got.ExternalID)
	require.Equal(t, demandID, got.KoreEntityID)
	require.NotNil(t, got.ExternalRef)
	require.Equal(t, 7, *got.ExternalRef)
}

func TestTaigaLinks_UpsertSameDemandReplacesExternalID(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()
	now := time.Now().UTC()

	first := domain.ExternalLink{
		TenantID:       tenant,
		Provider:       "taiga",
		ExternalType:   "userstory",
		ExternalID:     "smoke-old",
		ExternalRef:    intPtr(42),
		KoreEntityType: "demand",
		KoreEntityID:   demandID,
		Metadata:       map[string]any{"action": "create"},
		LastSyncAt:     &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	require.NoError(t, repo.UpsertExternalLink(ctx, first))

	second := first
	second.ExternalID = "1"
	ref := 1
	second.ExternalRef = &ref
	second.Metadata = map[string]any{"action": "change"}
	require.NoError(t, repo.UpsertExternalLink(ctx, second))

	got, err := repo.FindExternalLinkByKore(ctx, tenant, "demand", demandID)
	require.NoError(t, err)
	require.Equal(t, "1", got.ExternalID)
	require.NotNil(t, got.ExternalRef)
	require.Equal(t, 1, *got.ExternalRef)
}

func TestTaigaUserMappings_UniqueExternalUser(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	now := time.Now().UTC()

	m1 := domain.UserMapping{
		TenantID:         tenant,
		Provider:         "taiga",
		ExternalUserID:   "99",
		ExternalUsername: "alice",
		KoreUserID:       uuid.New(),
		MatchMethod:      "email",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	require.NoError(t, repo.UpsertUserMapping(ctx, m1))

	m2 := m1
	m2.KoreUserID = uuid.New()
	m2.ExternalUsername = "alice-updated"
	require.NoError(t, repo.UpsertUserMapping(ctx, m2))
}

func intPtr(n int) *int {
	return &n
}
