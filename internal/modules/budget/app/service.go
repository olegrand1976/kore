package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/budget/domain"
	"github.com/kore/kore/internal/modules/budget/ports"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
)

const consumptionCacheTTL = 5 * time.Minute

type service struct {
	repo     ports.BudgetRepository
	cra      ports.CRAReader
	alerts   ports.BudgetAlertPublisher
	appCache cache.Cache
	keys     cache.KeyBuilder
	clock    ports.Clock
}

func NewService(repo ports.BudgetRepository, cra ports.CRAReader, opts ...Option) ports.BudgetService {
	s := &service{
		repo:  repo,
		cra:   cra,
		clock: realClock{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Option func(*service)

func WithAlerts(publisher ports.BudgetAlertPublisher) Option {
	return func(s *service) { s.alerts = publisher }
}

func WithCache(appCache cache.Cache, keys cache.KeyBuilder) Option {
	return func(s *service) {
		s.appCache = appCache
		s.keys = keys
	}
}

func WithClock(clock ports.Clock) Option {
	return func(s *service) { s.clock = clock }
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (s *service) CreateBudget(ctx context.Context, cmd ports.CreateBudgetCommand) (domain.Budget, error) {
	currency := cmd.Currency
	if currency == "" {
		currency = "EUR"
	}
	budget := domain.NewBudget(cmd.TenantID, cmd.ApplicationID, cmd.Type, domain.ConsumptionTriple{
		Days:   cmd.PlannedDays,
		UO:     cmd.PlannedUO,
		Amount: cmd.PlannedAmount,
	}, currency)
	return budget, s.repo.Save(ctx, budget)
}

func (s *service) AddEstimate(ctx context.Context, cmd ports.EstimateCommand) (domain.Estimate, error) {
	estimate := domain.Estimate{
		ID:       uuid.New(),
		TenantID: cmd.TenantID,
		BudgetID: cmd.BudgetID,
		DemandID: cmd.DemandID,
		Effort: domain.Effort{
			Days: cmd.EffortDays,
			UO:   cmd.EffortUO,
		},
	}
	return estimate, s.repo.SaveEstimate(ctx, estimate)
}

func (s *service) AddQuote(ctx context.Context, cmd ports.QuoteCommand) (domain.Quote, error) {
	quote := domain.Quote{
		ID:                   uuid.New(),
		TenantID:             cmd.TenantID,
		BudgetID:             cmd.BudgetID,
		DemandID:             cmd.DemandID,
		Amount:               cmd.Amount,
		Effort:               domain.Effort{Days: cmd.EffortDays, UO: cmd.EffortUO},
		SupersedesEstimateID: cmd.SupersedesEstimateID,
	}
	if cmd.SupersedesEstimateID != nil {
		estimate, err := s.repo.GetEstimate(ctx, cmd.TenantID, cmd.DemandID)
		if err == nil {
			estimate.Superseded = true
			_ = s.repo.SaveEstimate(ctx, estimate)
		}
	}
	return quote, s.repo.SaveQuote(ctx, quote)
}

func (s *service) RecomputeConsumption(ctx context.Context, tenant kernel.TenantID, budgetID uuid.UUID, period kernel.Period) (domain.ConsumptionTriple, error) {
	budget, err := s.repo.Get(ctx, tenant, budgetID)
	if err != nil {
		return domain.ConsumptionTriple{}, err
	}
	items, err := s.cra.ConsumedByApplication(ctx, tenant, budget.ApplicationID, period)
	if err != nil {
		return domain.ConsumptionTriple{}, err
	}
	var triple domain.ConsumptionTriple
	for _, item := range items {
		triple.Days += item.Days
		triple.UO += item.UO
		triple.Amount += item.Amount
	}
	budget.ApplyConsumption(triple)
	if err := s.repo.Save(ctx, budget); err != nil {
		return domain.ConsumptionTriple{}, err
	}
	consumption := domain.Consumption{
		ID:       uuid.New(),
		TenantID: tenant,
		BudgetID: budgetID,
		Period:   period,
		Triple:   triple,
	}
	if err := s.repo.SaveConsumption(ctx, consumption); err != nil {
		return domain.ConsumptionTriple{}, err
	}
	if budget.IsOverrun() && s.alerts != nil {
		_ = s.alerts.Publish(ctx, ports.BudgetOverrun{
			TenantID:      tenant,
			BudgetID:      budgetID,
			ApplicationID: budget.ApplicationID,
			Remaining:     budget.Remaining,
		})
	}
	s.invalidateConsumptionCache(ctx, tenant, budget.ApplicationID, period)
	return triple, nil
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, budgetID uuid.UUID) (domain.Budget, error) {
	return s.repo.Get(ctx, tenant, budgetID)
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID) ([]domain.Budget, error) {
	return s.repo.List(ctx, tenant)
}

func (s *service) Approve(ctx context.Context, cmd ports.ApproveConsumptionCommand) error {
	consumption, err := s.repo.GetConsumption(ctx, cmd.TenantID, cmd.BudgetID, cmd.Period)
	if err != nil {
		return err
	}
	if consumption.ApprovedAt != nil {
		return domain.ErrBudgetAlreadyApproved
	}
	now := s.clock.Now().UTC()
	consumption.ApprovedAt = &now
	consumption.ApprovedBy = &cmd.ApprovedBy
	if err := s.repo.SaveConsumption(ctx, consumption); err != nil {
		return err
	}
	budget, err := s.repo.Get(ctx, cmd.TenantID, cmd.BudgetID)
	if err != nil {
		return err
	}
	s.invalidateConsumptionCache(ctx, cmd.TenantID, budget.ApplicationID, cmd.Period)
	return nil
}

func (s *service) HasDefaultBudget(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (bool, error) {
	_, err := s.repo.FindDefaultByApplication(ctx, tenant, appID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *service) Consumption(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, period kernel.Period) (domain.ConsumptionTriple, error) {
	if s.appCache != nil && s.keys != nil {
		key := s.keys.Key(tenant, "budget", "consumption", appID.String(), period.Start.Format("2006-01"), period.End.Format("2006-01"))
		var triple domain.ConsumptionTriple
		err := s.appCache.GetOrLoad(ctx, key, consumptionCacheTTL, func(ctx context.Context) (any, error) {
			return s.loadConsumption(ctx, tenant, appID, period)
		}, &triple)
		return triple, err
	}
	return s.loadConsumption(ctx, tenant, appID, period)
}

func (s *service) loadConsumption(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, period kernel.Period) (domain.ConsumptionTriple, error) {
	budget, err := s.repo.GetByApplication(ctx, tenant, appID)
	if err != nil {
		return domain.ConsumptionTriple{}, err
	}
	consumption, err := s.repo.GetConsumption(ctx, tenant, budget.ID, period)
	if err != nil {
		return domain.ConsumptionTriple{}, err
	}
	return consumption.Triple, nil
}

func (s *service) invalidateConsumptionCache(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, period kernel.Period) {
	if s.appCache == nil || s.keys == nil {
		return
	}
	key := s.keys.Key(tenant, "budget", "consumption", appID.String(), period.Start.Format("2006-01"), period.End.Format("2006-01"))
	_ = s.appCache.Delete(ctx, key)
}

var (
	_ ports.BudgetService = (*service)(nil)
	_ ports.BudgetReader  = (*service)(nil)
)
