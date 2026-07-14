package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/pkg/kernel"
)

type ConnectCommand struct {
	TenantID       kernel.TenantID
	Type           domain.ConnectionType
	Provider       string
	CredentialsRef string
}

type SyncCommand struct {
	TenantID     kernel.TenantID
	ConnectionID uuid.UUID
}

type CreateApiKeyCommand struct {
	TenantID kernel.TenantID
	Name     string
}

type CreateWebhookCommand struct {
	TenantID kernel.TenantID
	URL      string
	Events   []string
	Secret   string
}

type OutboundEvent struct {
	ID         uuid.UUID
	TenantID   kernel.TenantID
	Type       string
	OccurredAt time.Time
	Data       map[string]any
}

type ApiKeyCreated struct {
	ApiKey   domain.ApiKey
	PlainKey string
}

type IntegrationService interface {
	Connect(ctx context.Context, cmd ConnectCommand) (domain.IntegrationConnection, error)
	Disconnect(ctx context.Context, tenant kernel.TenantID, connID uuid.UUID) error
	Sync(ctx context.Context, cmd SyncCommand) (domain.SyncJob, error)
	ListConnections(ctx context.Context, tenant kernel.TenantID) ([]domain.IntegrationConnection, error)
	GetConnection(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.IntegrationConnection, error)
	ListSyncLogs(ctx context.Context, tenant kernel.TenantID) ([]domain.SyncJob, error)
	CreateWebhook(ctx context.Context, cmd CreateWebhookCommand) (domain.WebhookSubscription, error)
	ListWebhooks(ctx context.Context, tenant kernel.TenantID) ([]domain.WebhookSubscription, error)
	DeleteWebhook(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
}

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, evt OutboundEvent) error
}

type ApiKeyService interface {
	CreateKey(ctx context.Context, cmd CreateApiKeyCommand) (ApiKeyCreated, error)
	RevokeKey(ctx context.Context, tenant kernel.TenantID, keyID uuid.UUID) error
	ListKeys(ctx context.Context, tenant kernel.TenantID) ([]domain.ApiKey, error)
}

type IntegrationRepository interface {
	SaveConnection(ctx context.Context, c domain.IntegrationConnection) error
	GetConnection(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.IntegrationConnection, error)
	ListConnections(ctx context.Context, tenant kernel.TenantID) ([]domain.IntegrationConnection, error)
	DeleteConnection(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
	SaveSyncJob(ctx context.Context, j domain.SyncJob) error
	ListSyncJobs(ctx context.Context, tenant kernel.TenantID) ([]domain.SyncJob, error)
	SaveApiKey(ctx context.Context, k domain.ApiKey) error
	GetApiKey(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.ApiKey, error)
	GetApiKeyByHash(ctx context.Context, keyHash string) (domain.ApiKey, error)
	ListApiKeys(ctx context.Context, tenant kernel.TenantID) ([]domain.ApiKey, error)
	SaveWebhookSubscription(ctx context.Context, sub domain.WebhookSubscription) error
	ListWebhookSubscriptions(ctx context.Context, tenant kernel.TenantID) ([]domain.WebhookSubscription, error)
	DeleteWebhookSubscription(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error
}
