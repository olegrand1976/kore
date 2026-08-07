package notifications_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	notifdomain "github.com/kore/kore/internal/modules/notifications/domain"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	wfnotif "github.com/kore/kore/internal/modules/workflow/adapters/notifications"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResolver struct {
	emails []string
}

func (s stubResolver) ResolveUserEmails(_ context.Context, _ kernel.TenantID, _ []uuid.UUID) ([]string, error) {
	return s.emails, nil
}
func (s stubResolver) ResolveEquipeUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return s.emails, nil
}
func (s stubResolver) ResolveApplicationUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return s.emails, nil
}
func (s stubResolver) ResolveServiceUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return s.emails, nil
}
func (s stubResolver) ResolveTenantUserEmails(_ context.Context, _ kernel.TenantID) ([]string, error) {
	return s.emails, nil
}
func (s stubResolver) ResolveEquipeUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s stubResolver) ResolveApplicationUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s stubResolver) ResolveServiceUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

type captureTransactional struct {
	msgs []notifports.TransactionalMessage
}

func (c *captureTransactional) NotifyTransactional(_ context.Context, msg notifports.TransactionalMessage) error {
	c.msgs = append(c.msgs, msg)
	return nil
}

func TestEffectsExecutorSendsEmail(t *testing.T) {
	tx := &captureTransactional{}
	exec := wfnotif.NewEffectsExecutor(stubResolver{emails: []string{"a@kore.local"}}, tx)
	tenant := kernel.NewTenantID(uuid.New())
	instanceID := uuid.New()
	actorID := uuid.New()

	err := exec.Execute(context.Background(), []domain.SideEffect{
		{
			Type:         domain.SideEffectTypeEmail,
			Recipients:   domain.EffectRecipients{Scope: domain.RecipientScopeAll},
			Subject:      "Entity {{entityId}}",
			BodyTemplate: "Action {{action}}",
		},
	}, ports.SideEffectContext{
		TenantID:       tenant,
		InstanceID:     instanceID,
		DefinitionCode: "leave.request",
		EntityID:       "entity-1",
		FromState:      "pending",
		ToState:        "approved",
		Action:         "approve",
		ActorID:        actorID,
	})
	require.NoError(t, err)
	require.Len(t, tx.msgs, 1)
	assert.Equal(t, "Entity entity-1", tx.msgs[0].Subject)
	assert.Contains(t, tx.msgs[0].Body, "Action approve")
	assert.Equal(t, []string{"a@kore.local"}, tx.msgs[0].Recipients)
}

func TestEffectsExecutorRequesterScope(t *testing.T) {
	tx := &captureTransactional{}
	resolver := &captureUserResolver{emails: []string{"requester@kore.local"}}
	exec := wfnotif.NewEffectsExecutor(resolver, tx)
	requesterID := uuid.New()

	err := exec.Execute(context.Background(), []domain.SideEffect{
		{
			Type:       domain.SideEffectTypeEmail,
			Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeRequester},
			Subject:    "Pour {{requesterId}}",
		},
	}, ports.SideEffectContext{
		TenantID:       kernel.NewTenantID(uuid.New()),
		InstanceID:     uuid.New(),
		DefinitionCode: "leave.request",
		EntityID:       "entity-1",
		ToState:        "approved",
		RequesterID:    requesterID,
	})
	require.NoError(t, err)
	require.Len(t, resolver.userIDs, 1)
	assert.Equal(t, []uuid.UUID{requesterID}, resolver.userIDs[0])
	require.Len(t, tx.msgs, 1)
	assert.Equal(t, "Pour "+requesterID.String(), tx.msgs[0].Subject)
}

func TestEffectsExecutorRequesterNilReturnsNoRecipients(t *testing.T) {
	tx := &captureTransactional{}
	exec := wfnotif.NewEffectsExecutor(&captureUserResolver{emails: []string{"x@kore.local"}}, tx)

	err := exec.Execute(context.Background(), []domain.SideEffect{
		{
			Type:       domain.SideEffectTypeEmail,
			Recipients: domain.EffectRecipients{Scope: domain.RecipientScopeRequester},
			Subject:    "Hi",
		},
	}, ports.SideEffectContext{
		TenantID:       kernel.NewTenantID(uuid.New()),
		InstanceID:     uuid.New(),
		DefinitionCode: "leave.request",
		EntityID:       "entity-1",
	})
	require.ErrorIs(t, err, notifdomain.ErrNoRecipients)
	assert.Empty(t, tx.msgs)
}

type captureUserResolver struct {
	emails  []string
	userIDs [][]uuid.UUID
}

func (s *captureUserResolver) ResolveUserEmails(_ context.Context, _ kernel.TenantID, ids []uuid.UUID) ([]string, error) {
	s.userIDs = append(s.userIDs, ids)
	return s.emails, nil
}
func (s *captureUserResolver) ResolveEquipeUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return s.emails, nil
}
func (s *captureUserResolver) ResolveApplicationUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return s.emails, nil
}
func (s *captureUserResolver) ResolveServiceUserEmails(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]string, error) {
	return s.emails, nil
}
func (s *captureUserResolver) ResolveTenantUserEmails(_ context.Context, _ kernel.TenantID) ([]string, error) {
	return s.emails, nil
}
func (s *captureUserResolver) ResolveEquipeUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *captureUserResolver) ResolveApplicationUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *captureUserResolver) ResolveServiceUserIDs(_ context.Context, _ kernel.TenantID, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
