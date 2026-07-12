package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, req domain.LeaveRequest) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO conges.leave_requests (
			id, tenant_id, user_id, type, start_date, end_date, motif, status, decided_by, decided_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			decided_by = EXCLUDED.decided_by,
			decided_at = EXCLUDED.decided_at
	`, req.ID, req.TenantID.UUID(), req.UserID, string(req.Type),
		req.Period.From, req.Period.To, req.Motif, string(req.Status), req.DecidedBy, req.DecidedAt)
	return err
}

func (r *Repository) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.LeaveRequest, error) {
	return r.scanRequest(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, type, start_date, end_date, motif, status, decided_by, decided_at
		FROM conges.leave_requests WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) ListByUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveRequest, error) {
	return r.list(ctx, tenant, &userID, nil)
}

func (r *Repository) List(ctx context.Context, tenant kernel.TenantID, userID *uuid.UUID, status *domain.LeaveStatus) ([]domain.LeaveRequest, error) {
	return r.list(ctx, tenant, userID, status)
}

func (r *Repository) list(ctx context.Context, tenant kernel.TenantID, userID *uuid.UUID, status *domain.LeaveStatus) ([]domain.LeaveRequest, error) {
	query := `
		SELECT id, tenant_id, user_id, type, start_date, end_date, motif, status, decided_by, decided_at
		FROM conges.leave_requests WHERE tenant_id = $1`
	args := []any{tenant.UUID()}
	argPos := 2
	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, string(*status))
	}
	query += " ORDER BY start_date DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.LeaveRequest
	for rows.Next() {
		req, err := r.scanRequestRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, req)
	}
	return out, rows.Err()
}

func (r *Repository) ListBalances(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveBalance, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, user_id, type, acquired, taken, remaining
		FROM conges.leave_balances WHERE tenant_id = $1 AND user_id = $2
	`, tenant.UUID(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.LeaveBalance
	for rows.Next() {
		var b domain.LeaveBalance
		var tenantID uuid.UUID
		var leaveType string
		if err := rows.Scan(&b.ID, &tenantID, &b.UserID, &leaveType, &b.Acquired, &b.Taken, &b.Remaining); err != nil {
			return nil, err
		}
		b.TenantID = kernel.NewTenantID(tenantID)
		b.Type = domain.LeaveType(leaveType)
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) scanRequest(row pgx.Row) (domain.LeaveRequest, error) {
	return r.scanRequestRow(row)
}

func (r *Repository) scanRequestRow(row pgx.Row) (domain.LeaveRequest, error) {
	var req domain.LeaveRequest
	var tenantID uuid.UUID
	var leaveType, status string
	var start, end time.Time
	err := row.Scan(&req.ID, &tenantID, &req.UserID, &leaveType, &start, &end, &req.Motif, &status, &req.DecidedBy, &req.DecidedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LeaveRequest{}, fmt.Errorf("leave request not found: %w", err)
		}
		return domain.LeaveRequest{}, err
	}
	req.TenantID = kernel.NewTenantID(tenantID)
	req.Type = domain.LeaveType(leaveType)
	req.Status = domain.LeaveStatus(status)
	req.Period = kernel.DateRange{From: start, To: end}
	return req, nil
}

var _ ports.LeaveRepository = (*Repository)(nil)
