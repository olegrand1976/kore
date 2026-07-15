package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/pkg/kernel"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

type fakeRepo struct {
	mu       sync.Mutex
	rules    map[string]domain.NotificationRule
	messages map[uuid.UUID]domain.NotificationMessage
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		rules:    map[string]domain.NotificationRule{},
		messages: map[uuid.UUID]domain.NotificationMessage{},
	}
}

func (r *fakeRepo) SaveRule(_ context.Context, rule domain.NotificationRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[rule.Trigger] = rule
	return nil
}

func (r *fakeRepo) GetRuleByTrigger(_ context.Context, _ kernel.TenantID, trigger string) (domain.NotificationRule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rule, ok := r.rules[trigger]
	if !ok {
		return domain.NotificationRule{}, domain.ErrRuleNotFound
	}
	return rule, nil
}

func (r *fakeRepo) GetRuleByCode(_ context.Context, _ kernel.TenantID, code string) (domain.NotificationRule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rule := range r.rules {
		if rule.Code == code {
			return rule, nil
		}
	}
	return domain.NotificationRule{}, domain.ErrRuleNotFound
}

func (r *fakeRepo) ListRules(_ context.Context, _ kernel.TenantID) ([]domain.NotificationRule, error) {
	return nil, nil
}

func (r *fakeRepo) SaveMessage(_ context.Context, m domain.NotificationMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages[m.ID] = m
	return nil
}

func (r *fakeRepo) ListMessages(_ context.Context, _ ports.SentFilter) ([]domain.NotificationMessage, error) {
	return nil, nil
}

func (r *fakeRepo) ListPending(_ context.Context, _ kernel.TenantID) ([]domain.NotificationMessage, error) {
	return nil, nil
}

func (r *fakeRepo) ListDue(_ context.Context, now time.Time, limit int) ([]domain.NotificationMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.NotificationMessage
	for _, m := range r.messages {
		if m.Status != domain.MessageStatusPending {
			continue
		}
		if m.ScheduledFor != nil && m.ScheduledFor.After(now) {
			continue
		}
		out = append(out, m)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

type fakeSender struct {
	mu   sync.Mutex
	sent int
}

func (s *fakeSender) Send(_ context.Context, _ ports.Email) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent++
	return nil
}

type fakeResolver struct{}

func (fakeResolver) ResolveUserEmails(_ context.Context, _ kernel.TenantID, _ []uuid.UUID) ([]string, error) {
	return []string{"user@example.com"}, nil
}

func (fakeResolver) ResolveEquipeUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return []string{"user@example.com"}, nil
}

func (fakeResolver) ResolveApplicationUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return []string{"user@example.com"}, nil
}

func (fakeResolver) ResolveServiceUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return []string{"user@example.com"}, nil
}

func (fakeResolver) ResolveEquipeUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.New()}, nil
}

func (fakeResolver) ResolveApplicationUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.New()}, nil
}

func (fakeResolver) ResolveServiceUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.New()}, nil
}

func TestScheduledMessageNotSentBeforeDue(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	sender := &fakeSender{}
	// Tuesday 2026-01-06 10:00 UTC.
	clock := &fakeClock{now: time.Date(2026, time.January, 6, 10, 0, 0, 0, time.UTC)}
	svc := NewService(repo, sender, fakeResolver{}, WithClock(clock))

	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	if err := svc.DefineRule(ctx, domain.NotificationRule{
		TenantID:         tenant,
		Code:             "weekly-recap",
		Trigger:          "cra.reminder",
		Frequency:        domain.FrequencyMonday,
		RecipientsPolicy: domain.RecipientPolicy{UserIDs: []uuid.UUID{userID}},
		Template:         "Rappel CRA",
	}); err != nil {
		t.Fatalf("DefineRule: %v", err)
	}

	if err := svc.Publish(ctx, ports.NotificationEvent{TenantID: tenant, Trigger: "cra.reminder"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// Still Tuesday: the monday message must not be dispatched.
	n, err := svc.ProcessPending(ctx)
	if err != nil {
		t.Fatalf("ProcessPending: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 dispatched on tuesday, got %d", n)
	}
	if sender.sent != 0 {
		t.Fatalf("expected 0 emails sent on tuesday, got %d", sender.sent)
	}

	// Advance to next Monday 09:00: the message is now due.
	clock.now = time.Date(2026, time.January, 12, 9, 0, 0, 0, time.UTC)
	n, err = svc.ProcessPending(ctx)
	if err != nil {
		t.Fatalf("ProcessPending after due: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 dispatched on monday, got %d", n)
	}
	if sender.sent != 1 {
		t.Fatalf("expected 1 email sent on monday, got %d", sender.sent)
	}
}
