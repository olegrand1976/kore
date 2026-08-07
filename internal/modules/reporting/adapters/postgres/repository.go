package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetReportDefinition(ctx context.Context, tenant kernel.TenantID, code string) (domain.ReportDefinition, error) {
	return r.scanReportDefinition(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, code, name, config, active, created_at
		FROM reporting.report_definitions WHERE tenant_id = $1 AND code = $2
	`, tenant.UUID(), code))
}

func (r *Repository) ListReportDefinitions(ctx context.Context, tenant kernel.TenantID) ([]domain.ReportDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, code, name, config, active, created_at
		FROM reporting.report_definitions WHERE tenant_id = $1 AND active = TRUE ORDER BY name
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ReportDefinition
	for rows.Next() {
		def, err := r.scanReportDefinition(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, def)
	}
	return out, rows.Err()
}

func (r *Repository) GetDashboardSnapshot(ctx context.Context, tenant kernel.TenantID, code string) (domain.Dashboard, error) {
	var dash domain.Dashboard
	var payload []byte
	err := r.pool.QueryRow(ctx, `
		SELECT dashboard_code, period_start, period_end, payload, computed_at
		FROM reporting.dashboard_snapshots
		WHERE tenant_id = $1 AND dashboard_code = $2
		ORDER BY computed_at DESC LIMIT 1
	`, tenant.UUID(), code).Scan(&dash.Code, &dash.Period.Start, &dash.Period.End, &payload, &dash.ComputedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dashboard{}, domain.ErrDashboardNotFound
		}
		return domain.Dashboard{}, err
	}
	dash.Payload = decodeJSON(payload)
	return dash, nil
}

func (r *Repository) UpsertDashboardSnapshot(ctx context.Context, tenant kernel.TenantID, dash domain.Dashboard) error {
	payload, err := json.Marshal(dash.Payload)
	if err != nil {
		return err
	}
	if payload == nil {
		payload = []byte("{}")
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO reporting.dashboard_snapshots (
			id, tenant_id, dashboard_code, period_start, period_end, payload, computed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, dashboard_code, period_start, period_end)
		DO UPDATE SET
			payload = EXCLUDED.payload,
			computed_at = EXCLUDED.computed_at
	`, uuid.New(), tenant.UUID(), dash.Code, dash.Period.Start, dash.Period.End, payload, dash.ComputedAt)
	return err
}

func (r *Repository) ListTenantIDsForSnapshotRefresh(ctx context.Context) ([]kernel.TenantID, error) {
	// Tenants with an active-ish subscription and CRA entitlement (cra dashboard payload).
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT e.tenant_id
		FROM billing.module_entitlements e
		INNER JOIN billing.subscriptions s ON s.tenant_id = e.tenant_id
		WHERE e.module_code = 'cra'
		  AND e.enabled = TRUE
		  AND s.status IN ('trial', 'active', 'past_due')
		ORDER BY e.tenant_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []kernel.TenantID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, kernel.NewTenantID(id))
	}
	return out, rows.Err()
}

func (r *Repository) scanReportDefinition(row pgx.Row) (domain.ReportDefinition, error) {
	var def domain.ReportDefinition
	var tenantID uuid.UUID
	var config []byte
	err := row.Scan(&def.ID, &tenantID, &def.Code, &def.Name, &config, &def.Active, &def.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ReportDefinition{}, domain.ErrReportNotFound
		}
		return domain.ReportDefinition{}, err
	}
	def.TenantID = kernel.NewTenantID(tenantID)
	def.Config = decodeJSON(config)
	return def, nil
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

var _ ports.ReportingRepository = (*Repository)(nil)
