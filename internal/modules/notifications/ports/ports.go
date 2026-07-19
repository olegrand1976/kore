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
	GetRuleByCode(ctx context.Context, tenant kernel.TenantID, code string) (domain.NotificationRule, error)
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
	ResolveEquipeUserEmails(ctx context.Context, tenant kernel.TenantID, equipeID uuid.UUID) ([]string, error)
	ResolveApplicationUserEmails(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) ([]string, error)
	ResolveServiceUserEmails(ctx context.Context, tenant kernel.TenantID, serviceID uuid.UUID) ([]string, error)
	ResolveTenantUserEmails(ctx context.Context, tenant kernel.TenantID) ([]string, error)
	ResolveEquipeUserIDs(ctx context.Context, tenant kernel.TenantID, equipeID uuid.UUID) ([]uuid.UUID, error)
	ResolveApplicationUserIDs(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) ([]uuid.UUID, error)
	ResolveServiceUserIDs(ctx context.Context, tenant kernel.TenantID, serviceID uuid.UUID) ([]uuid.UUID, error)
}

type Clock interface {
	Now() time.Time
}

type RegisterDeviceCommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	Platform string
	Token    string
}

type UnregisterDeviceCommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	Token    string
}

type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

type DeviceRepository interface {
	UpsertDeviceToken(ctx context.Context, token domain.DeviceToken) error
	DeleteDeviceToken(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, token string) error
	DeleteDeviceTokenByValue(ctx context.Context, tenant kernel.TenantID, token string) error
	ListDeviceTokens(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]domain.DeviceToken, error)
}

type PushSender interface {
	Send(ctx context.Context, tokens []string, msg PushMessage) error
}

type DeviceService interface {
	RegisterDevice(ctx context.Context, cmd RegisterDeviceCommand) error
	UnregisterDevice(ctx context.Context, cmd UnregisterDeviceCommand) error
}
