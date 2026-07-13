package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/admin/domain"
	"github.com/kore/kore/internal/modules/admin/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetParameterSet(ctx context.Context, tenant kernel.TenantID, code string) (domain.ParameterSet, error) {
	return r.scanParameterSet(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, code, payload, updated_at
		FROM admin.parameter_sets WHERE tenant_id = $1 AND code = $2
	`, tenant.UUID(), code))
}

func (r *Repository) SaveParameterSet(ctx context.Context, ps domain.ParameterSet) error {
	payload, err := json.Marshal(ps.Payload)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO admin.parameter_sets (id, tenant_id, code, payload, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, code) DO UPDATE SET payload = EXCLUDED.payload, updated_at = EXCLUDED.updated_at
	`, ps.ID, ps.TenantID.UUID(), ps.Code, payload, ps.UpdatedAt)
	return err
}

func (r *Repository) ListTemplates(ctx context.Context, tenant kernel.TenantID) ([]domain.Template, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, type, name, content, active, created_at
		FROM admin.templates WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Template
	for rows.Next() {
		t, err := r.scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) GetTemplate(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Template, error) {
	return r.scanTemplate(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, type, name, content, active, created_at
		FROM admin.templates WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) SaveTemplate(ctx context.Context, t domain.Template) error {
	content, err := json.Marshal(t.Content)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO admin.templates (id, tenant_id, type, name, content, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, t.ID, t.TenantID.UUID(), t.Type, t.Name, content, t.Active, t.CreatedAt)
	return err
}

func (r *Repository) ListPhoneDirectory(ctx context.Context, tenant kernel.TenantID) ([]domain.PhoneDirectoryEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, user_id, label, phone, visibility, created_at
		FROM admin.phone_directory WHERE tenant_id = $1 ORDER BY label
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PhoneDirectoryEntry
	for rows.Next() {
		e, err := r.scanPhoneEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) SavePhoneEntry(ctx context.Context, e domain.PhoneDirectoryEntry) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin.phone_directory (id, tenant_id, user_id, label, phone, visibility, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, e.ID, e.TenantID.UUID(), e.UserID, e.Label, e.Phone, e.Visibility, e.CreatedAt)
	return err
}

func (r *Repository) scanParameterSet(row pgx.Row) (domain.ParameterSet, error) {
	var ps domain.ParameterSet
	var tenantID uuid.UUID
	var payload []byte
	err := row.Scan(&ps.ID, &tenantID, &ps.Code, &payload, &ps.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ParameterSet{}, domain.ErrParameterSetNotFound
		}
		return domain.ParameterSet{}, err
	}
	ps.TenantID = kernel.NewTenantID(tenantID)
	ps.Payload = decodeJSON(payload)
	return ps, nil
}

func (r *Repository) scanTemplate(row pgx.Row) (domain.Template, error) {
	var t domain.Template
	var tenantID uuid.UUID
	var content []byte
	err := row.Scan(&t.ID, &tenantID, &t.Type, &t.Name, &content, &t.Active, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Template{}, domain.ErrTemplateNotFound
		}
		return domain.Template{}, err
	}
	t.TenantID = kernel.NewTenantID(tenantID)
	t.Content = decodeJSON(content)
	return t, nil
}

func (r *Repository) scanPhoneEntry(row pgx.Row) (domain.PhoneDirectoryEntry, error) {
	var e domain.PhoneDirectoryEntry
	var tenantID uuid.UUID
	err := row.Scan(&e.ID, &tenantID, &e.UserID, &e.Label, &e.Phone, &e.Visibility, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PhoneDirectoryEntry{}, domain.ErrPhoneDirectoryEntryNotFound
		}
		return domain.PhoneDirectoryEntry{}, err
	}
	e.TenantID = kernel.NewTenantID(tenantID)
	return e, nil
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

var _ ports.AdminRepository = (*Repository)(nil)
