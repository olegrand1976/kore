package project

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	projectdomain "github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type WipChecker struct {
	pool *db.Pool
}

func NewWipChecker(pool *db.Pool) *WipChecker {
	return &WipChecker{pool: pool}
}

func (w *WipChecker) CheckWip(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, targetStatus string, excludeDemandID *uuid.UUID) error {
	var profile string
	err := w.pool.QueryRow(ctx, `
		SELECT methodology_profile FROM org.applications
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), appID).Scan(&profile)
	if err != nil {
		return err
	}
	if profile != "agile_kanban" {
		return nil
	}

	var raw []byte
	err = w.pool.QueryRow(ctx, `
		SELECT columns FROM project.kanban_configs
		WHERE tenant_id = $1 AND application_id = $2
	`, tenant.UUID(), appID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	var columns []projectdomain.KanbanColumn
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &columns); err != nil {
			return err
		}
	}

	var wipLimit *int
	for _, col := range columns {
		if col.StateCode == targetStatus {
			wipLimit = col.WipLimit
			break
		}
	}
	if wipLimit == nil || *wipLimit <= 0 {
		return nil
	}

	query := `
		SELECT COUNT(*) FROM tma.demands
		WHERE tenant_id = $1 AND application_id = $2 AND status = $3 AND visible = TRUE
	`
	args := []any{tenant.UUID(), appID, targetStatus}
	if excludeDemandID != nil {
		query += ` AND id <> $4`
		args = append(args, *excludeDemandID)
	}

	var count int
	if err := w.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count >= *wipLimit {
		return projectdomain.ErrWipLimitExceeded
	}
	return nil
}
