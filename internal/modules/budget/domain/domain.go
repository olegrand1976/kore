package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrDefaultBudgetRequired = errors.New("default budget required")
	ErrBudgetAlreadyApproved = errors.New("budget consumption already approved")
	ErrQuoteReplacesEstimate = errors.New("quote supersedes estimate")
)

type BudgetType string

const (
	BudgetTypeDefault  BudgetType = "defaut"
	BudgetTypeSpecific BudgetType = "specifique"
)

type Effort struct {
	Days float64
	UO   float64
}

type ConsumptionTriple struct {
	Days   float64
	UO     float64
	Amount int64
}

type Budget struct {
	ID            uuid.UUID
	TenantID      kernel.TenantID
	ApplicationID uuid.UUID
	Type          BudgetType
	Planned       ConsumptionTriple
	Consumed      ConsumptionTriple
	Remaining     ConsumptionTriple
	Currency      string
}

type Estimate struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	BudgetID  uuid.UUID
	DemandID  uuid.UUID
	Effort    Effort
	Superseded bool
}

type Quote struct {
	ID                  uuid.UUID
	TenantID            kernel.TenantID
	BudgetID            uuid.UUID
	DemandID            uuid.UUID
	Amount              int64
	Effort              Effort
	SupersedesEstimateID *uuid.UUID
}

type Consumption struct {
	ID         uuid.UUID
	TenantID   kernel.TenantID
	BudgetID   uuid.UUID
	Period     kernel.Period
	Triple     ConsumptionTriple
	ApprovedAt *time.Time
	ApprovedBy *uuid.UUID
}

func NewBudget(tenant kernel.TenantID, appID uuid.UUID, budgetType BudgetType, planned ConsumptionTriple, currency string) Budget {
	return Budget{
		ID:            uuid.New(),
		TenantID:      tenant,
		ApplicationID: appID,
		Type:          budgetType,
		Planned:       planned,
		Remaining:     planned,
		Currency:      currency,
	}
}

func (b *Budget) ApplyConsumption(triple ConsumptionTriple) {
	b.Consumed = ConsumptionTriple{
		Days:   b.Consumed.Days + triple.Days,
		UO:     b.Consumed.UO + triple.UO,
		Amount: b.Consumed.Amount + triple.Amount,
	}
	b.Remaining = ConsumptionTriple{
		Days:   b.Planned.Days - b.Consumed.Days,
		UO:     b.Planned.UO - b.Consumed.UO,
		Amount: b.Planned.Amount - b.Consumed.Amount,
	}
}

func (b Budget) IsOverrun() bool {
	return b.Remaining.Days < 0 || b.Remaining.UO < 0 || b.Remaining.Amount < 0
}

// EffectiveEffort returns quote effort if present, otherwise estimate effort.
func EffectiveEffort(estimate *Estimate, quote *Quote) Effort {
	if quote != nil {
		return quote.Effort
	}
	if estimate != nil && !estimate.Superseded {
		return estimate.Effort
	}
	return Effort{}
}

func AggregateConsumption(items []ConsumptionTriple) ConsumptionTriple {
	var total ConsumptionTriple
	for _, item := range items {
		total.Days += item.Days
		total.UO += item.UO
		total.Amount += item.Amount
	}
	return total
}
