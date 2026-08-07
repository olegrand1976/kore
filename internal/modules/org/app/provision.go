package app

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

const defaultTrialSeats = 10

// DefaultTrialModules matches the seed trial entitlements.
func DefaultTrialModules() []string {
	return []string{
		"org", "cra", "conges", "budget", "tma", "workflow", "notifications", "billing",
	}
}

// AllowedTrialModules is the whitelist for platform-provisioned module codes.
func AllowedTrialModules() map[string]struct{} {
	out := make(map[string]struct{})
	for _, m := range DefaultTrialModules() {
		out[m] = struct{}{}
	}
	return out
}

func SanitizeTrialModules(modules []string) []string {
	if len(modules) == 0 {
		return DefaultTrialModules()
	}
	allowed := AllowedTrialModules()
	out := make([]string, 0, len(modules))
	seen := make(map[string]struct{})
	for _, raw := range modules {
		code := strings.ToLower(strings.TrimSpace(raw))
		if _, ok := allowed[code]; !ok {
			continue
		}
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	if len(out) == 0 {
		return DefaultTrialModules()
	}
	return out
}

func (s *platformService) ProvisionTenant(ctx context.Context, cmd ports.ProvisionTenantCommand) (ports.ProvisionTenantResult, error) {
	if s.orgRepo == nil || s.hasher == nil {
		return ports.ProvisionTenantResult{}, errors.New("provisioning not configured")
	}
	if s.trial == nil {
		return ports.ProvisionTenantResult{}, errors.New("trial provisioner not configured")
	}

	tenantName := strings.TrimSpace(cmd.TenantName)
	raison := strings.TrimSpace(cmd.RaisonSociale)
	if tenantName == "" {
		tenantName = raison
	}
	if tenantName == "" || raison == "" {
		return ports.ProvisionTenantResult{}, domain.ErrProvisionInputRequired
	}

	pays, ok := domain.NormalizeSocietePays(cmd.Pays)
	if !ok {
		return ports.ProvisionTenantResult{}, domain.ErrInvalidPays
	}

	login, err := domain.NewLogin(cmd.AdminLogin)
	if err != nil {
		return ports.ProvisionTenantResult{}, err
	}
	if err := domain.ValidatePassword(cmd.AdminPassword); err != nil {
		return ports.ProvisionTenantResult{}, err
	}
	email, err := normalizeProvisionEmail(cmd.AdminEmail)
	if err != nil {
		return ports.ProvisionTenantResult{}, err
	}

	if _, err := s.orgRepo.FindUserByLoginGlobal(ctx, string(login)); err == nil {
		return ports.ProvisionTenantResult{}, domain.ErrLoginAlreadyExists
	} else if !isLoginNotFound(err) {
		return ports.ProvisionTenantResult{}, err
	}

	seats := cmd.Seats
	if seats <= 0 {
		seats = defaultTrialSeats
	}
	modules := SanitizeTrialModules(cmd.Modules)

	hash, err := s.hasher.Hash(cmd.AdminPassword)
	if err != nil {
		return ports.ProvisionTenantResult{}, err
	}

	tenantID := kernel.NewTenantID(uuid.New())
	societeID := uuid.New()
	now := time.Now().UTC()
	admin := domain.User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Login:        login,
		PasswordHash: hash,
		Email:        email,
		Profiles:     []domain.Profile{domain.ProfileAdmin},
		Profile:      domain.ProfileAdmin,
		Active:       true,
		Period: domain.ActivationPeriod{
			Activation: now.Truncate(24 * time.Hour),
		},
	}
	admin.SyncPrimaryMemberships()

	societe := domain.Societe{
		ID:                 societeID,
		TenantID:           tenantID,
		RaisonSociale:      raison,
		Devise:             "EUR",
		Pays:               pays,
		WeekStartDay:       domain.DefaultWeekStartDay,
		DayCapacityMinutes: domain.DefaultDayCapacityMinutes,
		WeekSubmitPolicy:   domain.DefaultWeekSubmitPolicy,
	}

	if err := s.orgRepo.ProvisionCore(ctx, domain.Tenant{ID: tenantID.UUID(), Name: tenantName}, societe, admin); err != nil {
		if isUniqueViolation(err) {
			return ports.ProvisionTenantResult{}, domain.ErrLoginAlreadyExists
		}
		return ports.ProvisionTenantResult{}, err
	}

	if err := s.trial.EnsureTrial(ctx, tenantID, seats, modules); err != nil {
		_ = s.orgRepo.RollbackProvision(ctx, tenantID)
		return ports.ProvisionTenantResult{}, err
	}

	// Leave defaults are best-effort: admin can reset from settings if this fails.
	if s.leave != nil {
		_ = s.leave.BootstrapDefaults(ctx, tenantID, societeID)
	}

	return ports.ProvisionTenantResult{
		TenantID:      tenantID,
		SocieteID:     societeID,
		AdminUserID:   admin.ID,
		AdminLogin:    string(login),
		TenantName:    tenantName,
		RaisonSociale: raison,
		Pays:          pays,
	}, nil
}

func normalizeProvisionEmail(raw string) (string, error) {
	email := strings.TrimSpace(strings.ToLower(raw))
	if email == "" {
		return "", domain.ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address == "" {
		return "", domain.ErrInvalidEmail
	}
	parsed := strings.ToLower(strings.TrimSpace(addr.Address))
	at := strings.IndexByte(parsed, '@')
	if at <= 0 || at == len(parsed)-1 || !strings.Contains(parsed[at+1:], ".") {
		return "", domain.ErrInvalidEmail
	}
	return parsed, nil
}

func isLoginNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, domain.ErrUserNotFound) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no rows") || strings.Contains(msg, "user not found")
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "idx_org_users_login_global")
}
