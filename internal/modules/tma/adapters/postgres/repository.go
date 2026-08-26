package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var existingID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM tma.demands WHERE id = $1
	`, d.ID).Scan(&existingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		ticketNumber := d.TicketNumber
		if ticketNumber <= 0 {
			ticketNumber, err = nextTicketNumber(ctx, tx, d.TenantID.UUID())
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO tma.demands (
				id, tenant_id, application_id, ticket_number, type, subject, description, priority, due_at,
				workflow_instance_id, author_id, assignee_id, taken_over_by_id, status, visible, consumption_active, requires_chef_gate,
				epic_id, sprint_id, story_points, backlog_rank, resolved_at, reopen_reason, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		`, d.ID, d.TenantID.UUID(), d.ApplicationID, ticketNumber, string(d.Type), d.Subject, d.Description, string(d.Priority), d.DueAt,
			d.WorkflowInstanceID, d.AuthorID, d.AssigneeID, d.TakenOverByID, string(d.Status), d.Visible, d.ConsumptionActive, d.RequiresChefGate,
			d.EpicID, d.SprintID, d.StoryPoints, d.BacklogRank, d.ResolvedAt, d.ReopenReason, d.CreatedAt)
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE tma.demands SET
				assignee_id = $2,
				taken_over_by_id = $3,
				status = $4,
				visible = $5,
				consumption_active = $6,
				workflow_instance_id = $7,
				description = $8,
				priority = $9,
				due_at = $10,
				epic_id = $11,
				sprint_id = $12,
				story_points = $13,
				backlog_rank = $14,
				resolved_at = $15,
				reopen_reason = $16
			WHERE id = $1
		`, d.ID, d.AssigneeID, d.TakenOverByID, string(d.Status), d.Visible, d.ConsumptionActive,
			d.WorkflowInstanceID, d.Description, string(d.Priority), d.DueAt,
			d.EpicID, d.SprintID, d.StoryPoints, d.BacklogRank, d.ResolvedAt, d.ReopenReason)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func nextTicketNumber(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (int, error) {
	var n int
	err := tx.QueryRow(ctx, `
		INSERT INTO tma.tenant_ticket_counters (tenant_id, last_number)
		VALUES ($1, 1)
		ON CONFLICT (tenant_id) DO UPDATE
		SET last_number = tma.tenant_ticket_counters.last_number + 1
		RETURNING last_number
	`, tenantID).Scan(&n)
	return n, err
}

func (r *Repository) EnsureTicketNumbers(ctx context.Context, tenant kernel.TenantID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id FROM tma.demands
		WHERE tenant_id = $1 AND ticket_number IS NULL
		ORDER BY created_at ASC, id ASC
		FOR UPDATE
	`, tenant.UUID())
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	for _, id := range ids {
		n, err := nextTicketNumber(ctx, tx, tenant.UUID())
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE tma.demands SET ticket_number = $3
			WHERE tenant_id = $1 AND id = $2 AND ticket_number IS NULL
		`, tenant.UUID(), id, n); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Demand, error) {
	return r.scanDemand(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, ticket_number, type, subject, description, priority, due_at,
			workflow_instance_id, author_id, assignee_id, taken_over_by_id, status, visible, consumption_active, requires_chef_gate,
			epic_id, sprint_id, story_points, backlog_rank, resolved_at, reopen_reason, created_at
		FROM tma.demands WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), id))
}

func (r *Repository) SoftDelete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID, deletedAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE tma.demands
		SET deleted_at = $3
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), id, deletedAt.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDemandNotFound
	}
	return nil
}

func demandSelectColumns() string {
	return `id, tenant_id, application_id, ticket_number, type, subject, description, priority, due_at,
			workflow_instance_id, author_id, assignee_id, taken_over_by_id, status, visible, consumption_active, requires_chef_gate,
			epic_id, sprint_id, story_points, backlog_rank, resolved_at, reopen_reason, created_at`
}

func (r *Repository) List(ctx context.Context, tenant kernel.TenantID, filter ports.ExportFilter) ([]domain.Demand, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM tma.demands WHERE tenant_id = $1 AND deleted_at IS NULL`, demandSelectColumns())
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
	if filter.SprintID != nil {
		query += fmt.Sprintf(" AND sprint_id = $%d", argPos)
		args = append(args, *filter.SprintID)
		argPos++
	}
	if filter.EpicID != nil {
		query += fmt.Sprintf(" AND epic_id = $%d", argPos)
		args = append(args, *filter.EpicID)
	}
	if filter.BacklogOnly {
		query += " AND sprint_id IS NULL AND status NOT IN ('resolue')"
	}
	if filter.VisibleOnly {
		query += " AND visible = TRUE"
	}
	query += " ORDER BY ticket_number ASC NULLS LAST, backlog_rank NULLS LAST, created_at DESC"

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
	var demandType, status, priority string
	var ticketNumber *int
	err := row.Scan(
		&d.ID, &tenantID, &d.ApplicationID, &ticketNumber, &demandType, &d.Subject, &d.Description, &priority, &d.DueAt,
		&d.WorkflowInstanceID, &d.AuthorID, &d.AssigneeID, &d.TakenOverByID, &status, &d.Visible, &d.ConsumptionActive, &d.RequiresChefGate,
		&d.EpicID, &d.SprintID, &d.StoryPoints, &d.BacklogRank, &d.ResolvedAt, &d.ReopenReason, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Demand{}, domain.ErrDemandNotFound
		}
		return domain.Demand{}, err
	}
	d.TenantID = kernel.NewTenantID(tenantID)
	d.Type = domain.DemandType(demandType)
	d.Status = domain.DemandStatus(status)
	d.Priority = kernel.RequestPriority(priority)
	if ticketNumber != nil {
		d.TicketNumber = *ticketNumber
	}
	return d, nil
}

var _ ports.DemandRepository = (*Repository)(nil)
