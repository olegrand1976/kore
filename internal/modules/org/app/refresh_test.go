package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type refreshUserRepo struct {
	user domain.User
	err  error
}

func (r refreshUserRepo) SaveTenant(context.Context, domain.Tenant) error { return nil }
func (r refreshUserRepo) GetTenant(context.Context, kernel.TenantID) (domain.Tenant, error) {
	return domain.Tenant{}, nil
}
func (r refreshUserRepo) SaveSociete(context.Context, domain.Societe) error { return nil }
func (r refreshUserRepo) UpdateSociete(context.Context, domain.Societe) error { return nil }
func (r refreshUserRepo) ListSocietes(context.Context, kernel.TenantID) ([]domain.Societe, error) {
	return nil, nil
}
func (r refreshUserRepo) GetSociete(context.Context, kernel.TenantID, uuid.UUID) (domain.Societe, error) {
	return domain.Societe{}, nil
}
func (r refreshUserRepo) SaveSite(context.Context, domain.Site) error { return nil }
func (r refreshUserRepo) SaveService(context.Context, domain.Service) error { return nil }
func (r refreshUserRepo) SaveApplication(context.Context, domain.Application) error { return nil }
func (r refreshUserRepo) SaveUser(context.Context, domain.User) error { return nil }
func (r refreshUserRepo) FindUserByID(context.Context, kernel.TenantID, uuid.UUID) (domain.User, error) {
	return r.user, r.err
}
func (r refreshUserRepo) UpdateUser(context.Context, domain.User) error { return nil }
func (r refreshUserRepo) SoftDeleteUser(context.Context, kernel.TenantID, uuid.UUID, time.Time) error {
	return nil
}
func (r refreshUserRepo) FindUserByLogin(context.Context, kernel.TenantID, string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}
func (r refreshUserRepo) FindUserByLoginGlobal(context.Context, string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}
func (r refreshUserRepo) ExistsLogin(context.Context, kernel.TenantID, string) (bool, error) {
	return false, nil
}
func (r refreshUserRepo) CountActiveUsers(context.Context, kernel.TenantID) (int, error) {
	return 0, nil
}
func (r refreshUserRepo) ListUsers(context.Context, kernel.TenantID) ([]domain.User, error) {
	return nil, nil
}
func (r refreshUserRepo) SaveClient(context.Context, domain.Client) error { return nil }
func (r refreshUserRepo) ListClients(context.Context, kernel.TenantID) ([]domain.Client, error) {
	return nil, nil
}
func (r refreshUserRepo) GetPermissions(context.Context) (map[string]map[authx.Module]map[authx.Action]bool, error) {
	return nil, nil
}
func (r refreshUserRepo) ResolveUserEmails(context.Context, kernel.TenantID, []uuid.UUID) ([]string, error) {
	return nil, nil
}
func (r refreshUserRepo) ResolveSocieteIDForUser(context.Context, kernel.TenantID, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (r refreshUserRepo) SaveIdentityProvider(context.Context, domain.IdentityProvider) error { return nil }
func (r refreshUserRepo) GetIdentityProvider(context.Context, kernel.TenantID) (domain.IdentityProvider, error) {
	return domain.IdentityProvider{}, domain.ErrSSONotEnabled
}
func (r refreshUserRepo) ListIdentityProviders(context.Context, kernel.TenantID) ([]domain.IdentityProvider, error) {
	return nil, nil
}
func (r refreshUserRepo) LinkUserIdentity(context.Context, domain.UserIdentityLink) error { return nil }
func (r refreshUserRepo) FindUserIdentityBySubject(context.Context, kernel.TenantID, uuid.UUID, string) (domain.UserIdentityLink, error) {
	return domain.UserIdentityLink{}, domain.ErrUserNotFound
}
func (r refreshUserRepo) FindUserByEmail(context.Context, kernel.TenantID, string) (domain.User, error) {
	return domain.User{}, domain.ErrUserNotFound
}

func TestRefreshSession_reappliesPlatformAdminRole(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	issuer := authx.NewTokenIssuer("test-signing-key", time.Hour, time.Hour)
	user := domain.User{
		ID:       userID,
		TenantID: tenant,
		Login:    "ADM_admin",
		Profile:  domain.ProfileAdmin,
		Active:   true,
		Period: domain.ActivationPeriod{
			Activation: time.Now().UTC().Add(-time.Hour),
		},
	}
	initial, err := issuer.Issue(authx.Identity{
		UserID: userID, TenantID: tenant, Profile: authx.ProfileAdmin,
		Roles: []string{authx.RolePlatformAdmin},
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	svc := NewUserService(
		refreshUserRepo{user: user},
		NewArgon2Hasher(),
		issuer,
		nil,
		nil,
		nil,
		[]string{"ADM_admin"},
	)

	pair, err := svc.RefreshSession(context.Background(), initial.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	identity, err := issuer.ParseAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse access: %v", err)
	}
	if !authx.IsPlatformAdmin(identity) {
		t.Fatalf("expected platform_admin role after refresh, got roles=%v", identity.Roles)
	}
}
