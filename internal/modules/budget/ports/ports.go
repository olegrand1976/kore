package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/budget/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateBudgetCommand struct {
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	Type          domain.BudgetType
	PlannedDays   float64
	PlannedUO     float64
	PlannedAmount int64
	Currency      string
}

type EstimateCommand struct {
	TenantID   kernel.TenantID
	BudgetID   uuid.UUID
	DemandID   uuid.UUID
	EffortUO   float64
	EffortDays float64
}

type QuoteCommand struct {
	TenantID             kernel.TenantID
	BudgetID             uuid.UUID
	DemandID             uuid.UUID
	Amount               int64
	EffortUO             float64
	EffortDays           float64
	SupersedesEstimateID *uuid.UUID
}

type ApproveConsumptionCommand struct {
	TenantID   kernel.TenantID
	BudgetID   uuid.UUID
	Period     kernel.Period
	ApprovedBy uuid.UUID
}

type CRAConsumption struct {
	Days   float64
	UO     float64
	Amount int64
}

type BudgetOverrun struct {
	TenantID      kernel.TenantID
	BudgetID      uuid.UUID
	ApplicationID uuid.UUID
	Remaining     domain.ConsumptionTriple
}

type BudgetService interface {
	CreateBudget(ctx context.Context, cmd CreateBudgetCommand) (domain.Budget, error)
	AddEstimate(ctx context.Context, cmd EstimateCommand) (domain.Estimate, error)
	AddQuote(ctx context.Context, cmd QuoteCommand) (domain.Quote, error)
	RecomputeConsumption(ctx context.Context, tenant kernel.TenantID, budgetID uuid.UUID, period kernel.Period) (domain.ConsumptionTriple, error)
	Get(ctx context.Context, tenant kernel.TenantID, budgetID uuid.UUID) (domain.Budget, error)
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.Budget, error)
	Approve(ctx context.Context, cmd ApproveConsumptionCommand) error
}

type BudgetReader interface {
	HasDefaultBudget(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (bool, error)
	Consumption(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, period kernel.Period) (domain.ConsumptionTriple, error)
}

type BudgetRepository interface {
	Save(ctx context.Context, b domain.Budget) error
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Budget, error)
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.Budget, error)
	GetByApplication(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.Budget, error)
	FindDefaultByApplication(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.Budget, error)
	SaveEstimate(ctx context.Context, e domain.Estimate) error
	SaveQuote(ctx context.Context, q domain.Quote) error
	GetEstimate(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.Estimate, error)
	GetQuote(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.Quote, error)
	SaveConsumption(ctx context.Context, c domain.Consumption) error
	GetConsumption(ctx context.Context, tenant kernel.TenantID, budgetID uuid.UUID, period kernel.Period) (domain.Consumption, error)
}

type CRAReader interface {
	ConsumedByApplication(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, period kernel.Period) ([]CRAConsumption, error)
}

type BudgetAlertPublisher interface {
	Publish(ctx context.Context, evt BudgetOverrun) error
}

type Clock interface {
	Now() time.Time
}
