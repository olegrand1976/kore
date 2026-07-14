package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo     ports.DemandRepository
	workflow ports.WorkflowService
	cra      ports.CRAFeeder
	budget   ports.BudgetReader
	notifier ports.NotificationPublisher
	clock    ports.Clock
}

func NewService(
	repo ports.DemandRepository,
	workflow ports.WorkflowService,
	cra ports.CRAFeeder,
	budget ports.BudgetReader,
	opts ...Option,
) ports.TMAService {
	s := &service{
		repo:     repo,
		workflow: workflow,
		cra:      cra,
		budget:   budget,
		clock:    realClock{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Option func(*service)

func WithNotifier(notifier ports.NotificationPublisher) Option {
	return func(s *service) { s.notifier = notifier }
}

func WithClock(clock ports.Clock) Option {
	return func(s *service) { s.clock = clock }
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (s *service) CreateDemand(ctx context.Context, cmd ports.CreateDemandCommand) (domain.Demand, error) {
	hasBudget, err := s.budget.HasDefaultBudget(ctx, cmd.TenantID, cmd.ApplicationID)
	if err != nil {
		return domain.Demand{}, err
	}
	if !hasBudget {
		return domain.Demand{}, domain.ErrDefaultBudgetRequired
	}
	demand := domain.NewDemand(cmd.TenantID, cmd.ApplicationID, cmd.AuthorID, cmd.Subject, cmd.Description, kernel.NormalizeRequestPriority(cmd.Priority), cmd.DueAt, cmd.RequiresChefGate)
	if s.workflow != nil {
		inst, err := s.workflow.Start(ctx, ports.StartWorkflowCommand{
			TenantID:       cmd.TenantID,
			DefinitionCode: "tma.incident",
			EntityID:       demand.ID.String(),
		})
		if err != nil {
			return domain.Demand{}, err
		}
		demand.WorkflowInstanceID = &inst.ID
	}
	if err := s.repo.Save(ctx, demand); err != nil {
		return domain.Demand{}, err
	}
	return demand, nil
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Demand, error) {
	return s.repo.Get(ctx, tenant, id)
}

func (s *service) GetAnalysis(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.AnalysisDossier, error) {
	return s.repo.GetAnalysis(ctx, tenant, demandID)
}

func (s *service) ValidateCreation(ctx context.Context, cmd ports.ChefUtilisateurCommand) error {
	demand, err := s.repo.Get(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return err
	}
	if err := demand.ValidateCreation(); err != nil {
		return err
	}
	if s.workflow != nil && demand.WorkflowInstanceID != nil {
		_, err = s.workflow.Fire(ctx, ports.FireTransitionCommand{
			TenantID:   cmd.TenantID,
			InstanceID: *demand.WorkflowInstanceID,
			Action:     "validate_creation",
			ActorID:    cmd.ActorID,
		})
		if err != nil {
			return err
		}
	}
	return s.repo.Save(ctx, demand)
}

func (s *service) Assign(ctx context.Context, cmd ports.AssignCommand) error {
	demand, err := s.repo.Get(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return err
	}
	if err := demand.Assign(cmd.AssigneeID); err != nil {
		return err
	}
	if s.workflow != nil && demand.WorkflowInstanceID != nil {
		_, err = s.workflow.Fire(ctx, ports.FireTransitionCommand{
			TenantID:   cmd.TenantID,
			InstanceID: *demand.WorkflowInstanceID,
			Action:     "assign",
			ActorID:    cmd.ActorID,
		})
		if err != nil {
			return err
		}
	}
	return s.repo.Save(ctx, demand)
}

func (s *service) TakeOver(ctx context.Context, tenant kernel.TenantID, id, userID uuid.UUID) error {
	demand, err := s.repo.Get(ctx, tenant, id)
	if err != nil {
		return err
	}
	if err := demand.TakeOver(userID); err != nil {
		return err
	}
	return s.repo.Save(ctx, demand)
}

func (s *service) AddAnalysis(ctx context.Context, cmd ports.AnalysisCommand) error {
	demand, err := s.repo.Get(ctx, cmd.TenantID, cmd.DemandID)
	if err != nil {
		return err
	}
	if !demand.Visible {
		return domain.ErrDemandNotVisible
	}
	dossier := domain.AnalysisDossier{
		ID:           uuid.New(),
		TenantID:     cmd.TenantID,
		DemandID:     cmd.DemandID,
		Functional:   cmd.Functional,
		Technical:    cmd.Technical,
		Risks:        cmd.Risks,
		TestScenario: cmd.TestScenario,
	}
	if err := s.repo.SaveAnalysis(ctx, dossier); err != nil {
		return err
	}
	if s.notifier != nil {
		_ = s.notifier.Notify(ctx, ports.NotificationEvent{
			TenantID: cmd.TenantID,
			Trigger:  "tma.analysis.updated",
			Subject:  "TMA — analyse mise à jour",
			Vars: map[string]string{
				"demandId":  cmd.DemandID.String(),
				"subject":   demand.Subject,
				"authorId":  cmd.ActorID.String(),
				"status":    string(demand.Status),
				"entityId":  demand.ID.String(),
				"entityTyp": "tma.demand",
			},
		})
	}
	return nil
}

func (s *service) Resolve(ctx context.Context, tenant kernel.TenantID, id, userID uuid.UUID) error {
	demand, err := s.repo.Get(ctx, tenant, id)
	if err != nil {
		return err
	}
	if err := demand.Resolve(); err != nil {
		return err
	}
	if s.cra != nil && demand.AssigneeID != nil {
		duration, derr := kernel.NewDuration(480)
		if derr != nil {
			return derr
		}
		_ = s.cra.ProposeLines(ctx, []ports.ProposedLine{{
			TenantID:   tenant,
			UserID:     *demand.AssigneeID,
			SourceType: "tma",
			SourceID:   demand.ID,
			Day:        s.clock.Now().UTC().Truncate(24 * time.Hour),
			Duration:   duration,
			Comment:    demand.Subject,
		}})
	}
	if s.workflow != nil && demand.WorkflowInstanceID != nil {
		_, err = s.workflow.Fire(ctx, ports.FireTransitionCommand{
			TenantID:   tenant,
			InstanceID: *demand.WorkflowInstanceID,
			Action:     "resolve",
			ActorID:    userID,
		})
		if err != nil {
			return err
		}
	}
	if s.notifier != nil && demand.AssigneeID != nil {
		_ = s.notifier.Notify(ctx, ports.NotificationEvent{
			TenantID: tenant,
			UserID:   *demand.AssigneeID,
			Subject:  "Demande TMA résolue",
			Body:     demand.Subject,
		})
	}
	return s.repo.Save(ctx, demand)
}

func (s *service) Reopen(ctx context.Context, cmd ports.ReworkCommand) error {
	demand, err := s.repo.Get(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return err
	}
	if err := demand.Reopen(cmd.Reason); err != nil {
		return err
	}
	if s.workflow != nil && demand.WorkflowInstanceID != nil {
		_, err = s.workflow.Fire(ctx, ports.FireTransitionCommand{
			TenantID:   cmd.TenantID,
			InstanceID: *demand.WorkflowInstanceID,
			Action:     "reopen",
			ActorID:    cmd.ActorID,
		})
		if err != nil {
			return err
		}
	}
	return s.repo.Save(ctx, demand)
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID, filter ports.ExportFilter) ([]domain.Demand, error) {
	filter.TenantID = tenant
	return s.repo.List(ctx, tenant, filter)
}

func (s *service) ExportXML(ctx context.Context, filter ports.ExportFilter) ([]domain.XmlExportRow, error) {
	demands, err := s.List(ctx, filter.TenantID, filter)
	if err != nil {
		return nil, err
	}
	rows := make([]domain.XmlExportRow, 0, len(demands))
	for _, d := range demands {
		rows = append(rows, domain.ToXmlExportRow(d))
	}
	return rows, nil
}
