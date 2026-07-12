package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/pkg/kernel"
)

type NotificationEvent struct {
	TenantID kernel.TenantID
	Trigger  string
	Subject  string
	Vars     map[string]string
}

type TransactionalMessage struct {
	Subject       string
	Body          string
	Recipients    []string
	Attachments   []domain.Attachment
	SkipSignature bool
}

type SentFilter struct {
	TenantID kernel.TenantID
	Status   *domain.MessageStatus
	Limit    int
}

type Email struct {
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []domain.Attachment
}

type NotificationService interface {
	DefineRule(ctx context.Context, rule domain.NotificationRule) error
	ListRules(ctx context.Context, tenant kernel.TenantID) ([]domain.NotificationRule, error)
	Publish(ctx context.Context, evt NotificationEvent) error
	ListSent(ctx context.Context, filter SentFilter) ([]domain.NotificationMessage, error)
	ProcessPending(ctx context.Context) (int, error)
}

type NotificationPublisher interface {
	Notify(ctx context.Context, evt NotificationEvent) error
}

type TransactionalNotifier interface {
	NotifyTransactional(ctx context.Context, msg TransactionalMessage) error
}

type NotificationRepository interface {
	SaveRule(ctx context.Context, r domain.NotificationRule) error
	GetRuleByTrigger(ctx context.Context, tenant kernel.TenantID, trigger string) (domain.NotificationRule, error)
	ListRules(ctx context.Context, tenant kernel.TenantID) ([]domain.NotificationRule, error)
	SaveMessage(ctx context.Context, m domain.NotificationMessage) error
	ListMessages(ctx context.Context, filter SentFilter) ([]domain.NotificationMessage, error)
	ListPending(ctx context.Context, tenant kernel.TenantID) ([]domain.NotificationMessage, error)
	ListDue(ctx context.Context, now time.Time, limit int) ([]domain.NotificationMessage, error)
}

type EmailSender interface {
	Send(ctx context.Context, msg Email) error
}

type RecipientResolver interface {
	ResolveUserEmails(ctx context.Context, tenant kernel.TenantID, userIDs []uuid.UUID) ([]string, error)
}

type Clock interface {
	Now() time.Time
}
