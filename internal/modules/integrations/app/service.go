package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo ports.IntegrationRepository
}

func NewService(repo ports.IntegrationRepository) ports.IntegrationService {
	return &service{repo: repo}
}

func NewApiKeyService(repo ports.IntegrationRepository) ports.ApiKeyService {
	return &service{repo: repo}
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
	conn.LastSyncAt = &now
	if err := s.repo.SaveConnection(ctx, conn); err != nil {
		return domain.SyncJob{}, err
	}
	return job, s.repo.SaveSyncJob(ctx, job)
}

func (s *service) ListConnections(ctx context.Context, tenant kernel.TenantID) ([]domain.IntegrationConnection, error) {
	return s.repo.ListConnections(ctx, tenant)
}

func (s *service) GetConnection(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.IntegrationConnection, error) {
	return s.repo.GetConnection(ctx, tenant, id)
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

var (
	_ ports.IntegrationService = (*service)(nil)
	_ ports.ApiKeyService      = (*service)(nil)
)
