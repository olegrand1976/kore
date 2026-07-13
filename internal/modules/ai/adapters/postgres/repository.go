package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) IsCapabilityEnabled(ctx context.Context, code string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx, `SELECT enabled FROM ai.ai_capabilities WHERE code = $1`, code).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return enabled, err
}

func (r *Repository) GetTenantSettings(ctx context.Context, tenant kernel.TenantID) (domain.TenantSettings, error) {
	var s domain.TenantSettings
	var tenantUUID uuid.UUID
	var noticeAt, workersAt *time.Time
	var noticeBy *uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT tenant_id, enabled, notice_accepted_at, notice_accepted_by, workers_informed_at, llm_provider
		FROM ai.tenant_ai_settings WHERE tenant_id = $1`, tenant.UUID()).Scan(
		&tenantUUID, &s.Enabled, &noticeAt, &noticeBy, &workersAt, &s.LLMProvider,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TenantSettings{TenantID: tenant, Enabled: false, LLMProvider: "stub"}, nil
	}
	if err != nil {
		return domain.TenantSettings{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantUUID)
	s.NoticeAcceptedAt = noticeAt
	s.NoticeAcceptedBy = noticeBy
	s.WorkersInformedAt = workersAt
	return s, nil
}

func (r *Repository) UpsertTenantSettings(ctx context.Context, settings domain.TenantSettings) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai.tenant_ai_settings (tenant_id, enabled, notice_accepted_at, notice_accepted_by, workers_informed_at, llm_provider)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			notice_accepted_at = EXCLUDED.notice_accepted_at,
			notice_accepted_by = EXCLUDED.notice_accepted_by,
			workers_informed_at = EXCLUDED.workers_informed_at,
			llm_provider = EXCLUDED.llm_provider`,
		settings.TenantID.UUID(), settings.Enabled, settings.NoticeAcceptedAt, settings.NoticeAcceptedBy,
		settings.WorkersInformedAt, settings.LLMProvider,
	)
	return err
}

func (r *Repository) InsertRequestLog(ctx context.Context, log domain.RequestLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai.ai_request_log (
			id, tenant_id, user_id, capability_code, entity_type, entity_id,
			input_hash, output_json, model, explain_context, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		log.ID, log.TenantID.UUID(), log.UserID, log.CapabilityCode, nullStr(log.EntityType),
		log.EntityID, log.InputHash, log.OutputJSON, log.Model, encodeJSON(log.ExplainContext), log.CreatedAt,
	)
	return err
}

func (r *Repository) GetRequestLog(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestLog, error) {
	var log domain.RequestLog
	var tenantUUID uuid.UUID
	var entityType *string
	var entityID *uuid.UUID
	var output []byte
	var explain []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, capability_code, entity_type, entity_id,
		       input_hash, output_json, model, explain_context, created_at
		FROM ai.ai_request_log WHERE id = $1 AND tenant_id = $2`, id, tenant.UUID()).Scan(
		&log.ID, &tenantUUID, &log.UserID, &log.CapabilityCode, &entityType, &entityID,
		&log.InputHash, &output, &log.Model, &explain, &log.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RequestLog{}, domain.ErrRequestNotFound
	}
	if err != nil {
		return domain.RequestLog{}, err
	}
	log.TenantID = kernel.NewTenantID(tenantUUID)
	if entityType != nil {
		log.EntityType = *entityType
	}
	log.EntityID = entityID
	log.OutputJSON = output
	log.ExplainContext = decodeJSONMap(explain)
	return log, nil
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func encodeJSON(v map[string]any) []byte {
	if v == nil {
		return []byte("{}")
	}
	b, _ := json.Marshal(v)
	return b
}

func decodeJSONMap(b []byte) map[string]any {
	if len(b) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}

var _ ports.Repository = (*Repository)(nil)
