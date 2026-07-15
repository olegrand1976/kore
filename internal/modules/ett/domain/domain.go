package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrRecordNotFound = errors.New("work time record not found")
	ErrNotSalarieETT  = errors.New("user is not an ETT employee")
)

type WorkTimeRecord struct {
	ID             uuid.UUID
	TenantID       kernel.TenantID
	UserID         uuid.UUID
	WorkDate       time.Time
	ClockIn        *time.Time
	ClockOut       *time.Time
	EffectiveHours float64
	OvertimeHours  float64
	Status         string
	Origin         string
	CreatedAt      time.Time
}

type AuditEntry struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	RecordID  uuid.UUID
	Action    string
	ActorID   uuid.UUID
	Payload   map[string]any
	CreatedAt time.Time
	// Seq est une position monotone par tenant ; PrevHash/EntryHash forment
	// une chaîne de hachage rendant le journal tamper-evident (RG-ETT-01).
	Seq       int64
	PrevHash  string
	EntryHash string
}

type CountryWorkRule struct {
	ID            uuid.UUID
	TenantID      kernel.TenantID
	CountryCode   string
	MaxDailyHours float64
	MinRestHours  float64
	RetentionDays int
}

func NewWorkTimeRecord(tenant kernel.TenantID, userID uuid.UUID, workDate time.Time) WorkTimeRecord {
	return WorkTimeRecord{
		ID:        uuid.New(),
		TenantID:  tenant,
		UserID:    userID,
		WorkDate:  workDate,
		Status:    "pointed",
		Origin:    "web",
		CreatedAt: time.Now().UTC(),
	}
}

func NewAuditEntry(tenant kernel.TenantID, recordID, actorID uuid.UUID, action string, payload map[string]any) AuditEntry {
	if payload == nil {
		payload = map[string]any{}
	}
	return AuditEntry{
		ID:       uuid.New(),
		TenantID: tenant,
		RecordID: recordID,
		Action:   action,
		ActorID:  actorID,
		Payload:  payload,
		// Tronqué à la microseconde pour rester identique à la valeur
		// persistée (timestamptz PostgreSQL) et garder le hash reproductible.
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
}

// canonicalPayload normalise le payload (round-trip JSON) pour que le hash soit
// identique qu'il soit calculé depuis la valeur typée (à l'insertion) ou depuis
// la valeur générique relue en base (à la vérification).
func canonicalPayload(payload map[string]any) []byte {
	if payload == nil {
		payload = map[string]any{}
	}
	normalized := normalizePayloadMap(payload)
	raw, err := json.Marshal(normalized)
	if err != nil {
		return []byte("{}")
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return raw
	}
	out, err := json.Marshal(generic)
	if err != nil {
		return raw
	}
	return out
}

func normalizePayloadMap(payload map[string]any) map[string]any {
	out := make(map[string]any, len(payload))
	for k, v := range payload {
		out[k] = normalizePayloadValue(v)
	}
	return out
}

func normalizePayloadValue(v any) any {
	switch t := v.(type) {
	case time.Time:
		return formatAuditTimestamp(t)
	case *time.Time:
		if t == nil {
			return nil
		}
		return formatAuditTimestamp(*t)
	case string:
		if ts, ok := parseAuditTimestampString(t); ok {
			return formatAuditTimestamp(ts)
		}
		return v
	default:
		return v
	}
}

func parseAuditTimestampString(s string) (time.Time, bool) {
	for _, layout := range []string{
		"2006-01-02T15:04:05.000000Z",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	} {
		if ts, err := time.Parse(layout, s); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}

// formatAuditTimestamp aligne le format horodatage avec le backfill SQL (6 décimales UTC).
func formatAuditTimestamp(t time.Time) string {
	return t.UTC().Truncate(time.Microsecond).Format("2006-01-02T15:04:05.000000Z")
}

// ComputeHash calcule le hash SHA-256 de l'entrée chaînée au hash précédent.
func (e AuditEntry) ComputeHash(prevHash string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%d|%s|%s|%s|%s|%s|%s",
		prevHash,
		e.Seq,
		e.TenantID.UUID(),
		e.RecordID,
		e.Action,
		e.ActorID,
		formatAuditTimestamp(e.CreatedAt),
		canonicalPayload(e.Payload),
	)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyChain vérifie l'intégrité d'une chaîne d'entrées ordonnée par Seq.
// Retourne la position de la première rupture (0 si aucune) et un booléen valide.
func VerifyChain(entries []AuditEntry) (int64, bool) {
	prev := ""
	for _, entry := range entries {
		if entry.PrevHash != prev {
			return entry.Seq, false
		}
		if entry.ComputeHash(prev) != entry.EntryHash {
			return entry.Seq, false
		}
		prev = entry.EntryHash
	}
	return 0, true
}

func (r *WorkTimeRecord) ClockInAt(t time.Time) {
	r.ClockIn = &t
}

func (r *WorkTimeRecord) ClockOutAt(t time.Time) {
	r.ClockOut = &t
	if r.ClockIn != nil {
		hours := t.Sub(*r.ClockIn).Hours()
		r.EffectiveHours = hours
	}
}
