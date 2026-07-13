package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrConnectionNotFound = errors.New("integration connection not found")
	ErrApiKeyNotFound     = errors.New("api key not found")
	ErrConnectionInactive = errors.New("integration connection inactive")
)

type ConnectionType string

const (
	ConnectionTypeAccounting ConnectionType = "accounting"
	ConnectionTypePDP        ConnectionType = "pdp"
	ConnectionTypeHRIS       ConnectionType = "hris"
	ConnectionTypeCalendar   ConnectionType = "calendar"
)

type ConnectionStatus string

const (
	ConnectionStatusActive   ConnectionStatus = "active"
	ConnectionStatusError    ConnectionStatus = "error"
	ConnectionStatusDisabled ConnectionStatus = "disabled"
)

type IntegrationConnection struct {
	ID             uuid.UUID
	TenantID       kernel.TenantID
	Type           ConnectionType
	Provider       string
	Status         ConnectionStatus
	CredentialsRef string
	LastSyncAt     *time.Time
	CreatedAt      time.Time
}

type ApiKey struct {
	ID         uuid.UUID
	TenantID   kernel.TenantID
	Name       string
	KeyPrefix  string
	KeyHash    string
	RevokedAt  *time.Time
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

type WebhookSubscription struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	URL       string
	Events    []string
	SecretRef string
	Active    bool
	CreatedAt time.Time
}

type SyncJob struct {
	ID           uuid.UUID
	TenantID     kernel.TenantID
	ConnectionID uuid.UUID
	Status       string
	StartedAt    time.Time
	FinishedAt   *time.Time
	ErrorMessage string
}

func NewConnection(tenant kernel.TenantID, connType ConnectionType, provider, credentialsRef string) IntegrationConnection {
	return IntegrationConnection{
		ID:             uuid.New(),
		TenantID:       tenant,
		Type:           connType,
		Provider:       provider,
		Status:         ConnectionStatusActive,
		CredentialsRef: credentialsRef,
		CreatedAt:      time.Now().UTC(),
	}
}

func (c IntegrationConnection) CanSync() bool {
	return c.Status == ConnectionStatusActive
}

func NewApiKey(tenant kernel.TenantID, name, prefix, hash string) ApiKey {
	return ApiKey{
		ID:        uuid.New(),
		TenantID:  tenant,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   hash,
		CreatedAt: time.Now().UTC(),
	}
}

func (k ApiKey) IsRevoked() bool {
	return k.RevokedAt != nil
}
