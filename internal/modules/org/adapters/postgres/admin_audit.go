package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

func (r *Repository) RecordAdminAuditEvent(
	ctx context.Context,
	tenant kernel.TenantID,
	actorUserID uuid.UUID,
	action string,
	payload map[string]interface{},
) error {
	if actorUserID == uuid.Nil || action == "" {
		return nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO org.admin_audit_events (id, tenant_id, actor_user_id, action, payload, created_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
	`, uuid.New(), tenant.UUID(), actorUserID, action, string(raw), time.Now().UTC())
	return err
}
