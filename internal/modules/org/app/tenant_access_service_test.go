package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type fakeTenantAccessRepo struct {
	tenantsByEmail map[string][]kernel.TenantID
	tokens         map[string]ports.AccessTokenRow // key = tokenHash
}

func (r *fakeTenantAccessRepo) FindTenantIDsByEmail(ctx context.Context, email string) ([]kernel.TenantID, error) {
	_ = ctx
	return r.tenantsByEmail[email], nil
}

func (r *fakeTenantAccessRepo) SaveAccessToken(ctx context.Context, tokenHash string, tenant kernel.TenantID, email, kind string, expiresAt time.Time) error {
	_ = ctx
	if r.tokens == nil {
		r.tokens = make(map[string]ports.AccessTokenRow)
	}
	r.tokens[tokenHash] = ports.AccessTokenRow{
		TokenHash: tokenHash,
		TenantID:  tenant,
		Email:     email,
		Kind:      kind,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	return nil
}

func (r *fakeTenantAccessRepo) ConsumeAccessToken(ctx context.Context, tokenHash string, now time.Time) (ports.AccessTokenRow, bool, error) {
	_ = ctx
	row, ok := r.tokens[tokenHash]
	if !ok {
		return ports.AccessTokenRow{}, false, nil
	}
	if row.UsedAt != nil || !row.ExpiresAt.After(now) {
		return ports.AccessTokenRow{}, false, nil
	}
	used := now
	row.UsedAt = &used
	r.tokens[tokenHash] = row
	return row, true, nil
}

type fakeMailer struct {
	sentTo []string
}

func (m *fakeMailer) SendTenantAccessEmail(ctx context.Context, to string, subject string, body string) error {
	_ = ctx
	_ = subject
	_ = body
	m.sentTo = append(m.sentTo, to)
	return nil
}

func TestTenantAccessService_Resolve_IsSingleUse(t *testing.T) {
	repo := &fakeTenantAccessRepo{
		tenantsByEmail: map[string][]kernel.TenantID{},
		tokens:         map[string]ports.AccessTokenRow{},
	}
	mailer := &fakeMailer{}
	svc := NewTenantAccessService(repo, mailer)
	now := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)
	svc.clock = func() time.Time { return now }

	tenant := kernel.NewTenantID(uuid.MustParse("00000000-0000-4000-8000-000000000001"))
	token, tokenHash, err := newToken()
	require.NoError(t, err)
	expiresAt := now.Add(24 * time.Hour)
	require.NoError(t, repo.SaveAccessToken(context.Background(), tokenHash, tenant, "user@example.com", string(AccessTokenKindInvite), expiresAt))

	// Resolve with wrong token should fail
	_, err = svc.Resolve(context.Background(), "wrong-token")
	require.ErrorIs(t, err, domain.ErrAccessTokenInvalid)

	res, err := svc.Resolve(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, tenant, res.TenantID)
	require.Equal(t, string(AccessTokenKindInvite), res.Kind)

	// Second resolve should be invalid (single use).
	_, err = svc.Resolve(context.Background(), token)
	require.ErrorIs(t, err, domain.ErrAccessTokenInvalid)

	// Control: direct consume should also reject.
	row, ok, err := repo.ConsumeAccessToken(context.Background(), tokenHash, now)
	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, ports.AccessTokenRow{}, row)

	require.Len(t, mailer.sentTo, 0)
}

