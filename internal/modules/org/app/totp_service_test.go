package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/cryptox"
	"github.com/kore/kore/pkg/kernel"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/require"
)

type totpUserRepo struct {
	user    domain.User
	societe domain.Societe
}

func (r *totpUserRepo) SaveTenant(context.Context, domain.Tenant) error { return nil }
func (r *totpUserRepo) GetTenant(context.Context, kernel.TenantID) (domain.Tenant, error) {
	return domain.Tenant{}, nil
}
func (r *totpUserRepo) SaveSociete(context.Context, domain.Societe) error   { return nil }
func (r *totpUserRepo) UpdateSociete(context.Context, domain.Societe) error { return nil }
func (r *totpUserRepo) ListSocietes(_ context.Context, _ kernel.TenantID) ([]domain.Societe, error) {
	return []domain.Societe{r.societe}, nil
}
func (r *totpUserRepo) GetSociete(context.Context, kernel.TenantID, uuid.UUID) (domain.Societe, error) {
	return r.societe, nil
}
func (r *totpUserRepo) SaveSite(context.Context, domain.Site) error               { return nil }
func (r *totpUserRepo) SaveService(context.Context, domain.Service) error         { return nil }
func (r *totpUserRepo) SaveApplication(context.Context, domain.Application) error { return nil }
func (r *totpUserRepo) ListApplications(context.Context, kernel.TenantID) ([]domain.Application, error) {
	return nil, nil
}
func (r *totpUserRepo) GetApplication(context.Context, kernel.TenantID, uuid.UUID) (domain.Application, error) {
	return domain.Application{}, domain.ErrUserNotFound
}
func (r *totpUserRepo) SaveUser(context.Context, domain.User) error { return nil }
func (r *totpUserRepo) FindUserByID(context.Context, kernel.TenantID, uuid.UUID) (domain.User, error) {
	return r.user, nil
}
func (r *totpUserRepo) FindUserDetailByID(context.Context, kernel.TenantID, uuid.UUID) (ports.UserDetail, error) {
	return ports.UserDetail{}, domain.ErrUserNotFound
}
func (r *totpUserRepo) GetReleaseNotesPreferences(context.Context, kernel.TenantID, uuid.UUID) (ports.ReleaseNotesPreferences, error) {
	return ports.ReleaseNotesPreferences{}, nil
}
func (r *totpUserRepo) SetReleaseNotesAutoShow(context.Context, kernel.TenantID, uuid.UUID, bool) error {
	return nil
}
func (r *totpUserRepo) SetLastSeenVersion(context.Context, kernel.TenantID, uuid.UUID, string) error {
	return nil
}
func (r *totpUserRepo) UpdateUser(context.Context, domain.User) error { return nil }
func (r *totpUserRepo) SoftDeleteUser(context.Context, kernel.TenantID, uuid.UUID, time.Time) error {
	return nil
}
func (r *totpUserRepo) FindUserByLogin(context.Context, kernel.TenantID, string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}
func (r *totpUserRepo) FindUserByLoginGlobal(_ context.Context, login string) (domain.User, error) {
	if string(r.user.Login) == login {
		return r.user, nil
	}
	return domain.User{}, domain.ErrInvalidCredentials
}
func (r *totpUserRepo) ExistsLogin(context.Context, kernel.TenantID, string) (bool, error) {
	return false, nil
}
func (r *totpUserRepo) CountActiveUsers(context.Context, kernel.TenantID) (int, error) { return 0, nil }
func (r *totpUserRepo) ListUsers(context.Context, kernel.TenantID) ([]domain.User, error) {
	return nil, nil
}
func (r *totpUserRepo) SaveClient(context.Context, domain.Client) error { return nil }
func (r *totpUserRepo) GetClient(context.Context, kernel.TenantID, uuid.UUID) (domain.Client, error) {
	return domain.Client{}, domain.ErrUserNotFound
}
func (r *totpUserRepo) ListClients(context.Context, kernel.TenantID) ([]domain.Client, error) {
	return nil, nil
}
func (r *totpUserRepo) GetPermissions(context.Context) (map[string]map[authx.Module]map[authx.Action]bool, error) {
	return nil, nil
}
func (r *totpUserRepo) ResolveUserEmails(context.Context, kernel.TenantID, []uuid.UUID) ([]string, error) {
	return nil, nil
}
func (r *totpUserRepo) ResolveSocieteIDForUser(context.Context, kernel.TenantID, uuid.UUID) (uuid.UUID, error) {
	return r.societe.ID, nil
}
func (r *totpUserRepo) ResolveSocieteIDForEquipe(context.Context, kernel.TenantID, uuid.UUID) (uuid.UUID, error) {
	return r.societe.ID, nil
}
func (r *totpUserRepo) ListSocietesCraMailAuto(context.Context) ([]ports.CraMailReminderTarget, error) {
	return nil, nil
}
func (r *totpUserRepo) SaveIdentityProvider(context.Context, domain.IdentityProvider) error { return nil }
func (r *totpUserRepo) GetIdentityProvider(context.Context, kernel.TenantID) (domain.IdentityProvider, error) {
	return domain.IdentityProvider{}, domain.ErrSSONotEnabled
}
func (r *totpUserRepo) ListIdentityProviders(context.Context, kernel.TenantID) ([]domain.IdentityProvider, error) {
	return nil, nil
}
func (r *totpUserRepo) LinkUserIdentity(context.Context, domain.UserIdentityLink) error { return nil }
func (r *totpUserRepo) FindUserIdentityBySubject(context.Context, kernel.TenantID, uuid.UUID, string) (domain.UserIdentityLink, error) {
	return domain.UserIdentityLink{}, domain.ErrUserNotFound
}
func (r *totpUserRepo) FindUserByEmail(context.Context, kernel.TenantID, string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}
func (r *totpUserRepo) FindTenantIDsByEmail(context.Context, string) ([]kernel.TenantID, error) {
	return nil, nil
}
func (r *totpUserRepo) SaveAccessToken(context.Context, string, kernel.TenantID, string, string, time.Time) error {
	return nil
}
func (r *totpUserRepo) ConsumeAccessToken(context.Context, string, time.Time) (ports.AccessTokenRow, bool, error) {
	return ports.AccessTokenRow{}, false, nil
}
func (r *totpUserRepo) UpdateUserTotp(_ context.Context, u domain.User) error {
	r.user = u
	return nil
}
func (r *totpUserRepo) SaveTotpBackupCodes(context.Context, kernel.TenantID, uuid.UUID, []string) error {
	return nil
}
func (r *totpUserRepo) ConsumeTotpBackupCode(context.Context, kernel.TenantID, uuid.UUID, string, time.Time) (bool, error) {
	return false, nil
}
func (r *totpUserRepo) DeleteTotpBackupCodes(context.Context, kernel.TenantID, uuid.UUID) error {
	return nil
}
func (r *totpUserRepo) ListUnusedTotpBackupCodeHashes(context.Context, kernel.TenantID, uuid.UUID) ([]string, error) {
	return nil, nil
}
func (r *totpUserRepo) MarkTotpEnrollmentRequiredForSocieteUsers(context.Context, kernel.TenantID, uuid.UUID) (int, error) {
	return 0, nil
}
func (r *totpUserRepo) ClearTotpEnrollmentRequiredForSocieteUsers(context.Context, kernel.TenantID, uuid.UUID) error {
	return nil
}

func TestAuthenticate_requires2FAWhenEnabled(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	hasher := NewArgon2Hasher()
	hash, err := hasher.Hash("Admin123!")
	require.NoError(t, err)

	repo := &totpUserRepo{
		societe: domain.Societe{ID: uuid.New(), TenantID: tenant, TotpUserConfigurable: true},
		user: domain.User{
			ID: userID, TenantID: tenant, Login: "ADM_admin", PasswordHash: hash, Profile: domain.ProfileAdmin,
			Active: true, TotpEnabled: true,
			Period: domain.ActivationPeriod{Activation: time.Now().UTC().Add(-time.Hour)},
		},
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Kore", AccountName: "ADM_admin"})
	require.NoError(t, err)
	encrypted, err := cryptox.Encrypt(cryptox.DevKeyFromJWTSigningKey("test"), []byte(key.Secret()))
	require.NoError(t, err)
	repo.user.TotpSecretEncrypted = encrypted

	svc := NewUserService(repo, hasher, authx.NewTokenIssuer("test", time.Hour, time.Hour), nil, cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"), nil, cryptox.DevKeyFromJWTSigningKey("test"))
	result, err := svc.Authenticate(context.Background(), "ADM_admin", "Admin123!")
	require.NoError(t, err)
	require.True(t, result.Requires2FA)
	require.NotEmpty(t, result.ChallengeToken)
	require.Empty(t, result.AccessToken)
}

func TestAuthenticate_requiresEnrollmentWhenPolicyDefault(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	hasher := NewArgon2Hasher()
	hash, err := hasher.Hash("Admin123!")
	require.NoError(t, err)

	repo := &totpUserRepo{
		societe: domain.Societe{ID: uuid.New(), TenantID: tenant, TotpDefaultEnabled: true, TotpUserConfigurable: false},
		user: domain.User{
			ID: userID, TenantID: tenant, Login: "ADM_admin", PasswordHash: hash, Profile: domain.ProfileAdmin,
			Active: true,
			Period: domain.ActivationPeriod{Activation: time.Now().UTC().Add(-time.Hour)},
		},
	}
	svc := NewUserService(repo, hasher, authx.NewTokenIssuer("test", time.Hour, time.Hour), nil, cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"), nil, cryptox.DevKeyFromJWTSigningKey("test"))
	result, err := svc.Authenticate(context.Background(), "ADM_admin", "Admin123!")
	require.NoError(t, err)
	require.True(t, result.Requires2FAEnrollment)
	require.NotEmpty(t, result.EnrollmentToken)
}

func TestVerify2FAEnrollment_invalidCodeKeepsToken(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	hasher := NewArgon2Hasher()
	hash, err := hasher.Hash("Admin123!")
	require.NoError(t, err)

	repo := &totpUserRepo{
		societe: domain.Societe{ID: uuid.New(), TenantID: tenant, TotpDefaultEnabled: true, TotpUserConfigurable: false},
		user: domain.User{
			ID: userID, TenantID: tenant, Login: "ADM_admin", PasswordHash: hash, Profile: domain.ProfileAdmin,
			Active: true,
			Period: domain.ActivationPeriod{Activation: time.Now().UTC().Add(-time.Hour)},
		},
	}
	svc := NewUserService(repo, hasher, authx.NewTokenIssuer("test", time.Hour, time.Hour), nil, cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"), nil, cryptox.DevKeyFromJWTSigningKey("test"))
	auth, err := svc.Authenticate(context.Background(), "ADM_admin", "Admin123!")
	require.NoError(t, err)
	require.True(t, auth.Requires2FAEnrollment)

	_, err = svc.Setup2FAWithEnrollmentToken(context.Background(), auth.EnrollmentToken)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		_, err = svc.Verify2FAEnrollment(context.Background(), auth.EnrollmentToken, "000000", "")
		require.ErrorIs(t, err, domain.Err2FAInvalidCode)
	}
}

func TestVerify2FAEnrollment_rateLimited(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	hasher := NewArgon2Hasher()
	hash, err := hasher.Hash("Admin123!")
	require.NoError(t, err)

	repo := &totpUserRepo{
		societe: domain.Societe{ID: uuid.New(), TenantID: tenant, TotpDefaultEnabled: true, TotpUserConfigurable: false},
		user: domain.User{
			ID: userID, TenantID: tenant, Login: "ADM_admin", PasswordHash: hash, Profile: domain.ProfileAdmin,
			Active: true,
			Period: domain.ActivationPeriod{Activation: time.Now().UTC().Add(-time.Hour)},
		},
	}
	svc := NewUserService(repo, hasher, authx.NewTokenIssuer("test", time.Hour, time.Hour), nil, cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"), nil, cryptox.DevKeyFromJWTSigningKey("test"))
	auth, err := svc.Authenticate(context.Background(), "ADM_admin", "Admin123!")
	require.NoError(t, err)

	_, err = svc.Setup2FAWithEnrollmentToken(context.Background(), auth.EnrollmentToken)
	require.NoError(t, err)

	for i := 0; i < totpRateLimitMax; i++ {
		_, err = svc.Verify2FAEnrollment(context.Background(), auth.EnrollmentToken, "000000", "")
		require.ErrorIs(t, err, domain.Err2FAInvalidCode)
	}
	_, err = svc.Verify2FAEnrollment(context.Background(), auth.EnrollmentToken, "000000", "")
	require.ErrorIs(t, err, domain.Err2FARateLimited)
}

func TestVerify2FAChallenge_rateLimited(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	hasher := NewArgon2Hasher()
	hash, err := hasher.Hash("Admin123!")
	require.NoError(t, err)

	repo := &totpUserRepo{
		societe: domain.Societe{ID: uuid.New(), TenantID: tenant},
		user: domain.User{
			ID: userID, TenantID: tenant, Login: "ADM_admin", PasswordHash: hash, Profile: domain.ProfileAdmin,
			Active: true, TotpEnabled: true,
			Period: domain.ActivationPeriod{Activation: time.Now().UTC().Add(-time.Hour)},
		},
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Kore", AccountName: "ADM_admin"})
	require.NoError(t, err)
	encrypted, err := cryptox.Encrypt(cryptox.DevKeyFromJWTSigningKey("test"), []byte(key.Secret()))
	require.NoError(t, err)
	repo.user.TotpSecretEncrypted = encrypted

	svc := NewUserService(repo, hasher, authx.NewTokenIssuer("test", time.Hour, time.Hour), nil, cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"), nil, cryptox.DevKeyFromJWTSigningKey("test"))
	auth, err := svc.Authenticate(context.Background(), "ADM_admin", "Admin123!")
	require.NoError(t, err)

	for i := 0; i < totpRateLimitMax; i++ {
		_, err := svc.Verify2FAChallenge(context.Background(), auth.ChallengeToken, "000000")
		require.ErrorIs(t, err, domain.Err2FAInvalidCode)
	}
	_, err = svc.Verify2FAChallenge(context.Background(), auth.ChallengeToken, "000000")
	require.ErrorIs(t, err, domain.Err2FARateLimited)
}

func TestBeginTotpSetup_includesQRDataURL(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	repo := &totpUserRepo{
		societe: domain.Societe{ID: uuid.New(), TenantID: tenant, TotpUserConfigurable: true},
		user: domain.User{
			ID: userID, TenantID: tenant, Login: "ADM_admin", PasswordHash: "hash", Profile: domain.ProfileAdmin,
			Active: true,
		},
	}
	svc := NewUserService(repo, NewArgon2Hasher(), authx.NewTokenIssuer("test", time.Hour, time.Hour), nil, cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"), nil, cryptox.DevKeyFromJWTSigningKey("test"))
	result, err := svc.Setup2FA(context.Background(), tenant, userID)
	require.NoError(t, err)
	require.NotEmpty(t, result.QrCodeDataURL)
	require.True(t, strings.HasPrefix(result.QrCodeDataURL, "data:image/png;base64,"))
}
