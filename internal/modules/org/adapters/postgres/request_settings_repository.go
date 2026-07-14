package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
)

func (r *Repository) GetTenantRequestSettings(ctx context.Context, tenant kernel.TenantID) (domain.TenantRequestSettings, bool, error) {
	var (
		rawChannels []byte
		guides      bool
		updatedAt   time.Time
	)
	err := r.pool.QueryRow(ctx, `
		SELECT channels_enabled, guides_enabled, updated_at
		FROM org.tenant_request_settings
		WHERE tenant_id = $1
	`, tenant.UUID()).Scan(&rawChannels, &guides, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TenantRequestSettings{}, false, nil
	}
	if err != nil {
		return domain.TenantRequestSettings{}, false, err
	}
	var channels domain.ChannelsEnabled
	if err := json.Unmarshal(rawChannels, &channels); err != nil {
		return domain.TenantRequestSettings{}, false, err
	}
	return domain.TenantRequestSettings{
		TenantID:        tenant,
		ChannelsEnabled: channels,
		GuidesEnabled:   guides,
		UpdatedAt:       updatedAt,
	}, true, nil
}

func (r *Repository) SaveTenantRequestSettings(ctx context.Context, settings domain.TenantRequestSettings) error {
	raw, err := json.Marshal(settings.ChannelsEnabled)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO org.tenant_request_settings (tenant_id, channels_enabled, guides_enabled, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id) DO UPDATE SET
			channels_enabled = EXCLUDED.channels_enabled,
			guides_enabled = EXCLUDED.guides_enabled,
			updated_at = EXCLUDED.updated_at
	`, settings.TenantID.UUID(), raw, settings.GuidesEnabled, settings.UpdatedAt)
	return err
}
