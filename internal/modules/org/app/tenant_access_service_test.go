package app

import (
	"context"
	"errors"
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

func (r *fakeTenantAccessRepo) FindAccessToken(ctx context.Context, tokenHash string) (ports.AccessTokenRow, bool, error) {
	_ = ctx
	row, ok := r.tokens[tokenHash]
	if !ok {
		return ports.AccessTokenRow{}, false, nil
	}
	return row, true, nil
}

func (r *fakeTenantAccessRepo) InvalidateUnusedAccessTokens(ctx context.Context, tenant kernel.TenantID, email, kind string, now time.Time) error {
	_ = ctx
	for hash, row := range r.tokens {
		if row.TenantID == tenant && row.Email == email && row.Kind == kind && row.UsedAt == nil {
			used := now
			row.UsedAt = &used
			r.tokens[hash] = row
		}
	}
	return nil
}

func (r *fakeTenantAccessRepo) ConsumeAccessToken(ctx context.Context, tokenHash string, now time.Time) (ports.AccessTokenRow, bool, error) {
	return r.ConsumeAccessTokenOfKind(ctx, tokenHash, "", now)
}

func (r *fakeTenantAccessRepo) ConsumeAccessTokenOfKind(ctx context.Context, tokenHash, kind string, now time.Time) (ports.AccessTokenRow, bool, error) {
	_ = ctx
	row, ok := r.tokens[tokenHash]
	if !ok {
		return ports.AccessTokenRow{}, false, nil
	}
	if row.UsedAt != nil || !row.ExpiresAt.After(now) {
		return row, false, nil
	}
	if kind != "" && row.Kind != kind {
		return row, false, nil
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
	require.ErrorIs(t, err, domain.ErrAccessTokenUsed)

	// Control: direct consume should also reject.
	row, ok, err := repo.ConsumeAccessToken(context.Background(), tokenHash, now)
	require.NoError(t, err)
	require.False(t, ok)
	require.NotEmpty(t, row.TokenHash)

	require.Len(t, mailer.sentTo, 0)
}

func TestTenantAccessService_RequestPasswordReset_SendsPerTenant(t *testing.T) {
	tenant1 := kernel.NewTenantID(uuid.MustParse("00000000-0000-4000-8000-000000000001"))
	tenant2 := kernel.NewTenantID(uuid.MustParse("00000000-0000-4000-8000-0000000000a1"))
	repo := &fakeTenantAccessRepo{
		tenantsByEmail: map[string][]kernel.TenantID{
			"user@example.com": {tenant1, tenant2},
		},
		tokens: map[string]ports.AccessTokenRow{},
	}
	mailer := &fakeMailer{}
	svc := NewTenantAccessService(repo, mailer)
	now := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	svc.clock = func() time.Time { return now }

	require.NoError(t, svc.RequestPasswordReset(context.Background(), "user@example.com", "https://kore.example/reset-password"))
	require.Len(t, mailer.sentTo, 2)
	require.Len(t, repo.tokens, 2)
	for _, row := range repo.tokens {
		require.Equal(t, string(AccessTokenKindPasswordReset), row.Kind)
		require.Equal(t, now.Add(time.Hour), row.ExpiresAt)
	}
}

func TestTenantAccessService_RequestPasswordReset_UnknownEmailSilent(t *testing.T) {
	repo := &fakeTenantAccessRepo{
		tenantsByEmail: map[string][]kernel.TenantID{},
		tokens:         map[string]ports.AccessTokenRow{},
	}
	mailer := &fakeMailer{}
	svc := NewTenantAccessService(repo, mailer)
	require.NoError(t, svc.RequestPasswordReset(context.Background(), "unknown@example.com", "https://kore.example/reset-password"))
	require.Empty(t, mailer.sentTo)
	require.Empty(t, repo.tokens)
}

func TestTenantAccessService_ConfirmPasswordReset(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.MustParse("00000000-0000-4000-8000-0000000000a1"))
	repo := &fakeTenantAccessRepo{
		tenantsByEmail: map[string][]kernel.TenantID{},
		tokens:         map[string]ports.AccessTokenRow{},
	}
	mailer := &fakeMailer{}
	users := &fakePasswordResetUsers{
		byEmail: map[string]domain.User{
			"user@example.com": {
				ID:           uuid.MustParse("11111111-1111-4111-8111-111111111111"),
				TenantID:     tenant,
				Login:        "ADM_olivier",
				Email:        "user@example.com",
				PasswordHash: "old",
				Active:       true,
				Period:       domain.ActivationPeriod{Activation: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
	}
	hasher := NewArgon2Hasher()
	svc := NewTenantAccessService(repo, mailer).WithPasswordReset(users, hasher)
	now := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	svc.clock = func() time.Time { return now }

	token, tokenHash, err := newToken()
	require.NoError(t, err)
	require.NoError(t, repo.SaveAccessToken(context.Background(), tokenHash, tenant, "user@example.com", string(AccessTokenKindPasswordReset), now.Add(time.Hour)))

	require.NoError(t, svc.ConfirmPasswordReset(context.Background(), token, "Smicer22$1"))
	updated := users.byEmail["user@example.com"]
	require.NotEqual(t, "old", updated.PasswordHash)
	require.True(t, hasher.Verify(updated.PasswordHash, "Smicer22$1"))

	require.ErrorIs(t, svc.ConfirmPasswordReset(context.Background(), token, "Smicer22$1"), domain.ErrAccessTokenUsed)
}

func TestTenantAccessService_ConfirmPasswordReset_RejectsWrongKind(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.MustParse("00000000-0000-4000-8000-000000000001"))
	repo := &fakeTenantAccessRepo{tokens: map[string]ports.AccessTokenRow{}}
	svc := NewTenantAccessService(repo, &fakeMailer{}).WithPasswordReset(&fakePasswordResetUsers{}, NewArgon2Hasher())
	now := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	svc.clock = func() time.Time { return now }
	token, tokenHash, err := newToken()
	require.NoError(t, err)
	require.NoError(t, repo.SaveAccessToken(context.Background(), tokenHash, tenant, "user@example.com", string(AccessTokenKindDiscovery), now.Add(time.Hour)))
	require.ErrorIs(t, svc.ConfirmPasswordReset(context.Background(), token, "Smicer22$1"), domain.ErrAccessTokenInvalid)

	// Discovery token must remain usable after a wrong-kind confirm attempt.
	res, err := svc.Resolve(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, string(AccessTokenKindDiscovery), res.Kind)
}

func TestTenantAccessService_ConfirmPasswordReset_KeepsTokenOnUpdateFailure(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.MustParse("00000000-0000-4000-8000-0000000000a1"))
	repo := &fakeTenantAccessRepo{tokens: map[string]ports.AccessTokenRow{}}
	users := &fakePasswordResetUsers{failUpdate: true}
	svc := NewTenantAccessService(repo, &fakeMailer{}).WithPasswordReset(users, NewArgon2Hasher())
	now := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	svc.clock = func() time.Time { return now }
	users.byEmail = map[string]domain.User{
		"user@example.com": {
			ID:       uuid.MustParse("11111111-1111-4111-8111-111111111111"),
			TenantID: tenant,
			Email:    "user@example.com",
			Active:   true,
			Period:   domain.ActivationPeriod{Activation: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	token, tokenHash, err := newToken()
	require.NoError(t, err)
	require.NoError(t, repo.SaveAccessToken(context.Background(), tokenHash, tenant, "user@example.com", string(AccessTokenKindPasswordReset), now.Add(time.Hour)))
	require.Error(t, svc.ConfirmPasswordReset(context.Background(), token, "Smicer22$1"))
	row, ok := repo.tokens[tokenHash]
	require.True(t, ok)
	require.Nil(t, row.UsedAt)
}

func TestTenantAccessService_RequestPasswordReset_InvalidatesPrevious(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.MustParse("00000000-0000-4000-8000-000000000001"))
	repo := &fakeTenantAccessRepo{
		tenantsByEmail: map[string][]kernel.TenantID{"user@example.com": {tenant}},
		tokens:         map[string]ports.AccessTokenRow{},
	}
	svc := NewTenantAccessService(repo, &fakeMailer{})
	now := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	svc.clock = func() time.Time { return now }
	oldHash := "old-hash"
	require.NoError(t, repo.SaveAccessToken(context.Background(), oldHash, tenant, "user@example.com", string(AccessTokenKindPasswordReset), now.Add(time.Hour)))
	require.NoError(t, svc.RequestPasswordReset(context.Background(), "user@example.com", "https://kore.example/reset-password"))
	require.NotNil(t, repo.tokens[oldHash].UsedAt)
	require.Len(t, repo.tokens, 2)
}

func TestTenantAccessService_ConfirmPasswordReset_WeakPassword(t *testing.T) {
	svc := NewTenantAccessService(&fakeTenantAccessRepo{}, &fakeMailer{}).WithPasswordReset(&fakePasswordResetUsers{}, NewArgon2Hasher())
	require.ErrorIs(t, svc.ConfirmPasswordReset(context.Background(), "tok", "short"), domain.ErrWeakPassword)
}

type fakePasswordResetUsers struct {
	byEmail    map[string]domain.User
	failUpdate bool
}

func (f *fakePasswordResetUsers) FindUserByEmail(_ context.Context, tenant kernel.TenantID, email string) (domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok || u.TenantID != tenant {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}

func (f *fakePasswordResetUsers) UpdateUser(_ context.Context, u domain.User) error {
	if f.failUpdate {
		return errors.New("update failed")
	}
	if f.byEmail == nil {
		f.byEmail = map[string]domain.User{}
	}
	f.byEmail[u.Email] = u
	return nil
}
