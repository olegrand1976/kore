package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type oidcRepoStub struct {
	idp   domain.IdentityProvider
	users map[uuid.UUID]domain.User
	links []domain.UserIdentityLink
}

func (s *oidcRepoStub) SaveTenant(context.Context, domain.Tenant) error { return nil }
func (s *oidcRepoStub) GetTenant(context.Context, kernel.TenantID) (domain.Tenant, error) {
	return domain.Tenant{}, nil
}
func (s *oidcRepoStub) SaveSociete(context.Context, domain.Societe) error   { return nil }
func (s *oidcRepoStub) UpdateSociete(context.Context, domain.Societe) error { return nil }
func (s *oidcRepoStub) ListSocietes(context.Context, kernel.TenantID) ([]domain.Societe, error) {
	return nil, nil
}
func (s *oidcRepoStub) GetSociete(context.Context, kernel.TenantID, uuid.UUID) (domain.Societe, error) {
	return domain.Societe{}, nil
}
func (s *oidcRepoStub) SaveSite(context.Context, domain.Site) error               { return nil }
func (s *oidcRepoStub) SaveService(context.Context, domain.Service) error         { return nil }
func (s *oidcRepoStub) SaveApplication(context.Context, domain.Application) error { return nil }
func (s *oidcRepoStub) ListApplications(context.Context, kernel.TenantID) ([]domain.Application, error) {
	return nil, nil
}
func (s *oidcRepoStub) GetApplication(context.Context, kernel.TenantID, uuid.UUID) (domain.Application, error) {
	return domain.Application{}, domain.ErrUserNotFound
}
func (s *oidcRepoStub) SaveUser(_ context.Context, u domain.User) error {
	if s.users == nil {
		s.users = map[uuid.UUID]domain.User{}
	}
	s.users[u.ID] = u
	return nil
}
func (s *oidcRepoStub) FindUserByID(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.User, error) {
	u, ok := s.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
}
func (s *oidcRepoStub) FindUserDetailByID(context.Context, kernel.TenantID, uuid.UUID) (ports.UserDetail, error) {
	return ports.UserDetail{}, domain.ErrUserNotFound
}
func (s *oidcRepoStub) GetReleaseNotesPreferences(context.Context, kernel.TenantID, uuid.UUID) (ports.ReleaseNotesPreferences, error) {
	return ports.ReleaseNotesPreferences{LastSeenVersion: nil, AutoShowEnabled: true}, nil
}
func (s *oidcRepoStub) SetReleaseNotesAutoShow(context.Context, kernel.TenantID, uuid.UUID, bool) error {
	return nil
}
func (s *oidcRepoStub) SetLastSeenVersion(context.Context, kernel.TenantID, uuid.UUID, string) error {
	return nil
}
func (s *oidcRepoStub) UpdateUser(context.Context, domain.User) error { return nil }
func (s *oidcRepoStub) SoftDeleteUser(context.Context, kernel.TenantID, uuid.UUID, time.Time) error {
	return nil
}
func (s *oidcRepoStub) FindUserByLogin(context.Context, kernel.TenantID, string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}
func (s *oidcRepoStub) FindUserByLoginGlobal(context.Context, string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}
func (s *oidcRepoStub) ExistsLogin(context.Context, kernel.TenantID, string) (bool, error) {
	return false, nil
}
func (s *oidcRepoStub) CountActiveUsers(context.Context, kernel.TenantID) (int, error) {
	return len(s.users), nil
}
func (s *oidcRepoStub) ListUsers(context.Context, kernel.TenantID) ([]domain.User, error) {
	return nil, nil
}
func (s *oidcRepoStub) SaveClient(context.Context, domain.Client) error { return nil }
func (s *oidcRepoStub) ListClients(context.Context, kernel.TenantID) ([]domain.Client, error) {
	return nil, nil
}
func (s *oidcRepoStub) GetClient(context.Context, kernel.TenantID, uuid.UUID) (domain.Client, error) {
	return domain.Client{}, domain.ErrUserNotFound
}
func (s *oidcRepoStub) GetPermissions(context.Context) (map[string]map[authx.Module]map[authx.Action]bool, error) {
	return nil, nil
}
func (s *oidcRepoStub) ResolveUserEmails(context.Context, kernel.TenantID, []uuid.UUID) ([]string, error) {
	return nil, nil
}
func (s *oidcRepoStub) ResolveSocieteIDForUser(context.Context, kernel.TenantID, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (s *oidcRepoStub) ListSocietesCraMailAuto(context.Context) ([]ports.CraMailReminderTarget, error) {
	return nil, nil
}
func (s *oidcRepoStub) SaveIdentityProvider(_ context.Context, idp domain.IdentityProvider) error {
	s.idp = idp
	return nil
}
func (s *oidcRepoStub) GetIdentityProvider(context.Context, kernel.TenantID) (domain.IdentityProvider, error) {
	if !s.idp.Enabled {
		return domain.IdentityProvider{}, domain.ErrSSONotEnabled
	}
	return s.idp, nil
}
func (s *oidcRepoStub) ListIdentityProviders(context.Context, kernel.TenantID) ([]domain.IdentityProvider, error) {
	return []domain.IdentityProvider{s.idp}, nil
}
func (s *oidcRepoStub) LinkUserIdentity(_ context.Context, link domain.UserIdentityLink) error {
	s.links = append(s.links, link)
	return nil
}
func (s *oidcRepoStub) FindUserIdentityBySubject(context.Context, kernel.TenantID, uuid.UUID, string) (domain.UserIdentityLink, error) {
	return domain.UserIdentityLink{}, domain.ErrUserNotFound
}
func (s *oidcRepoStub) FindUserByEmail(_ context.Context, _ kernel.TenantID, email string) (domain.User, error) {
	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (s *oidcRepoStub) FindTenantIDsByEmail(context.Context, string) ([]kernel.TenantID, error) {
	return nil, nil
}
func (s *oidcRepoStub) SaveAccessToken(context.Context, string, kernel.TenantID, string, string, time.Time) error {
	return nil
}
func (s *oidcRepoStub) ConsumeAccessToken(context.Context, string, time.Time) (ports.AccessTokenRow, bool, error) {
	return ports.AccessTokenRow{}, false, nil
}

type entitlementStub struct {
	limit int
}

func (e entitlementStub) IsModuleEnabled(context.Context, kernel.TenantID, authx.Module) (bool, error) {
	return true, nil
}
func (e entitlementStub) GetSeatLimit(context.Context, kernel.TenantID) (int, error) {
	return e.limit, nil
}

func TestOIDCAuthorizeRequiresEnabledIdP(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &oidcRepoStub{idp: domain.IdentityProvider{Enabled: false}}
	svc := NewOIDCService(repo, authx.NewTokenIssuer("k", time.Hour, time.Hour), entitlementStub{limit: 10}, NewArgon2Hasher(), cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"))
	_, err := svc.AuthorizeURL(context.Background(), ports.OIDCAuthorizeCommand{
		TenantID:      tenant,
		RedirectURI:   "http://localhost/callback",
		CodeChallenge: "abc",
	})
	require.ErrorIs(t, err, domain.ErrSSONotEnabled)
}

func TestOIDCStatusRequiresEnabledIdP(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &oidcRepoStub{idp: domain.IdentityProvider{Enabled: false}}
	svc := NewOIDCService(repo, authx.NewTokenIssuer("k", time.Hour, time.Hour), entitlementStub{limit: 10}, NewArgon2Hasher(), cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"))
	status, err := svc.Status(context.Background(), tenant)
	require.NoError(t, err)
	require.False(t, status.Enabled)
}

func TestOIDCStatusReturnsProviderName(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &oidcRepoStub{idp: domain.IdentityProvider{Enabled: true, Name: "Google"}}
	svc := NewOIDCService(repo, authx.NewTokenIssuer("k", time.Hour, time.Hour), entitlementStub{limit: 10}, NewArgon2Hasher(), cache.NewInMemoryCache(), cache.NewKeyBuilder("kore"))
	status, err := svc.Status(context.Background(), tenant)
	require.NoError(t, err)
	require.True(t, status.Enabled)
	require.Equal(t, "Google", status.ProviderName)
}

func TestIdentityProviderConfigure(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &oidcRepoStub{}
	svc := NewIdentityProviderService(repo)
	idp, err := svc.Configure(context.Background(), ports.ConfigureIdPCommand{
		ID:       uuid.New(),
		TenantID: tenant,
		Name:     "Google",
		Issuer:   "https://accounts.google.com",
		ClientID: "client",
		Enabled:  true,
	})
	require.NoError(t, err)
	require.True(t, idp.Enabled)
	require.Equal(t, "Google", idp.Name)
}
