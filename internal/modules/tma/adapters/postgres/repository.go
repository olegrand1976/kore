package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, d domain.Demand) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tma.demands (
			id, tenant_id, application_id, type, subject, workflow_instance_id,
			author_id, assignee_id, status, visible, consumption_active, requires_chef_gate, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			assignee_id = EXCLUDED.assignee_id,
			status = EXCLUDED.status,
			visible = EXCLUDED.visible,
			consumption_active = EXCLUDED.consumption_active,
			workflow_instance_id = EXCLUDED.workflow_instance_id
	`, d.ID, d.TenantID.UUID(), d.ApplicationID, string(d.Type), d.Subject, d.WorkflowInstanceID,
		d.AuthorID, d.AssigneeID, string(d.Status), d.Visible, d.ConsumptionActive, d.RequiresChefGate, d.CreatedAt)
	return err
}

func (r *Repository) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Demand, error) {
	return r.scanDemand(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, type, subject, workflow_instance_id,
			author_id, assignee_id, status, visible, consumption_active, requires_chef_gate, created_at
		FROM tma.demands WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) List(ctx context.Context, tenant kernel.TenantID, filter ports.ExportFilter) ([]domain.Demand, error) {
	query := `
		SELECT id, tenant_id, application_id, type, subject, workflow_instance_id,
			author_id, assignee_id, status, visible, consumption_active, requires_chef_gate, created_at
		FROM tma.demands WHERE tenant_id = $1`
	args := []any{tenant.UUID()}
	argPos := 2
	if filter.ApplicationID != nil {
		query += fmt.Sprintf(" AND application_id = $%d", argPos)
		args = append(args, *filter.ApplicationID)
		argPos++
	}
	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, string(*filter.Status))
		argPos++
	}
	if filter.VisibleOnly {
		query += " AND visible = TRUE"
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Demand
	for rows.Next() {
		d, err := r.scanDemandRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repository) GetAnalysis(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.AnalysisDossier, error) {
	var dossier domain.AnalysisDossier
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, demand_id, functional, technical, risks, test_scenario
		FROM tma.analysis_dossiers
		WHERE tenant_id = $1 AND demand_id = $2
		ORDER BY id DESC
		LIMIT 1
	`, tenant.UUID(), demandID).Scan(
		&dossier.ID, &tenantID, &dossier.DemandID,
		&dossier.Functional, &dossier.Technical, &dossier.Risks, &dossier.TestScenario,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AnalysisDossier{}, domain.ErrAnalysisNotFound
		}
		return domain.AnalysisDossier{}, err
	}
	dossier.TenantID = kernel.NewTenantID(tenantID)
	return dossier, nil
}

func (r *Repository) SaveAnalysis(ctx context.Context, dossier domain.AnalysisDossier) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tma.analysis_dossiers (id, tenant_id, demand_id, functional, technical, risks, test_scenario)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (demand_id) DO UPDATE SET
			functional = EXCLUDED.functional,
			technical = EXCLUDED.technical,
			risks = EXCLUDED.risks,
			test_scenario = EXCLUDED.test_scenario
	`, dossier.ID, dossier.TenantID.UUID(), dossier.DemandID,
		dossier.Functional, dossier.Technical, dossier.Risks, dossier.TestScenario)
	return err
}

func (r *Repository) scanDemand(row pgx.Row) (domain.Demand, error) {
	return r.scanDemandRow(row)
}

func (r *Repository) scanDemandRow(row pgx.Row) (domain.Demand, error) {
	var d domain.Demand
	var tenantID uuid.UUID
	var demandType, status string
	err := row.Scan(
		&d.ID, &tenantID, &d.ApplicationID, &demandType, &d.Subject, &d.WorkflowInstanceID,
		&d.AuthorID, &d.AssigneeID, &status, &d.Visible, &d.ConsumptionActive, &d.RequiresChefGate, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Demand{}, fmt.Errorf("demand not found: %w", err)
		}
		return domain.Demand{}, err
	}
	d.TenantID = kernel.NewTenantID(tenantID)
	d.Type = domain.DemandType(demandType)
	d.Status = domain.DemandStatus(status)
	return d, nil
}

var _ ports.DemandRepository = (*Repository)(nil)
