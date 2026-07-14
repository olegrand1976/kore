package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

type syncRepo struct {
	jobs []domain.SyncJob
	conn domain.IntegrationConnection
}

func (r *syncRepo) SaveConnection(context.Context, domain.IntegrationConnection) error { return nil }
func (r *syncRepo) GetConnection(_ context.Context, _ kernel.TenantID, _ uuid.UUID) (domain.IntegrationConnection, error) {
	if r.conn.Provider != "" {
		return r.conn, nil
	}
	return domain.IntegrationConnection{
		ID:       uuid.New(),
		Status:   domain.ConnectionStatusActive,
		Provider: "fec",
	}, nil
}
func (r *syncRepo) ListConnections(context.Context, kernel.TenantID) ([]domain.IntegrationConnection, error) {
	return nil, nil
}
func (r *syncRepo) DeleteConnection(context.Context, kernel.TenantID, uuid.UUID) error { return nil }
func (r *syncRepo) SaveSyncJob(_ context.Context, j domain.SyncJob) error {
	r.jobs = append(r.jobs, j)
	return nil
}
func (r *syncRepo) ListSyncJobs(context.Context, kernel.TenantID) ([]domain.SyncJob, error) {
	return r.jobs, nil
}
func (r *syncRepo) SaveApiKey(context.Context, domain.ApiKey) error { return nil }
func (r *syncRepo) GetApiKey(context.Context, kernel.TenantID, uuid.UUID) (domain.ApiKey, error) {
	return domain.ApiKey{}, domain.ErrApiKeyNotFound
}
func (r *syncRepo) GetApiKeyByHash(context.Context, string) (domain.ApiKey, error) {
	return domain.ApiKey{}, domain.ErrApiKeyNotFound
}
func (r *syncRepo) ListApiKeys(context.Context, kernel.TenantID) ([]domain.ApiKey, error) {
	return nil, nil
}
func (r *syncRepo) SaveWebhookSubscription(context.Context, domain.WebhookSubscription) error {
	return nil
}
func (r *syncRepo) ListWebhookSubscriptions(context.Context, kernel.TenantID) ([]domain.WebhookSubscription, error) {
	return nil, nil
}
func (r *syncRepo) DeleteWebhookSubscription(context.Context, kernel.TenantID, uuid.UUID) error {
	return nil
}

func TestSyncFECProvider(t *testing.T) {
	repo := &syncRepo{}
	svc := NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	job, err := svc.Sync(context.Background(), ports.SyncCommand{
		TenantID:     tenant,
		ConnectionID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
	if job.ErrorMessage == "" {
		t.Fatal("expected fec export message")
	}
}

func TestSyncLuccaProvider(t *testing.T) {
	repo := &syncRepo{
		conn: domain.IntegrationConnection{
			ID:       uuid.New(),
			Status:   domain.ConnectionStatusActive,
			Type:     domain.ConnectionTypeHRIS,
			Provider: "lucca",
		},
	}
	svc := NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	job, err := svc.Sync(context.Background(), ports.SyncCommand{
		TenantID:     tenant,
		ConnectionID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
	if job.ErrorMessage == "" {
		t.Fatal("expected hris sync message")
	}
}

func TestSyncGoogleCalendarProvider(t *testing.T) {
	repo := &syncRepo{
		conn: domain.IntegrationConnection{
			ID:       uuid.New(),
			Status:   domain.ConnectionStatusActive,
			Type:     domain.ConnectionTypeCalendar,
			Provider: "google",
		},
	}
	svc := NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	job, err := svc.Sync(context.Background(), ports.SyncCommand{
		TenantID:     tenant,
		ConnectionID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
	if job.FinishedAt == nil || job.FinishedAt.Before(time.Now().Add(-time.Minute)) {
		t.Fatal("expected finished timestamp")
	}
}
