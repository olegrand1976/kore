package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/ett/domain"
	"github.com/kore/kore/internal/modules/ett/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveRecord(ctx context.Context, rec domain.WorkTimeRecord) error {
	_, err := r.pool.Exec(ctx, saveRecordSQL,
		rec.ID, rec.TenantID.UUID(), rec.UserID, rec.WorkDate, rec.ClockIn, rec.ClockOut,
		rec.EffectiveHours, rec.OvertimeHours, rec.Status, rec.Origin, rec.CreatedAt)
	return err
}

func (r *Repository) SaveRecordAndAudit(ctx context.Context, rec domain.WorkTimeRecord, entry domain.AuditEntry) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, saveRecordSQL,
			rec.ID, rec.TenantID.UUID(), rec.UserID, rec.WorkDate, rec.ClockIn, rec.ClockOut,
			rec.EffectiveHours, rec.OvertimeHours, rec.Status, rec.Origin, rec.CreatedAt); err != nil {
			return err
		}
		return r.appendAuditEntryTx(ctx, tx, entry)
	})
}

const saveRecordSQL = `
	INSERT INTO ett.work_time_records (
		id, tenant_id, user_id, work_date, clock_in, clock_out,
		effective_hours, overtime_hours, status, origin, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (tenant_id, user_id, work_date) DO UPDATE SET
		clock_in = EXCLUDED.clock_in,
		clock_out = EXCLUDED.clock_out,
		effective_hours = EXCLUDED.effective_hours,
		overtime_hours = EXCLUDED.overtime_hours,
		status = EXCLUDED.status
`

func (r *Repository) GetRecord(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkTimeRecord, error) {
	return r.scanRecord(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, work_date, clock_in, clock_out,
			effective_hours, overtime_hours, status, origin, created_at
		FROM ett.work_time_records WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) FindRecordByUserDate(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, workDate time.Time) (domain.WorkTimeRecord, error) {
	return r.scanRecord(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, work_date, clock_in, clock_out,
			effective_hours, overtime_hours, status, origin, created_at
		FROM ett.work_time_records WHERE tenant_id = $1 AND user_id = $2 AND work_date = $3
	`, tenant.UUID(), userID, workDate))
}

func (r *Repository) ListRecords(ctx context.Context, q ports.RecordsQuery) ([]domain.WorkTimeRecord, error) {
	query := `
		SELECT id, tenant_id, user_id, work_date, clock_in, clock_out,
			effective_hours, overtime_hours, status, origin, created_at
		FROM ett.work_time_records WHERE tenant_id = $1`
	args := []any{q.TenantID.UUID()}
	if q.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", len(args)+1)
		args = append(args, *q.UserID)
	}
	query += " ORDER BY work_date DESC"
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WorkTimeRecord
	for rows.Next() {
		rec, err := r.scanRecord(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r *Repository) AppendAuditEntry(ctx context.Context, entry domain.AuditEntry) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		return r.appendAuditEntryTx(ctx, tx, entry)
	})
}

func (r *Repository) appendAuditEntryTx(ctx context.Context, tx pgx.Tx, entry domain.AuditEntry) error {
	payload, err := json.Marshal(entry.Payload)
	if err != nil {
		return err
	}
	// Sérialise les ajouts d'un même tenant pour garantir un chaînage cohérent.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, entry.TenantID.UUID().String()); err != nil {
		return err
	}
	var lastSeq int64
	var lastHash string
	err = tx.QueryRow(ctx, `
		SELECT seq, entry_hash FROM ett.audit_journal
		WHERE tenant_id = $1 ORDER BY seq DESC LIMIT 1
	`, entry.TenantID.UUID()).Scan(&lastSeq, &lastHash)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	entry.Seq = lastSeq + 1
	entry.PrevHash = lastHash
	entry.EntryHash = entry.ComputeHash(lastHash)
	_, err = tx.Exec(ctx, `
		INSERT INTO ett.audit_journal (id, tenant_id, record_id, action, actor_id, payload, created_at, seq, prev_hash, entry_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, entry.ID, entry.TenantID.UUID(), entry.RecordID, entry.Action, entry.ActorID, payload, entry.CreatedAt, entry.Seq, entry.PrevHash, entry.EntryHash)
	return err
}

func (r *Repository) ListAuditEntries(ctx context.Context, tenant kernel.TenantID, recordID uuid.UUID) ([]domain.AuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, record_id, action, actor_id, payload, created_at, seq, prev_hash, entry_hash
		FROM ett.audit_journal WHERE tenant_id = $1 AND record_id = $2 ORDER BY seq
	`, tenant.UUID(), recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAuditEntries(rows)
}

func (r *Repository) ListTenantAuditEntries(ctx context.Context, tenant kernel.TenantID) ([]domain.AuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, record_id, action, actor_id, payload, created_at, seq, prev_hash, entry_hash
		FROM ett.audit_journal WHERE tenant_id = $1 ORDER BY seq
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAuditEntries(rows)
}

func scanAuditEntries(rows pgx.Rows) ([]domain.AuditEntry, error) {
	var out []domain.AuditEntry
	for rows.Next() {
		var entry domain.AuditEntry
		var tenantID uuid.UUID
		var payload []byte
		if err := rows.Scan(&entry.ID, &tenantID, &entry.RecordID, &entry.Action, &entry.ActorID, &payload, &entry.CreatedAt, &entry.Seq, &entry.PrevHash, &entry.EntryHash); err != nil {
			return nil, err
		}
		entry.TenantID = kernel.NewTenantID(tenantID)
		entry.Payload = decodeJSON(payload)
		out = append(out, entry)
	}
	return out, rows.Err()
}

func (r *Repository) GetCountryRule(ctx context.Context, tenant kernel.TenantID, countryCode string) (domain.CountryWorkRule, error) {
	var rule domain.CountryWorkRule
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, country_code, max_daily_hours, min_rest_hours, retention_days
		FROM ett.country_work_rules WHERE tenant_id = $1 AND country_code = $2
	`, tenant.UUID(), countryCode).Scan(&rule.ID, &tenantID, &rule.CountryCode, &rule.MaxDailyHours, &rule.MinRestHours, &rule.RetentionDays)
	if err != nil {
		return domain.CountryWorkRule{}, err
	}
	rule.TenantID = kernel.NewTenantID(tenantID)
	return rule, nil
}

func (r *Repository) scanRecord(row pgx.Row) (domain.WorkTimeRecord, error) {
	var rec domain.WorkTimeRecord
	var tenantID uuid.UUID
	err := row.Scan(&rec.ID, &tenantID, &rec.UserID, &rec.WorkDate, &rec.ClockIn, &rec.ClockOut,
		&rec.EffectiveHours, &rec.OvertimeHours, &rec.Status, &rec.Origin, &rec.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkTimeRecord{}, domain.ErrRecordNotFound
		}
		return domain.WorkTimeRecord{}, err
	}
	rec.TenantID = kernel.NewTenantID(tenantID)
	return rec, nil
}

func decodeJSON(b []byte) map[string]any {
	if len(b) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]any{}
	}
	return out
}

var _ ports.ETTRepository = (*Repository)(nil)
