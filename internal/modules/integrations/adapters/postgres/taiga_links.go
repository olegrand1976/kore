package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

var _ ports.TaigaRepository = (*Repository)(nil)

func (r *Repository) UpsertExternalLink(ctx context.Context, link domain.ExternalLink) error {
	meta, err := json.Marshal(link.Metadata)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		DELETE FROM integrations.external_links
		WHERE tenant_id = $1 AND provider = $2 AND external_type = $3 AND external_id = $4
			AND NOT (kore_entity_type = $5 AND kore_entity_id = $6)
	`, link.TenantID.UUID(), link.Provider, link.ExternalType, link.ExternalID,
		link.KoreEntityType, link.KoreEntityID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO integrations.external_links (
			id, tenant_id, provider, external_type, external_id,
			external_project_id, external_ref, external_url,
			kore_entity_type, kore_entity_id, metadata, last_sync_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (tenant_id, kore_entity_type, kore_entity_id) DO UPDATE SET
			external_id = EXCLUDED.external_id,
			external_project_id = EXCLUDED.external_project_id,
			external_ref = EXCLUDED.external_ref,
			external_url = EXCLUDED.external_url,
			metadata = EXCLUDED.metadata,
			last_sync_at = EXCLUDED.last_sync_at,
			updated_at = EXCLUDED.updated_at
	`, link.ID, link.TenantID.UUID(), link.Provider, link.ExternalType, link.ExternalID,
		link.ExternalProjectID, link.ExternalRef, link.ExternalURL,
		link.KoreEntityType, link.KoreEntityID, meta, link.LastSyncAt, link.CreatedAt, now)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// InsertApplicationProjectLink atomically links one Kore application to one Taiga project per tenant.
// Returns ErrTaigaProjectLinked when the Taiga project is already linked (including concurrent imports).
func (r *Repository) InsertApplicationProjectLink(ctx context.Context, link domain.ExternalLink) error {
	meta, err := json.Marshal(link.Metadata)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	if link.CreatedAt.IsZero() {
		link.CreatedAt = now
	}
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO integrations.external_links (
			id, tenant_id, provider, external_type, external_id,
			external_project_id, external_ref, external_url,
			kore_entity_type, kore_entity_id, metadata, last_sync_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (tenant_id, provider, external_type, external_id) DO NOTHING
	`, link.ID, link.TenantID.UUID(), link.Provider, link.ExternalType, link.ExternalID,
		link.ExternalProjectID, link.ExternalRef, link.ExternalURL,
		link.KoreEntityType, link.KoreEntityID, meta, link.LastSyncAt, link.CreatedAt, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTaigaProjectLinked
	}
	return nil
}

func (r *Repository) FindExternalLinkByKore(ctx context.Context, tenant kernel.TenantID, koreEntityType string, koreEntityID uuid.UUID) (domain.ExternalLink, error) {
	return r.scanExternalLink(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, provider, external_type, external_id,
			external_project_id, external_ref, external_url,
			kore_entity_type, kore_entity_id, metadata, last_sync_at, created_at, updated_at
		FROM integrations.external_links
		WHERE tenant_id = $1 AND kore_entity_type = $2 AND kore_entity_id = $3
		LIMIT 1
	`, tenant.UUID(), koreEntityType, koreEntityID))
}

func (r *Repository) ListLinkedTaigaProjectIDs(ctx context.Context, tenant kernel.TenantID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT external_id
		FROM integrations.external_links
		WHERE tenant_id = $1 AND provider = 'taiga' AND external_type = 'project'
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) UpsertUserMapping(ctx context.Context, mapping domain.UserMapping) error {
	now := time.Now().UTC()
	if mapping.ID == uuid.Nil {
		mapping.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO integrations.user_mappings (
			id, tenant_id, provider, external_user_id, external_username,
			kore_user_id, match_method, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, provider, external_user_id) DO UPDATE SET
			external_username = EXCLUDED.external_username,
			kore_user_id = EXCLUDED.kore_user_id,
			match_method = EXCLUDED.match_method,
			updated_at = EXCLUDED.updated_at
	`, mapping.ID, mapping.TenantID.UUID(), mapping.Provider, mapping.ExternalUserID,
		mapping.ExternalUsername, mapping.KoreUserID, mapping.MatchMethod, mapping.CreatedAt, now)
	return err
}

func (r *Repository) scanExternalLink(row pgx.Row) (domain.ExternalLink, error) {
	var link domain.ExternalLink
	var tenantID uuid.UUID
	var meta []byte
	err := row.Scan(
		&link.ID, &tenantID, &link.Provider, &link.ExternalType, &link.ExternalID,
		&link.ExternalProjectID, &link.ExternalRef, &link.ExternalURL,
		&link.KoreEntityType, &link.KoreEntityID, &meta, &link.LastSyncAt, &link.CreatedAt, &link.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ExternalLink{}, domain.ErrExternalLinkNotFound
		}
		return domain.ExternalLink{}, err
	}
	link.TenantID = kernel.NewTenantID(tenantID)
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &link.Metadata)
	}
	return link, nil
}
