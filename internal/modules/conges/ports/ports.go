package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/pkg/kernel"
)

type RequestLeaveCommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	Type     domain.LeaveType
	From     time.Time
	To       time.Time
	Motif    string
}

type DecideLeaveCommand struct {
	TenantID kernel.TenantID
	ID       uuid.UUID
	DecidedBy uuid.UUID
}

type ProposedLine struct {
	TenantID   kernel.TenantID
	UserID     uuid.UUID
	SourceType string
	SourceID   uuid.UUID
	Day        time.Time
	Duration   kernel.Duration
	Comment    string
}

type StartWorkflowCommand struct {
	TenantID       kernel.TenantID
	DefinitionCode string
	EntityID       string
}

type FireTransitionCommand struct {
	TenantID   kernel.TenantID
	InstanceID uuid.UUID
	Action     string
	ActorID    uuid.UUID
}

type WorkflowInstance struct {
	ID           uuid.UUID
	CurrentState string
}

type NotificationEvent struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	Subject  string
	Body     string
}

type LeaveService interface {
	Request(ctx context.Context, cmd RequestLeaveCommand) (domain.LeaveRequest, error)
	Approve(ctx context.Context, cmd DecideLeaveCommand) error
	Reject(ctx context.Context, cmd DecideLeaveCommand) error
	Balance(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveBalance, error)
	List(ctx context.Context, tenant kernel.TenantID, userID *uuid.UUID, status *domain.LeaveStatus) ([]domain.LeaveRequest, error)
}

type LeaveRepository interface {
	Save(ctx context.Context, r domain.LeaveRequest) error
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.LeaveRequest, error)
	ListByUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveRequest, error)
	List(ctx context.Context, tenant kernel.TenantID, userID *uuid.UUID, status *domain.LeaveStatus) ([]domain.LeaveRequest, error)
	ListBalances(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.LeaveBalance, error)
}

type CRAFeeder interface {
	ProposeLines(ctx context.Context, lines []ProposedLine) error
}

type WorkflowService interface {
	Start(ctx context.Context, cmd StartWorkflowCommand) (WorkflowInstance, error)
	Fire(ctx context.Context, cmd FireTransitionCommand) (WorkflowInstance, error)
}

type NotificationPublisher interface {
	Notify(ctx context.Context, evt NotificationEvent) error
}

type Clock interface {
	Now() time.Time
}
