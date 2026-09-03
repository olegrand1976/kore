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
	for i := range n {
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

// frozenAuditEntryHash est un vecteur figé : il a été calculé une fois et ne doit
// jamais changer. Il verrouille la représentation binaire produite par
// encoding/json (tri des clés de map, échappement HTML, format des nombres), dont
// dépend canonicalPayload et donc entry_hash tel qu'il est persisté en base.
//
// Une montée de toolchain Go ou un refactor de canonicalPayload qui ferait bouger
// ces octets basculerait tout l'historique déjà écrit en INTEGRITY_BROKEN, sans
// rattrapage possible (le chaînage PrevHash est irréversible). Si ce test casse,
// ne pas mettre le vecteur à jour : c'est la compatibilité des données existantes
// qui est en jeu.
const frozenAuditEntryHash = "091cc1738beead8976e4e68aa0a4ddb38349c1240a1293eed558c2f74517e72a"

func TestComputeHashFrozenVector(t *testing.T) {
	at := time.Date(2026, 3, 9, 17, 45, 12, 345678000, time.UTC)
	entry := AuditEntry{
		TenantID:  kernel.NewTenantID(uuid.MustParse("3f1c8c1e-5a4b-4c2d-9e8f-1a2b3c4d5e6f")),
		RecordID:  uuid.MustParse("7d9e0b21-8c33-4f10-a5b6-c7d8e9f0a1b2"),
		Action:    "clock_out",
		ActorID:   uuid.MustParse("11223344-5566-7788-99aa-bbccddeeff00"),
		CreatedAt: at,
		Seq:       42,
		Payload: map[string]any{
			// Clés volontairement hors ordre alphabétique : le hash dépend du tri
			// des clés appliqué par json.Marshal.
			"previousClockOut": nil,
			"note":             "Dupont & Fils <SA>",           // échappé en & / < / >
			"clockOut":         at,                             // normalisé en string 6 décimales
			"hours":            7.5,                            // format des flottants
			"comment":          "café — été",                   // UTF-8 non ASCII
			"meta":             map[string]any{"z": 1, "a": 2}, // map imbriquée, non normalisée
		},
	}

	got := entry.ComputeHash("prev-hash-fixe")
	if got != frozenAuditEntryHash {
		t.Fatalf("la représentation JSON canonique a changé : hash %s, attendu %s\npayload canonique : %s",
			got, frozenAuditEntryHash, canonicalPayload(entry.Payload))
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
