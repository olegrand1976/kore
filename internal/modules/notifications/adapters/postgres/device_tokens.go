package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/pkg/kernel"
)

func (r *Repository) UpsertDeviceToken(ctx context.Context, t domain.DeviceToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications.device_tokens (id, tenant_id, user_id, platform, token, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, user_id, token) DO UPDATE SET
			platform = EXCLUDED.platform,
			updated_at = EXCLUDED.updated_at
	`, t.ID, t.TenantID.UUID(), t.UserID, string(t.Platform), t.Token, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *Repository) DeleteDeviceToken(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, token string) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM notifications.device_tokens
		WHERE tenant_id = $1 AND user_id = $2 AND token = $3
	`, tenant.UUID(), userID, token)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("device token not found")
	}
	return nil
}

func (r *Repository) DeleteDeviceTokenByValue(ctx context.Context, tenant kernel.TenantID, token string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM notifications.device_tokens
		WHERE tenant_id = $1 AND token = $2
	`, tenant.UUID(), token)
	return err
}

func (r *Repository) ListDeviceTokens(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.DeviceToken, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, user_id, platform, token, created_at, updated_at
		FROM notifications.device_tokens
		WHERE tenant_id = $1 AND user_id = $2
		ORDER BY updated_at DESC
	`, tenant.UUID(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DeviceToken
	for rows.Next() {
		var t domain.DeviceToken
		var tenantUUID uuid.UUID
		var platform string
		if err := rows.Scan(&t.ID, &tenantUUID, &t.UserID, &platform, &t.Token, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.TenantID = kernel.NewTenantID(tenantUUID)
		t.Platform = domain.DevicePlatform(platform)
		out = append(out, t)
	}
	return out, rows.Err()
}
