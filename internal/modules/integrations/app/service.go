package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/adapters/calendar"
	"github.com/kore/kore/internal/modules/integrations/adapters/fec"
	"github.com/kore/kore/internal/modules/integrations/adapters/hris"
	"github.com/kore/kore/internal/modules/integrations/adapters/pennylane"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo       ports.IntegrationRepository
	fec        *fec.Exporter
	calendar   *calendar.StubGateway
	hris       *hris.StubGateway
	pennylane  *pennylane.Client
	webhooks   ports.WebhookDispatcher
}

func NewService(repo ports.IntegrationRepository, opts ...ServiceOption) ports.IntegrationService {
	s := &service{
		repo:     repo,
		fec:      fec.NewExporter(),
		calendar: calendar.NewStubGateway(slog.Default()),
		hris:     hris.NewStubGateway(slog.Default()),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type ServiceOption func(*service)

func WithCalendarGateway(gw *calendar.StubGateway) ServiceOption {
	return func(s *service) { s.calendar = gw }
}

func WithHRISGateway(gw *hris.StubGateway) ServiceOption {
	return func(s *service) { s.hris = gw }
}

func WithPennylaneClient(client *pennylane.Client) ServiceOption {
	return func(s *service) { s.pennylane = client }
}

func WithWebhookDispatcher(d ports.WebhookDispatcher) ServiceOption {
	return func(s *service) { s.webhooks = d }
}

func NewApiKeyService(repo ports.IntegrationRepository) ports.ApiKeyService {
	return &service{repo: repo, fec: fec.NewExporter(), calendar: calendar.NewStubGateway(slog.Default()), hris: hris.NewStubGateway(slog.Default())}
}

func (s *service) Connect(ctx context.Context, cmd ports.ConnectCommand) (domain.IntegrationConnection, error) {
	conn := domain.NewConnection(cmd.TenantID, cmd.Type, cmd.Provider, cmd.CredentialsRef)
	return conn, s.repo.SaveConnection(ctx, conn)
}

func (s *service) Disconnect(ctx context.Context, tenant kernel.TenantID, connID uuid.UUID) error {
	if _, err := s.repo.GetConnection(ctx, tenant, connID); err != nil {
		return err
	}
	return s.repo.DeleteConnection(ctx, tenant, connID)
}

func (s *service) Sync(ctx context.Context, cmd ports.SyncCommand) (domain.SyncJob, error) {
	conn, err := s.repo.GetConnection(ctx, cmd.TenantID, cmd.ConnectionID)
	if err != nil {
		return domain.SyncJob{}, err
	}
	if !conn.CanSync() {
		return domain.SyncJob{}, domain.ErrConnectionInactive
	}
	now := time.Now().UTC()
	job := domain.SyncJob{
		ID:           uuid.New(),
		TenantID:     cmd.TenantID,
		ConnectionID: cmd.ConnectionID,
		Status:       "completed",
		StartedAt:    now,
		FinishedAt:   &now,
	}
	if conn.Provider == "fec" {
		_, count, exportErr := s.fec.Export(ctx, cmd.TenantID, now.Format("2006-01"), 1)
		if exportErr != nil {
			job.Status = "failed"
			job.ErrorMessage = exportErr.Error()
		} else {
			job.ErrorMessage = fmt.Sprintf("fec export: %d records", count)
		}
	} else if conn.Provider == "pennylane" && s.pennylane != nil {
		count, syncErr := s.pennylane.SyncAccounting(ctx, cmd.TenantID, now.Format("2006-01"))
		if syncErr != nil {
			job.Status = "failed"
			job.ErrorMessage = syncErr.Error()
		} else {
			job.ErrorMessage = fmt.Sprintf("pennylane sync: %d records", count)
		}
	} else if conn.Type == domain.ConnectionTypeCalendar && (conn.Provider == "google" || conn.Provider == "googlecalendar") {
		count, syncErr := s.calendar.Sync(ctx, cmd.TenantID, conn.Provider)
		if syncErr != nil {
			job.Status = "failed"
			job.ErrorMessage = syncErr.Error()
		} else {
			job.ErrorMessage = fmt.Sprintf("calendar sync: %d records", count)
		}
	} else if conn.Type == domain.ConnectionTypeHRIS && conn.Provider == "lucca" {
		count, syncErr := s.hris.Sync(ctx, cmd.TenantID, conn.Provider)
		if syncErr != nil {
			job.Status = "failed"
			job.ErrorMessage = syncErr.Error()
		} else {
			job.ErrorMessage = fmt.Sprintf("hris sync: %d records", count)
		}
	}
	conn.LastSyncAt = &now
	if err := s.repo.SaveConnection(ctx, conn); err != nil {
		return domain.SyncJob{}, err
	}
	if err := s.repo.SaveSyncJob(ctx, job); err != nil {
		return domain.SyncJob{}, err
	}
	if job.Status == "completed" && s.webhooks != nil {
		_ = s.webhooks.Dispatch(ctx, ports.OutboundEvent{
			ID:         job.ID,
			TenantID:   cmd.TenantID,
			Type:       "integration.sync.completed",
			OccurredAt: now,
			Data: map[string]any{
				"connectionId": cmd.ConnectionID.String(),
				"provider":     conn.Provider,
				"message":      job.ErrorMessage,
			},
		})
	}
	return job, nil
}

func (s *service) ListConnections(ctx context.Context, tenant kernel.TenantID) ([]domain.IntegrationConnection, error) {
	return s.repo.ListConnections(ctx, tenant)
}

func (s *service) GetConnection(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.IntegrationConnection, error) {
	return s.repo.GetConnection(ctx, tenant, id)
}

func (s *service) ListSyncLogs(ctx context.Context, tenant kernel.TenantID) ([]domain.SyncJob, error) {
	return s.repo.ListSyncJobs(ctx, tenant)
}

func (s *service) CreateWebhook(ctx context.Context, cmd ports.CreateWebhookCommand) (domain.WebhookSubscription, error) {
	secret := cmd.Secret
	if secret == "" {
		var err error
		secret, err = randomSecret()
		if err != nil {
			return domain.WebhookSubscription{}, err
		}
	}
	sub := domain.WebhookSubscription{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		URL:       cmd.URL,
		Events:    cmd.Events,
		SecretRef: secret,
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}
	return sub, s.repo.SaveWebhookSubscription(ctx, sub)
}

func (s *service) ListWebhooks(ctx context.Context, tenant kernel.TenantID) ([]domain.WebhookSubscription, error) {
	return s.repo.ListWebhookSubscriptions(ctx, tenant)
}

func (s *service) DeleteWebhook(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error {
	return s.repo.DeleteWebhookSubscription(ctx, tenant, id)
}

func (s *service) CreateKey(ctx context.Context, cmd ports.CreateApiKeyCommand) (ports.ApiKeyCreated, error) {
	plain, prefix, hash, err := generateApiKey()
	if err != nil {
		return ports.ApiKeyCreated{}, err
	}
	key := domain.NewApiKey(cmd.TenantID, cmd.Name, prefix, hash)
	if err := s.repo.SaveApiKey(ctx, key); err != nil {
		return ports.ApiKeyCreated{}, err
	}
	return ports.ApiKeyCreated{ApiKey: key, PlainKey: plain}, nil
}

func (s *service) RevokeKey(ctx context.Context, tenant kernel.TenantID, keyID uuid.UUID) error {
	key, err := s.repo.GetApiKey(ctx, tenant, keyID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	key.RevokedAt = &now
	return s.repo.SaveApiKey(ctx, key)
}

func (s *service) ListKeys(ctx context.Context, tenant kernel.TenantID) ([]domain.ApiKey, error) {
	return s.repo.ListApiKeys(ctx, tenant)
}

func generateApiKey() (plain, prefix, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	plain = "kore_" + hex.EncodeToString(buf)
	prefix = plain[:12]
	sum := sha256.Sum256([]byte(plain))
	hash = fmt.Sprintf("%x", sum)
	return plain, prefix, hash, nil
}

func randomSecret() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

var (
	_ ports.IntegrationService = (*service)(nil)
	_ ports.ApiKeyService      = (*service)(nil)
)
