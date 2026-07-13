package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/pkg/kernel"
)

type captureResolver struct {
	lastCalled string
}

func (r *captureResolver) ResolveUserEmails(_ context.Context, _ kernel.TenantID, _ []uuid.UUID) ([]string, error) {
	r.lastCalled = "userIds"
	return []string{"a@kore.local"}, nil
}

func (r *captureResolver) ResolveEquipeUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	r.lastCalled = "equipeId"
	return []string{"a@kore.local"}, nil
}

func (r *captureResolver) ResolveApplicationUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	r.lastCalled = "applicationId"
	return []string{"a@kore.local"}, nil
}

func (r *captureResolver) ResolveServiceUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	r.lastCalled = "serviceId"
	return []string{"a@kore.local"}, nil
}

func TestRecipientPolicyResolutionPrecedence(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	sender := &fakeSender{}
	resolver := &captureResolver{}
	svc := NewService(repo, sender, resolver)

	tenant := kernel.NewTenantID(uuid.New())
	anyID := uuid.New()
	userID := uuid.New()
	trigger := "test.trigger"

	tests := []struct {
		name           string
		policy         domain.RecipientPolicy
		expectedCalled string
	}{
		{
			name:           "userIds_has_precedence",
			policy:         domain.RecipientPolicy{UserIDs: []uuid.UUID{userID}, EquipeID: &anyID, ApplicationID: &anyID, ServiceID: &anyID},
			expectedCalled: "userIds",
		},
		{
			name:           "equipeId_when_set",
			policy:         domain.RecipientPolicy{EquipeID: &anyID, ApplicationID: &anyID, ServiceID: &anyID},
			expectedCalled: "equipeId",
		},
		{
			name:           "applicationId_when_set",
			policy:         domain.RecipientPolicy{ApplicationID: &anyID, ServiceID: &anyID},
			expectedCalled: "applicationId",
		},
		{
			name:           "serviceId_when_set",
			policy:         domain.RecipientPolicy{ServiceID: &anyID},
			expectedCalled: "serviceId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver.lastCalled = ""
			if err := svc.DefineRule(ctx, domain.NotificationRule{
				TenantID:         tenant,
				Code:             "code-" + tt.name,
				Trigger:          trigger,
				Frequency:        domain.FrequencyImmediate,
				RecipientsPolicy: tt.policy,
				Template:         "hello",
			}); err != nil {
				t.Fatalf("DefineRule: %v", err)
			}
			if err := svc.Publish(ctx, ports.NotificationEvent{TenantID: tenant, Trigger: trigger}); err != nil {
				t.Fatalf("Publish: %v", err)
			}
			if resolver.lastCalled != tt.expectedCalled {
				t.Fatalf("expected resolver %q, got %q", tt.expectedCalled, resolver.lastCalled)
			}
		})
	}
}

