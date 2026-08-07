package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/app"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type provisionOrgRepo struct {
	tenants    []domain.Tenant
	societes   []domain.Societe
	users      []domain.User
	byLogin    map[string]domain.User
	rolledBack []kernel.TenantID
	failCore   error
}

func (r *provisionOrgRepo) FindUserByLoginGlobal(_ context.Context, login string) (domain.User, error) {
	if u, ok := r.byLogin[login]; ok {
		return u, nil
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (r *provisionOrgRepo) ProvisionCore(_ context.Context, tenant domain.Tenant, societe domain.Societe, admin domain.User) error {
	if r.failCore != nil {
		return r.failCore
	}
	r.tenants = append(r.tenants, tenant)
	r.societes = append(r.societes, societe)
	r.users = append(r.users, admin)
	if r.byLogin == nil {
		r.byLogin = map[string]domain.User{}
	}
	r.byLogin[string(admin.Login)] = admin
	return nil
}

func (r *provisionOrgRepo) RollbackProvision(_ context.Context, tenantID kernel.TenantID) error {
	r.rolledBack = append(r.rolledBack, tenantID)
	return nil
}

type stubPlatformRepo struct{}

func (stubPlatformRepo) ListTenantsUsage(context.Context) ([]ports.TenantUsageSummary, error) {
	return nil, nil
}
func (stubPlatformRepo) GetPlatformSettings(context.Context) (ports.PlatformSettings, error) {
	return ports.PlatformSettings{}, domain.ErrUserNotFound
}
func (stubPlatformRepo) SavePlatformSettings(context.Context, string, uuid.UUID, time.Time) error {
	return nil
}

type stubTrial struct {
	called bool
	seats  int
	err    error
}

func (t *stubTrial) EnsureTrial(_ context.Context, _ kernel.TenantID, seats int, _ []string) error {
	t.called = true
	t.seats = seats
	return t.err
}

type stubLeave struct {
	called bool
}

func (l *stubLeave) BootstrapDefaults(_ context.Context, _ kernel.TenantID, _ uuid.UUID) error {
	l.called = true
	return nil
}

type plainHasher struct{}

func (plainHasher) Hash(plain string) (string, error) { return "hash:" + plain, nil }
func (plainHasher) Verify(hash, plain string) bool    { return hash == "hash:"+plain }

func TestProvisionTenant_HappyPath(t *testing.T) {
	org := &provisionOrgRepo{byLogin: map[string]domain.User{}}
	trial := &stubTrial{}
	leave := &stubLeave{}
	svc := app.NewPlatformServiceFull(stubPlatformRepo{}, org, plainHasher{}, trial, leave, "gemini-test")

	got, err := svc.ProvisionTenant(context.Background(), ports.ProvisionTenantCommand{
		TenantName:    "Acme",
		RaisonSociale: "Acme SARL",
		Pays:          "MA",
		AdminLogin:    "ADM_acme",
		AdminEmail:    "admin@acme.test",
		AdminPassword: "Admin123!",
		Seats:         25,
	})
	require.NoError(t, err)
	require.Equal(t, "Acme", got.TenantName)
	require.Equal(t, "MA", got.Pays)
	require.Equal(t, "ADM_acme", got.AdminLogin)
	require.Len(t, org.tenants, 1)
	require.Len(t, org.societes, 1)
	require.Len(t, org.users, 1)
	require.Equal(t, "admin@acme.test", org.users[0].Email)
	require.True(t, trial.called)
	require.Equal(t, 25, trial.seats)
	require.True(t, leave.called)
}

func TestProvisionTenant_LoginConflict(t *testing.T) {
	org := &provisionOrgRepo{byLogin: map[string]domain.User{
		"ADM_taken": {Login: "ADM_taken"},
	}}
	svc := app.NewPlatformServiceFull(stubPlatformRepo{}, org, plainHasher{}, &stubTrial{}, &stubLeave{}, "gemini-test")

	_, err := svc.ProvisionTenant(context.Background(), ports.ProvisionTenantCommand{
		RaisonSociale: "X",
		Pays:          "FR",
		AdminLogin:    "ADM_taken",
		AdminEmail:    "a@b.co",
		AdminPassword: "Admin123!",
	})
	require.ErrorIs(t, err, domain.ErrLoginAlreadyExists)
}

func TestProvisionTenant_InvalidPays(t *testing.T) {
	org := &provisionOrgRepo{byLogin: map[string]domain.User{}}
	svc := app.NewPlatformServiceFull(stubPlatformRepo{}, org, plainHasher{}, &stubTrial{}, &stubLeave{}, "gemini-test")

	_, err := svc.ProvisionTenant(context.Background(), ports.ProvisionTenantCommand{
		RaisonSociale: "X",
		Pays:          "DE",
		AdminLogin:    "ADM_new",
		AdminEmail:    "a@b.co",
		AdminPassword: "Admin123!",
	})
	require.ErrorIs(t, err, domain.ErrInvalidPays)
}

func TestProvisionTenant_InvalidEmail(t *testing.T) {
	org := &provisionOrgRepo{byLogin: map[string]domain.User{}}
	svc := app.NewPlatformServiceFull(stubPlatformRepo{}, org, plainHasher{}, &stubTrial{}, &stubLeave{}, "gemini-test")

	_, err := svc.ProvisionTenant(context.Background(), ports.ProvisionTenantCommand{
		RaisonSociale: "X",
		Pays:          "FR",
		AdminLogin:    "ADM_new",
		AdminEmail:    "a@",
		AdminPassword: "Admin123!",
	})
	require.ErrorIs(t, err, domain.ErrInvalidEmail)
}

func TestProvisionTenant_TrialFailureRollsBack(t *testing.T) {
	org := &provisionOrgRepo{byLogin: map[string]domain.User{}}
	trial := &stubTrial{err: errors.New("billing down")}
	svc := app.NewPlatformServiceFull(stubPlatformRepo{}, org, plainHasher{}, trial, &stubLeave{}, "gemini-test")

	_, err := svc.ProvisionTenant(context.Background(), ports.ProvisionTenantCommand{
		RaisonSociale: "X",
		Pays:          "FR",
		AdminLogin:    "ADM_new",
		AdminEmail:    "admin@acme.test",
		AdminPassword: "Admin123!",
	})
	require.Error(t, err)
	require.Len(t, org.rolledBack, 1)
}

func TestSanitizeTrialModules_DropsUnknown(t *testing.T) {
	got := app.SanitizeTrialModules([]string{"cra", "evil", "ORG", "cra"})
	require.Equal(t, []string{"cra", "org"}, got)
}
