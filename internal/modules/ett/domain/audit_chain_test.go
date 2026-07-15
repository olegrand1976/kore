package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

func chainedEntries(t *testing.T, n int) []AuditEntry {
	t.Helper()
	tenant := kernel.NewTenantID(uuid.New())
	recordID := uuid.New()
	actorID := uuid.New()
	base := time.Date(2027, 1, 4, 8, 0, 0, 0, time.UTC)
	entries := make([]AuditEntry, 0, n)
	prev := ""
	for i := 0; i < n; i++ {
		e := AuditEntry{
			ID:        uuid.New(),
			TenantID:  tenant,
			RecordID:  recordID,
			Action:    "clock_in",
			ActorID:   actorID,
			Payload:   map[string]any{"at": base.Add(time.Duration(i) * time.Hour).Format(time.RFC3339Nano)},
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
			Seq:       int64(i + 1),
			PrevHash:  prev,
		}
		e.EntryHash = e.ComputeHash(prev)
		prev = e.EntryHash
		entries = append(entries, e)
	}
	return entries
}

func TestVerifyChainValid(t *testing.T) {
	entries := chainedEntries(t, 3)
	if broken, ok := VerifyChain(entries); !ok {
		t.Fatalf("expected valid chain, broken at seq %d", broken)
	}
}

func TestVerifyChainDetectsPayloadTampering(t *testing.T) {
	entries := chainedEntries(t, 3)
	// Altère le payload de la 2e entrée sans recalculer le hash.
	entries[1].Payload = map[string]any{"at": "1999-01-01T00:00:00Z"}
	broken, ok := VerifyChain(entries)
	if ok {
		t.Fatal("expected tampering to be detected")
	}
	if broken != entries[1].Seq {
		t.Fatalf("expected break at seq %d, got %d", entries[1].Seq, broken)
	}
}

func TestVerifyChainDetectsBrokenLink(t *testing.T) {
	entries := chainedEntries(t, 3)
	// Supprime l'entrée du milieu : le prev_hash de la 3e ne suit plus.
	entries = []AuditEntry{entries[0], entries[2]}
	if _, ok := VerifyChain(entries); ok {
		t.Fatal("expected broken link to be detected after deletion")
	}
}

func TestComputeHashStableAfterDBRoundTrip(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	at := time.Date(2026, 7, 15, 9, 30, 0, 0, time.UTC)
	inserted := AuditEntry{
		TenantID: tenant, RecordID: uuid.New(), Action: "clock_in", ActorID: uuid.New(),
		Payload: map[string]any{"at": at}, CreatedAt: at, Seq: 1,
	}
	inserted.EntryHash = inserted.ComputeHash("")

	// Simule la relecture PostgreSQL : time.Time devient string RFC3339 sans fractions.
	fromDB := inserted
	fromDB.Payload = map[string]any{"at": "2026-07-15T09:30:00Z"}

	if fromDB.ComputeHash("") != inserted.EntryHash {
		t.Fatal("hash must stay stable after JSONB round-trip")
	}
}
