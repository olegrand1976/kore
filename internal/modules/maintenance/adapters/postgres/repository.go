package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/internal/modules/maintenance/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveWorkRequest(ctx context.Context, wr domain.WorkRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO maintenance.work_requests (
			id, tenant_id, application_id, subject, description, priority, due_at, state, assignee_id,
			consumption_days, created_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			state = EXCLUDED.state,
			assignee_id = EXCLUDED.assignee_id,
			consumption_days = EXCLUDED.consumption_days,
			completed_at = EXCLUDED.completed_at,
			description = EXCLUDED.description,
			priority = EXCLUDED.priority,
			due_at = EXCLUDED.due_at
	`, wr.ID, wr.TenantID.UUID(), wr.ApplicationID, wr.Subject, wr.Description, string(wr.Priority), wr.DueAt, string(wr.State),
		wr.AssigneeID, wr.ConsumptionDays, wr.CreatedAt, wr.CompletedAt)
	return err
}

func (r *Repository) GetWorkRequest(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkRequest, error) {
	return r.scanWorkRequest(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, subject, description, priority, due_at, state, assignee_id,
			consumption_days, created_at, completed_at
		FROM maintenance.work_requests WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) ListWorkRequests(ctx context.Context, tenant kernel.TenantID) ([]domain.WorkRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, application_id, subject, description, priority, due_at, state, assignee_id,
			consumption_days, created_at, completed_at
		FROM maintenance.work_requests WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WorkRequest
	for rows.Next() {
		wr, err := r.scanWorkRequest(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, wr)
	}
	return out, rows.Err()
}

func (r *Repository) scanWorkRequest(row pgx.Row) (domain.WorkRequest, error) {
	var wr domain.WorkRequest
	var tenantID uuid.UUID
	var state, priority string
	err := row.Scan(&wr.ID, &tenantID, &wr.ApplicationID, &wr.Subject, &wr.Description, &priority, &wr.DueAt, &state,
		&wr.AssigneeID, &wr.ConsumptionDays, &wr.CreatedAt, &wr.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkRequest{}, domain.ErrWorkRequestNotFound
		}
		return domain.WorkRequest{}, err
	}
	wr.TenantID = kernel.NewTenantID(tenantID)
	wr.State = domain.WorkState(state)
	wr.Priority = kernel.RequestPriority(priority)
	return wr, nil
}

var _ ports.MaintenanceRepository = (*Repository)(nil)
