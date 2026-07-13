package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/pkg/kernel"
)

func (r *Repository) ListLeaveTypeConfigs(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, activeOnly bool) ([]domain.LeaveTypeConfig, error) {
	query := `
		SELECT id, tenant_id, societe_id, code, label, tracks_balance, active, sort_order, created_at, updated_at
		FROM conges.leave_type_configs
		WHERE tenant_id = $1 AND societe_id = $2`
	args := []any{tenant.UUID(), societeID}
	if activeOnly {
		query += " AND active = TRUE"
	}
	query += " ORDER BY sort_order, label"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.LeaveTypeConfig
	for rows.Next() {
		cfg, err := scanLeaveTypeConfig(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cfg)
	}
	return out, rows.Err()
}

func (r *Repository) GetLeaveTypeConfig(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.LeaveTypeConfig, error) {
	return scanLeaveTypeConfig(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, societe_id, code, label, tracks_balance, active, sort_order, created_at, updated_at
		FROM conges.leave_type_configs WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) GetLeaveTypeConfigByCode(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, code string) (domain.LeaveTypeConfig, error) {
	return scanLeaveTypeConfig(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, societe_id, code, label, tracks_balance, active, sort_order, created_at, updated_at
		FROM conges.leave_type_configs WHERE tenant_id = $1 AND societe_id = $2 AND code = $3
	`, tenant.UUID(), societeID, code))
}

func (r *Repository) SaveLeaveTypeConfig(ctx context.Context, cfg domain.LeaveTypeConfig) error {
	now := time.Now().UTC()
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = now
	}
	cfg.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO conges.leave_type_configs (
			id, tenant_id, societe_id, code, label, tracks_balance, active, sort_order, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, societe_id, code) DO UPDATE SET
			label = EXCLUDED.label,
			tracks_balance = EXCLUDED.tracks_balance,
			active = EXCLUDED.active,
			sort_order = EXCLUDED.sort_order,
			updated_at = EXCLUDED.updated_at
	`, cfg.ID, cfg.TenantID.UUID(), cfg.SocieteID, cfg.Code, cfg.Label, cfg.TracksBalance, cfg.Active, cfg.SortOrder, cfg.CreatedAt, cfg.UpdatedAt)
	return err
}

func (r *Repository) DeleteLeaveTypeConfig(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM conges.leave_type_configs WHERE tenant_id = $1 AND id = $2`, tenant.UUID(), id)
	return err
}

func (r *Repository) IsLeaveTypeCodeUsed(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, code string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT (
			(SELECT COUNT(*) FROM conges.leave_requests WHERE tenant_id = $1 AND type = $2) +
			(SELECT COUNT(*) FROM conges.leave_balances WHERE tenant_id = $1 AND type = $2)
		)
	`, tenant.UUID(), code).Scan(&count)
	return count > 0, err
}

func (r *Repository) UpsertLeaveTypeDefaults(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, templates []domain.LeaveTypeTemplate) error {
	for _, tpl := range templates {
		cfg := domain.LeaveTypeConfig{
			ID:            uuid.New(),
			TenantID:      tenant,
			SocieteID:     societeID,
			Code:          tpl.Code,
			Label:         tpl.Label,
			TracksBalance: tpl.TracksBalance,
			Active:        true,
			SortOrder:     tpl.SortOrder,
		}
		if err := r.SaveLeaveTypeConfig(ctx, cfg); err != nil {
			return err
		}
	}
	return nil
}

func scanLeaveTypeConfig(row pgx.Row) (domain.LeaveTypeConfig, error) {
	var cfg domain.LeaveTypeConfig
	var tenantID uuid.UUID
	err := row.Scan(
		&cfg.ID, &tenantID, &cfg.SocieteID, &cfg.Code, &cfg.Label,
		&cfg.TracksBalance, &cfg.Active, &cfg.SortOrder, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LeaveTypeConfig{}, domain.ErrLeaveTypeNotFound
		}
		return domain.LeaveTypeConfig{}, err
	}
	cfg.TenantID = kernel.NewTenantID(tenantID)
	return cfg, nil
}

// LeaveTypeConfigRepoAdapter implements ports.LeaveTypeConfigRepository on Repository.
type LeaveTypeConfigRepoAdapter struct {
	*Repository
}

func NewLeaveTypeConfigRepoAdapter(repo *Repository) *LeaveTypeConfigRepoAdapter {
	return &LeaveTypeConfigRepoAdapter{Repository: repo}
}

func (a *LeaveTypeConfigRepoAdapter) ListBySociete(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, activeOnly bool) ([]domain.LeaveTypeConfig, error) {
	return a.ListLeaveTypeConfigs(ctx, tenant, societeID, activeOnly)
}

func (a *LeaveTypeConfigRepoAdapter) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.LeaveTypeConfig, error) {
	return a.GetLeaveTypeConfig(ctx, tenant, id)
}

func (a *LeaveTypeConfigRepoAdapter) GetByCode(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, code string) (domain.LeaveTypeConfig, error) {
	return a.GetLeaveTypeConfigByCode(ctx, tenant, societeID, code)
}

func (a *LeaveTypeConfigRepoAdapter) Save(ctx context.Context, cfg domain.LeaveTypeConfig) error {
	return a.SaveLeaveTypeConfig(ctx, cfg)
}

func (a *LeaveTypeConfigRepoAdapter) Delete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error {
	return a.DeleteLeaveTypeConfig(ctx, tenant, id)
}

func (a *LeaveTypeConfigRepoAdapter) IsCodeUsed(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, code string) (bool, error) {
	return a.IsLeaveTypeCodeUsed(ctx, tenant, societeID, code)
}

func (a *LeaveTypeConfigRepoAdapter) UpsertDefaults(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, templates []domain.LeaveTypeTemplate) error {
	return a.UpsertLeaveTypeDefaults(ctx, tenant, societeID, templates)
}

var _ ports.LeaveTypeConfigRepository = (*LeaveTypeConfigRepoAdapter)(nil)
