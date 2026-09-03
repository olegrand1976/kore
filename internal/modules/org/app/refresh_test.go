package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cryptox"
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
func (r refreshUserRepo) SaveSociete(context.Context, domain.Societe) error   { return nil }
func (r refreshUserRepo) UpdateSociete(context.Context, domain.Societe) error { return nil }
func (r refreshUserRepo) SaveSocieteLogo(context.Context, kernel.TenantID, uuid.UUID, []byte, string) error {
	return nil
}
func (r refreshUserRepo) GetTenantLogo(context.Context, kernel.TenantID) ([]byte, string, error) {
	return nil, "", domain.ErrLogoNotFound
}
func (r refreshUserRepo) ListSocietes(context.Context, kernel.TenantID) ([]domain.Societe, error) {
	return nil, nil
}
func (r refreshUserRepo) GetSociete(context.Context, kernel.TenantID, uuid.UUID) (domain.Societe, error) {
	return domain.Societe{}, nil
}
func (r refreshUserRepo) SaveSite(context.Context, domain.Site) error { return nil }
func (r refreshUserRepo) UpdateSite(context.Context, kernel.TenantID, uuid.UUID, string) (domain.SiteSummary, error) {
	return domain.SiteSummary{}, domain.ErrSiteNotFound
}
func (r refreshUserRepo) SaveService(context.Context, domain.Service) error { return nil }
func (r refreshUserRepo) UpdateService(context.Context, kernel.TenantID, uuid.UUID, string) (domain.ServiceSummary, error) {
	return domain.ServiceSummary{}, domain.ErrServiceNotFound
}
func (r refreshUserRepo) SaveApplication(context.Context, domain.Application) error { return nil }
func (r refreshUserRepo) SaveEquipe(context.Context, domain.Equipe) error           { return nil }
func (r refreshUserRepo) GetEquipe(context.Context, kernel.TenantID, uuid.UUID) (domain.Equipe, error) {
	return domain.Equipe{}, domain.ErrEquipeNotFound
}
func (r refreshUserRepo) UpdateEquipe(context.Context, kernel.TenantID, uuid.UUID, string, *uuid.UUID) (domain.Equipe, error) {
	return domain.Equipe{}, domain.ErrEquipeNotFound
}
func (r refreshUserRepo) ListSites(context.Context, kernel.TenantID) ([]domain.SiteSummary, error) {
	return nil, nil
}
func (r refreshUserRepo) ListApplications(context.Context, kernel.TenantID, ports.ApplicationListFilter) ([]domain.Application, error) {
	return nil, nil
}
func (r refreshUserRepo) UpdateApplication(context.Context, domain.Application, bool) error {
	return nil
}
func (r refreshUserRepo) ListEquipes(context.Context, kernel.TenantID, ports.EquipeListFilter) ([]domain.Equipe, error) {
	return nil, nil
}
func (r refreshUserRepo) ListServices(context.Context, kernel.TenantID) ([]domain.ServiceSummary, error) {
	return nil, nil
}
func (r refreshUserRepo) GetApplication(context.Context, kernel.TenantID, uuid.UUID) (domain.Application, error) {
	return domain.Application{}, domain.ErrUserNotFound
}
func (r refreshUserRepo) AssertApplicationSharesExist(context.Context, kernel.TenantID, []uuid.UUID, []uuid.UUID, []uuid.UUID) error {
	return nil
}
func (r refreshUserRepo) BudgetBelongsToApplication(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) (bool, error) {
	return false, nil
}
func (r refreshUserRepo) CountProjectArtifacts(context.Context, kernel.TenantID, uuid.UUID) (int, error) {
	return 0, nil
}
func (r refreshUserRepo) MergeApplications(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) (domain.Application, error) {
	return domain.Application{}, domain.ErrApplicationNotFound
}
func (r refreshUserRepo) RecordAdminAuditEvent(context.Context, kernel.TenantID, uuid.UUID, string, map[string]any) error {
	return nil
}
func (r refreshUserRepo) SaveUser(context.Context, domain.User) error { return nil }
func (r refreshUserRepo) FindUserByID(context.Context, kernel.TenantID, uuid.UUID) (domain.User, error) {
	return r.user, r.err
}
func (r refreshUserRepo) FindUserDetailByID(context.Context, kernel.TenantID, uuid.UUID) (ports.UserDetail, error) {
	return ports.UserDetail{}, domain.ErrUserNotFound
}
func (r refreshUserRepo) GetReleaseNotesPreferences(context.Context, kernel.TenantID, uuid.UUID) (ports.ReleaseNotesPreferences, error) {
	return ports.ReleaseNotesPreferences{LastSeenVersion: nil, AutoShowEnabled: true}, nil
}
func (r refreshUserRepo) SetReleaseNotesAutoShow(context.Context, kernel.TenantID, uuid.UUID, bool) error {
	return nil
}
func (r refreshUserRepo) SetLastSeenVersion(context.Context, kernel.TenantID, uuid.UUID, string) error {
	return nil
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
func (r refreshUserRepo) SaveClient(context.Context, domain.Client) error   { return nil }
func (r refreshUserRepo) UpdateClient(context.Context, domain.Client) error { return nil }
func (r refreshUserRepo) UpdateClientContacts(context.Context, kernel.TenantID, uuid.UUID, []domain.ClientContact) error {
	return nil
}
func (r refreshUserRepo) ListClients(context.Context, kernel.TenantID) ([]domain.Client, error) {
	return nil, nil
}
func (r refreshUserRepo) GetClient(context.Context, kernel.TenantID, uuid.UUID) (domain.Client, error) {
	return domain.Client{}, domain.ErrClientNotFound
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
func (r refreshUserRepo) ResolveSocieteIDForEquipe(context.Context, kernel.TenantID, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (r refreshUserRepo) ListSocietesCraMailAuto(context.Context) ([]ports.CraMailReminderTarget, error) {
	return nil, nil
}
func (r refreshUserRepo) SaveIdentityProvider(context.Context, domain.IdentityProvider) error {
	return nil
}
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

func (r refreshUserRepo) FindTenantIDsByEmail(context.Context, string) ([]kernel.TenantID, error) {
	return nil, nil
}
func (r refreshUserRepo) SaveAccessToken(context.Context, string, kernel.TenantID, string, string, time.Time) error {
	return nil
}
func (r refreshUserRepo) ConsumeAccessToken(context.Context, string, time.Time) (ports.AccessTokenRow, bool, error) {
	return ports.AccessTokenRow{}, false, nil
}
func (r refreshUserRepo) UpdateUserTotp(context.Context, domain.User) error { return nil }
func (r refreshUserRepo) SaveTotpBackupCodes(context.Context, kernel.TenantID, uuid.UUID, []string) error {
	return nil
}
func (r refreshUserRepo) ConsumeTotpBackupCode(context.Context, kernel.TenantID, uuid.UUID, string, time.Time) (bool, error) {
	return false, nil
}
func (r refreshUserRepo) DeleteTotpBackupCodes(context.Context, kernel.TenantID, uuid.UUID) error {
	return nil
}
func (r refreshUserRepo) ListUnusedTotpBackupCodeHashes(context.Context, kernel.TenantID, uuid.UUID) ([]string, error) {
	return nil, nil
}
func (r refreshUserRepo) MarkTotpEnrollmentRequiredForSocieteUsers(context.Context, kernel.TenantID, uuid.UUID) (int, error) {
	return 0, nil
}
func (r refreshUserRepo) ClearTotpEnrollmentRequiredForSocieteUsers(context.Context, kernel.TenantID, uuid.UUID) error {
	return nil
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
		cryptox.DevKeyFromJWTSigningKey("test-signing-key"),
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
