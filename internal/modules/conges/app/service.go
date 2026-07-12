package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo     ports.LeaveRepository
	cra      ports.CRAFeeder
	workflow ports.WorkflowService
	notifier ports.NotificationPublisher
	clock    ports.Clock
}

func NewService(
	repo ports.LeaveRepository,
	cra ports.CRAFeeder,
	workflow ports.WorkflowService,
	opts ...Option,
) ports.LeaveService {
	s := &service{
		repo:     repo,
		cra:      cra,
		workflow: workflow,
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

func (s *service) Request(ctx context.Context, cmd ports.RequestLeaveCommand) (domain.LeaveRequest, error) {
	period, err := kernel.NewDateRange(cmd.From, cmd.To)
	if err != nil {
		return domain.LeaveRequest{}, domain.ErrInvalidDateRange
	}
	req := domain.NewLeaveRequest(cmd.TenantID, cmd.UserID, cmd.Type, period, cmd.Motif)
	if s.workflow != nil {
		_, err = s.workflow.Start(ctx, ports.StartWorkflowCommand{
			TenantID:       cmd.TenantID,
			DefinitionCode: "leave.request",
			EntityID:       req.ID.String(),
		})
		if err != nil {
			return domain.LeaveRequest{}, err
		}
	}
	if err := s.repo.Save(ctx, req); err != nil {
		return domain.LeaveRequest{}, err
	}
	return req, nil
}

func (s *service) Approve(ctx context.Context, cmd ports.DecideLeaveCommand) error {
	req, err := s.repo.Get(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if err := req.Approve(now, cmd.DecidedBy); err != nil {
		return err
	}
	futureDays := domain.FutureDays(req.Period, now)
	if len(futureDays) == 0 && req.Period.To.After(now) {
		// period includes today or past only — still valid approval, no CRA lines
	} else if s.cra != nil && len(futureDays) > 0 {
		lines := make([]ports.ProposedLine, 0, len(futureDays))
		duration, derr := kernel.NewDuration(480) // full day
		if derr != nil {
			return derr
		}
		for _, day := range futureDays {
			lines = append(lines, ports.ProposedLine{
				TenantID:   req.TenantID,
				UserID:     req.UserID,
				SourceType: "leave",
				SourceID:   req.ID,
				Day:        day,
				Duration:   duration,
				Comment:    req.Motif,
			})
		}
		if err := s.cra.ProposeLines(ctx, lines); err != nil {
			return err
		}
	}
	if s.workflow != nil {
		_, err = s.workflow.Fire(ctx, ports.FireTransitionCommand{
			TenantID:   cmd.TenantID,
			InstanceID: req.ID,
			Action:     "approve",
			ActorID:    cmd.DecidedBy,
		})
		if err != nil {
			return err
		}
	}
	return s.repo.Save(ctx, req)
}

func (s *service) Reject(ctx context.Context, cmd ports.DecideLeaveCommand) error {
	req, err := s.repo.Get(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return err
	}
	if err := req.Reject(s.clock.Now(), cmd.DecidedBy); err != nil {
		return err
	}
	if s.notifier != nil {
		_ = s.notifier.Notify(ctx, ports.NotificationEvent{
			TenantID: req.TenantID,
			UserID:   req.UserID,
			Subject:  "Demande de congé refusée",
			Body:     req.Motif,
		})
	}
	if s.workflow != nil {
		_, err = s.workflow.Fire(ctx, ports.FireTransitionCommand{
			TenantID:   cmd.TenantID,
			InstanceID: req.ID,
			Action:     "reject",
			ActorID:    cmd.DecidedBy,
		})
		if err != nil {
			return err
		}
	}
	return s.repo.Save(ctx, req)
}

func (s *service) Balance(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveBalance, error) {
	return s.repo.ListBalances(ctx, tenant, userID)
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID, userID *uuid.UUID, status *domain.LeaveStatus) ([]domain.LeaveRequest, error) {
	return s.repo.List(ctx, tenant, userID, status)
}
