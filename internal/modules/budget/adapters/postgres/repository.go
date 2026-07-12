package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/budget/domain"
	"github.com/kore/kore/internal/modules/budget/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, b domain.Budget) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO budget.budgets (
			id, tenant_id, application_id, type,
			planned_days, planned_uo, planned_amount,
			consumed_days, consumed_uo, consumed_amount, currency
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			consumed_days = EXCLUDED.consumed_days,
			consumed_uo = EXCLUDED.consumed_uo,
			consumed_amount = EXCLUDED.consumed_amount
	`, b.ID, b.TenantID.UUID(), b.ApplicationID, string(b.Type),
		b.Planned.Days, b.Planned.UO, b.Planned.Amount,
		b.Consumed.Days, b.Consumed.UO, b.Consumed.Amount, b.Currency)
	return err
}

func (r *Repository) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Budget, error) {
	return r.scanBudget(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, type,
			planned_days, planned_uo, planned_amount,
			consumed_days, consumed_uo, consumed_amount, currency
		FROM budget.budgets WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) List(ctx context.Context, tenant kernel.TenantID) ([]domain.Budget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, application_id, type,
			planned_days, planned_uo, planned_amount,
			consumed_days, consumed_uo, consumed_amount, currency
		FROM budget.budgets
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Budget
	for rows.Next() {
		b, err := r.scanBudget(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) GetByApplication(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.Budget, error) {
	return r.scanBudget(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, type,
			planned_days, planned_uo, planned_amount,
			consumed_days, consumed_uo, consumed_amount, currency
		FROM budget.budgets WHERE tenant_id = $1 AND application_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, tenant.UUID(), appID))
}

func (r *Repository) FindDefaultByApplication(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.Budget, error) {
	return r.scanBudget(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, type,
			planned_days, planned_uo, planned_amount,
			consumed_days, consumed_uo, consumed_amount, currency
		FROM budget.budgets
		WHERE tenant_id = $1 AND application_id = $2 AND type = $3
		LIMIT 1
	`, tenant.UUID(), appID, string(domain.BudgetTypeDefault)))
}

func (r *Repository) SaveEstimate(ctx context.Context, e domain.Estimate) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO budget.estimates (id, tenant_id, budget_id, demand_id, effort_uo, effort_days, superseded)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET superseded = EXCLUDED.superseded
	`, e.ID, e.TenantID.UUID(), e.BudgetID, e.DemandID, e.Effort.UO, e.Effort.Days, e.Superseded)
	return err
}

func (r *Repository) SaveQuote(ctx context.Context, q domain.Quote) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO budget.quotes (id, tenant_id, budget_id, demand_id, amount, effort_uo, effort_days, supersedes_estimate_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, q.ID, q.TenantID.UUID(), q.BudgetID, q.DemandID, q.Amount, q.Effort.UO, q.Effort.Days, q.SupersedesEstimateID)
	return err
}

func (r *Repository) GetEstimate(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.Estimate, error) {
	var e domain.Estimate
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, budget_id, demand_id, effort_uo, effort_days, superseded
		FROM budget.estimates WHERE tenant_id = $1 AND demand_id = $2 AND superseded = FALSE
		ORDER BY created_at DESC LIMIT 1
	`, tenant.UUID(), demandID).Scan(&e.ID, &tenantID, &e.BudgetID, &e.DemandID, &e.Effort.UO, &e.Effort.Days, &e.Superseded)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Estimate{}, fmt.Errorf("estimate not found: %w", err)
		}
		return domain.Estimate{}, err
	}
	e.TenantID = kernel.NewTenantID(tenantID)
	return e, nil
}

func (r *Repository) GetQuote(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.Quote, error) {
	var q domain.Quote
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, budget_id, demand_id, amount, effort_uo, effort_days, supersedes_estimate_id
		FROM budget.quotes WHERE tenant_id = $1 AND demand_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, tenant.UUID(), demandID).Scan(&q.ID, &tenantID, &q.BudgetID, &q.DemandID, &q.Amount, &q.Effort.UO, &q.Effort.Days, &q.SupersedesEstimateID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Quote{}, fmt.Errorf("quote not found: %w", err)
		}
		return domain.Quote{}, err
	}
	q.TenantID = kernel.NewTenantID(tenantID)
	return q, nil
}

func (r *Repository) SaveConsumption(ctx context.Context, c domain.Consumption) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO budget.consumptions (
			id, tenant_id, budget_id, period_start, period_end, days, uo, amount, approved_at, approved_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, budget_id, period_start, period_end) DO UPDATE SET
			days = EXCLUDED.days,
			uo = EXCLUDED.uo,
			amount = EXCLUDED.amount,
			approved_at = EXCLUDED.approved_at,
			approved_by = EXCLUDED.approved_by
	`, c.ID, c.TenantID.UUID(), c.BudgetID, c.Period.Start, c.Period.End,
		c.Triple.Days, c.Triple.UO, c.Triple.Amount, c.ApprovedAt, c.ApprovedBy)
	return err
}

func (r *Repository) GetConsumption(ctx context.Context, tenant kernel.TenantID, budgetID uuid.UUID, period kernel.Period) (domain.Consumption, error) {
	var c domain.Consumption
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, budget_id, period_start, period_end, days, uo, amount, approved_at, approved_by
		FROM budget.consumptions
		WHERE tenant_id = $1 AND budget_id = $2 AND period_start = $3 AND period_end = $4
	`, tenant.UUID(), budgetID, period.Start, period.End).Scan(
		&c.ID, &tenantID, &c.BudgetID, &c.Period.Start, &c.Period.End,
		&c.Triple.Days, &c.Triple.UO, &c.Triple.Amount, &c.ApprovedAt, &c.ApprovedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Consumption{}, fmt.Errorf("consumption not found: %w", err)
		}
		return domain.Consumption{}, err
	}
	c.TenantID = kernel.NewTenantID(tenantID)
	return c, nil
}

func (r *Repository) HasDefaultBudget(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM budget.budgets
			WHERE tenant_id = $1 AND application_id = $2 AND type = $3
		)
	`, tenant.UUID(), appID, string(domain.BudgetTypeDefault)).Scan(&exists)
	return exists, err
}

func (r *Repository) Consumption(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, period kernel.Period) (domain.ConsumptionTriple, error) {
	budget, err := r.GetByApplication(ctx, tenant, appID)
	if err != nil {
		return domain.ConsumptionTriple{}, err
	}
	consumption, err := r.GetConsumption(ctx, tenant, budget.ID, period)
	if err != nil {
		return domain.ConsumptionTriple{}, err
	}
	return consumption.Triple, nil
}

func (r *Repository) scanBudget(row pgx.Row) (domain.Budget, error) {
	var b domain.Budget
	var tenantID uuid.UUID
	var budgetType string
	err := row.Scan(
		&b.ID, &tenantID, &b.ApplicationID, &budgetType,
		&b.Planned.Days, &b.Planned.UO, &b.Planned.Amount,
		&b.Consumed.Days, &b.Consumed.UO, &b.Consumed.Amount, &b.Currency,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Budget{}, fmt.Errorf("budget not found: %w", err)
		}
		return domain.Budget{}, err
	}
	b.TenantID = kernel.NewTenantID(tenantID)
	b.Type = domain.BudgetType(budgetType)
	b.Remaining = domain.ConsumptionTriple{
		Days:   b.Planned.Days - b.Consumed.Days,
		UO:     b.Planned.UO - b.Consumed.UO,
		Amount: b.Planned.Amount - b.Consumed.Amount,
	}
	return b, nil
}

var (
	_ ports.BudgetRepository = (*Repository)(nil)
	_ ports.BudgetReader     = (*Repository)(nil)
)
