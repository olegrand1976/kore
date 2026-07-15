//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ett/adapters/postgres"
	"github.com/kore/kore/internal/modules/ett/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestETT_AuditJournalHashChain(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	recordID := uuid.New()
	actorID := uuid.New()

	for i := 0; i < 3; i++ {
		entry := domain.NewAuditEntry(tenant, recordID, actorID, "clock_in", map[string]any{"i": i})
		require.NoError(t, repo.AppendAuditEntry(ctx, entry))
	}

	entries, err := repo.ListTenantAuditEntries(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, entries, 3)

	require.Equal(t, int64(1), entries[0].Seq)
	require.Empty(t, entries[0].PrevHash, "first entry has no predecessor")
	require.NotEmpty(t, entries[0].EntryHash)
	require.Equal(t, entries[0].EntryHash, entries[1].PrevHash, "chain must link entries")
	require.Equal(t, entries[1].EntryHash, entries[2].PrevHash)

	broken, ok := domain.VerifyChain(entries)
	require.True(t, ok, "chain must be valid, broken at %d", broken)
}

func TestETT_AuditJournalHashChainWithTimePayload(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	recordID := uuid.New()
	actorID := uuid.New()
	at := time.Date(2026, 7, 15, 9, 30, 0, 0, time.UTC)

	entry := domain.NewAuditEntry(tenant, recordID, actorID, "clock_in", map[string]any{
		"at": at,
	})
	require.NoError(t, repo.AppendAuditEntry(ctx, entry))

	entries, err := repo.ListTenantAuditEntries(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	broken, ok := domain.VerifyChain(entries)
	require.True(t, ok, "time payload chain must be valid, broken at %d", broken)
}

func TestETT_AuditJournalIsAppendOnly(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	recordID := uuid.New()
	entry := domain.NewAuditEntry(tenant, recordID, uuid.New(), "clock_in", map[string]any{"at": time.Now().UTC()})
	require.NoError(t, repo.AppendAuditEntry(ctx, entry))

	// Le trigger doit rejeter toute mutation directe (inaltérabilité RG-ETT-01).
	_, updErr := pool.Exec(ctx, `UPDATE ett.audit_journal SET action = 'tampered' WHERE tenant_id = $1`, tenant.UUID())
	require.Error(t, updErr, "UPDATE on audit_journal must be rejected")

	_, delErr := pool.Exec(ctx, `DELETE FROM ett.audit_journal WHERE tenant_id = $1`, tenant.UUID())
	require.Error(t, delErr, "DELETE on audit_journal must be rejected")
}
