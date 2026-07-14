package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveConnection(ctx context.Context, c domain.IntegrationConnection) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO integrations.connections (
			id, tenant_id, type, provider, status, credentials_ref, last_sync_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			credentials_ref = EXCLUDED.credentials_ref,
			last_sync_at = EXCLUDED.last_sync_at
	`, c.ID, c.TenantID.UUID(), string(c.Type), c.Provider, string(c.Status), c.CredentialsRef, c.LastSyncAt, c.CreatedAt)
	return err
}

func (r *Repository) GetConnection(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.IntegrationConnection, error) {
	return r.scanConnection(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, type, provider, status, credentials_ref, last_sync_at, created_at
		FROM integrations.connections WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) ListConnections(ctx context.Context, tenant kernel.TenantID) ([]domain.IntegrationConnection, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, type, provider, status, credentials_ref, last_sync_at, created_at
		FROM integrations.connections WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.IntegrationConnection
	for rows.Next() {
		c, err := r.scanConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) DeleteConnection(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM integrations.connections WHERE tenant_id = $1 AND id = $2`, tenant.UUID(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConnectionNotFound
	}
	return nil
}

func (r *Repository) SaveSyncJob(ctx context.Context, j domain.SyncJob) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO integrations.sync_jobs (id, tenant_id, connection_id, status, started_at, finished_at, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, j.ID, j.TenantID.UUID(), j.ConnectionID, j.Status, j.StartedAt, j.FinishedAt, j.ErrorMessage)
	return err
}

func (r *Repository) ListSyncJobs(ctx context.Context, tenant kernel.TenantID) ([]domain.SyncJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, connection_id, status, started_at, finished_at, error_message
		FROM integrations.sync_jobs WHERE tenant_id = $1 ORDER BY started_at DESC LIMIT 100
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SyncJob
	for rows.Next() {
		j, err := r.scanSyncJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *Repository) SaveApiKey(ctx context.Context, k domain.ApiKey) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO integrations.api_keys (id, tenant_id, name, key_prefix, key_hash, revoked_at, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET revoked_at = EXCLUDED.revoked_at, last_used_at = EXCLUDED.last_used_at
	`, k.ID, k.TenantID.UUID(), k.Name, k.KeyPrefix, k.KeyHash, k.RevokedAt, k.CreatedAt, k.LastUsedAt)
	return err
}

func (r *Repository) GetApiKey(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.ApiKey, error) {
	return r.scanApiKey(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, key_prefix, key_hash, revoked_at, created_at, last_used_at
		FROM integrations.api_keys WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) GetApiKeyByHash(ctx context.Context, keyHash string) (domain.ApiKey, error) {
	return r.scanApiKey(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, key_prefix, key_hash, revoked_at, created_at, last_used_at
		FROM integrations.api_keys WHERE key_hash = $1
	`, keyHash))
}

func (r *Repository) SaveWebhookSubscription(ctx context.Context, sub domain.WebhookSubscription) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO integrations.webhook_subscriptions (id, tenant_id, url, events, secret_ref, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, sub.ID, sub.TenantID.UUID(), sub.URL, sub.Events, sub.SecretRef, sub.Active, sub.CreatedAt)
	return err
}

func (r *Repository) ListWebhookSubscriptions(ctx context.Context, tenant kernel.TenantID) ([]domain.WebhookSubscription, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, url, events, secret_ref, active, created_at
		FROM integrations.webhook_subscriptions WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WebhookSubscription
	for rows.Next() {
		sub, err := r.scanWebhookSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (r *Repository) DeleteWebhookSubscription(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM integrations.webhook_subscriptions WHERE tenant_id = $1 AND id = $2`, tenant.UUID(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWebhookNotFound
	}
	return nil
}

func (r *Repository) ListApiKeys(ctx context.Context, tenant kernel.TenantID) ([]domain.ApiKey, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, key_prefix, key_hash, revoked_at, created_at, last_used_at
		FROM integrations.api_keys WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ApiKey
	for rows.Next() {
		k, err := r.scanApiKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *Repository) scanSyncJob(row pgx.Row) (domain.SyncJob, error) {
	var j domain.SyncJob
	var tenantID uuid.UUID
	err := row.Scan(&j.ID, &tenantID, &j.ConnectionID, &j.Status, &j.StartedAt, &j.FinishedAt, &j.ErrorMessage)
	if err != nil {
		return domain.SyncJob{}, err
	}
	j.TenantID = kernel.NewTenantID(tenantID)
	return j, nil
}

func (r *Repository) scanWebhookSubscription(row pgx.Row) (domain.WebhookSubscription, error) {
	var sub domain.WebhookSubscription
	var tenantID uuid.UUID
	err := row.Scan(&sub.ID, &tenantID, &sub.URL, &sub.Events, &sub.SecretRef, &sub.Active, &sub.CreatedAt)
	if err != nil {
		return domain.WebhookSubscription{}, err
	}
	sub.TenantID = kernel.NewTenantID(tenantID)
	return sub, nil
}

func (r *Repository) scanConnection(row pgx.Row) (domain.IntegrationConnection, error) {
	var c domain.IntegrationConnection
	var tenantID uuid.UUID
	var connType, status string
	err := row.Scan(&c.ID, &tenantID, &connType, &c.Provider, &status, &c.CredentialsRef, &c.LastSyncAt, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.IntegrationConnection{}, domain.ErrConnectionNotFound
		}
		return domain.IntegrationConnection{}, err
	}
	c.TenantID = kernel.NewTenantID(tenantID)
	c.Type = domain.ConnectionType(connType)
	c.Status = domain.ConnectionStatus(status)
	return c, nil
}

func (r *Repository) scanApiKey(row pgx.Row) (domain.ApiKey, error) {
	var k domain.ApiKey
	var tenantID uuid.UUID
	err := row.Scan(&k.ID, &tenantID, &k.Name, &k.KeyPrefix, &k.KeyHash, &k.RevokedAt, &k.CreatedAt, &k.LastUsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ApiKey{}, domain.ErrApiKeyNotFound
		}
		return domain.ApiKey{}, err
	}
	k.TenantID = kernel.NewTenantID(tenantID)
	return k, nil
}

var _ ports.IntegrationRepository = (*Repository)(nil)
