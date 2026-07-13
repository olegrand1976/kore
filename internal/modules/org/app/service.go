package app

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
	"golang.org/x/crypto/argon2"
)

type organizationService struct {
	repo ports.OrganizationRepository
}

func NewOrganizationService(repo ports.OrganizationRepository) ports.OrganizationService {
	return &organizationService{repo: repo}
}

func (s *organizationService) CreateSociete(ctx context.Context, cmd ports.CreateSocieteCommand) (domain.Societe, error) {
	societe := domain.Societe{
		ID:                 uuid.New(),
		TenantID:           cmd.TenantID,
		RaisonSociale:      cmd.RaisonSociale,
		Devise:             cmd.Devise,
		Pays:               cmd.Pays,
		WeekStartDay:       domain.DefaultWeekStartDay,
		DayCapacityMinutes: domain.DefaultDayCapacityMinutes,
		WeekSubmitPolicy:   domain.DefaultWeekSubmitPolicy,
	}
	if societe.Devise == "" {
		societe.Devise = "EUR"
	}
	if societe.Pays == "" {
		societe.Pays = "FR"
	}
	return societe, s.repo.SaveSociete(ctx, societe)
}

func (s *organizationService) CreateSite(ctx context.Context, cmd ports.CreateSiteCommand) (domain.Site, error) {
	pays := cmd.Pays
	if pays == "" {
		societe, err := s.repo.GetSociete(ctx, cmd.TenantID, cmd.SocieteID)
		if err == nil && societe.Pays != "" {
			pays = societe.Pays
		} else {
			pays = "FR"
		}
	}
	site := domain.Site{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		SocieteID: cmd.SocieteID,
		Libelle:   cmd.Libelle,
		Pays:      pays,
	}
	return site, s.repo.SaveSite(ctx, site)
}

func (s *organizationService) CreateService(ctx context.Context, cmd ports.CreateServiceCommand) (domain.Service, error) {
	if cmd.ResponsableID == uuid.Nil {
		return domain.Service{}, domain.ErrServiceWithoutResponsible
	}
	service := domain.Service{
		ID:            uuid.New(),
		TenantID:      cmd.TenantID,
		SiteID:        cmd.SiteID,
		ResponsableID: &cmd.ResponsableID,
	}
	return service, s.repo.SaveService(ctx, service)
}

func (s *organizationService) CreateApplication(ctx context.Context, cmd ports.CreateApplicationCommand) (domain.Application, error) {
	app := domain.Application{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		ServiceID: cmd.ServiceID,
		Libelle:   cmd.Libelle,
	}
	return app, s.repo.SaveApplication(ctx, app)
}

func (s *organizationService) ListApplications(ctx context.Context, tenant kernel.TenantID) ([]domain.Application, error) {
	return s.repo.ListApplications(ctx, tenant)
}

func (s *organizationService) GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Application, error) {
	return s.repo.GetApplication(ctx, tenant, id)
}

func (s *organizationService) ListSocietes(ctx context.Context, tenant kernel.TenantID) ([]domain.Societe, error) {
	return s.repo.ListSocietes(ctx, tenant)
}

func (s *organizationService) GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Societe, error) {
	return s.repo.GetSociete(ctx, tenant, id)
}

func (s *organizationService) UpdateSocieteBranding(ctx context.Context, cmd ports.UpdateSocieteBrandingCommand) (domain.Societe, error) {
	societe, err := s.repo.GetSociete(ctx, cmd.TenantID, cmd.SocieteID)
	if err != nil {
		return domain.Societe{}, err
	}
	if cmd.RaisonSociale != "" {
		societe.RaisonSociale = cmd.RaisonSociale
	}
	if cmd.Logo != "" {
		societe.Logo = cmd.Logo
	}
	if cmd.Adresse != "" {
		societe.Adresse = cmd.Adresse
	}
	if cmd.Siret != "" {
		societe.Siret = cmd.Siret
	}
	if cmd.URLTenant != "" {
		societe.URLTenant = cmd.URLTenant
	}
	if err := s.repo.UpdateSociete(ctx, societe); err != nil {
		return domain.Societe{}, err
	}
	return societe, nil
}

func (s *organizationService) UpdateSocieteSettings(ctx context.Context, cmd ports.UpdateSocieteSettingsCommand) (domain.Societe, error) {
	societe, err := s.repo.GetSociete(ctx, cmd.TenantID, cmd.SocieteID)
	if err != nil {
		return domain.Societe{}, err
	}
	if cmd.WeekStartDay != nil {
		day := *cmd.WeekStartDay
		if day < 0 || day > 6 {
			return domain.Societe{}, fmt.Errorf("weekStartDay must be between 0 and 6")
		}
		societe.WeekStartDay = day
	}
	if cmd.DayCapacityMinutes != nil {
		cap := *cmd.DayCapacityMinutes
		if cap <= 0 || cap > 1440 {
			return domain.Societe{}, fmt.Errorf("dayCapacityMinutes must be between 1 and 1440")
		}
		societe.DayCapacityMinutes = cap
	}
	if cmd.CraMailAuto != nil {
		societe.CraMailAuto = *cmd.CraMailAuto
	}
	if cmd.WeekSubmitPolicy != nil {
		policy := strings.TrimSpace(*cmd.WeekSubmitPolicy)
		switch policy {
		case "block", "warn", "none":
			societe.WeekSubmitPolicy = policy
		default:
			return domain.Societe{}, fmt.Errorf("weekSubmitPolicy must be block, warn, or none")
		}
	}
	if err := s.repo.UpdateSociete(ctx, societe); err != nil {
		return domain.Societe{}, err
	}
	return societe, nil
}

func (s *organizationService) CalendarSettingsForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.UserCalendarSettings, error) {
	defaults := ports.UserCalendarSettings{
		WeekStartDay:       domain.DefaultWeekStartDay,
		DayCapacityMinutes: domain.DefaultDayCapacityMinutes,
		WeekSubmitPolicy:   domain.DefaultWeekSubmitPolicy,
	}
	societeID, err := s.repo.ResolveSocieteIDForUser(ctx, tenant, userID)
	if err != nil {
		return defaults, nil
	}
	societe, err := s.repo.GetSociete(ctx, tenant, societeID)
	if err != nil {
		return defaults, nil
	}
	day := societe.WeekStartDay
	if day < 0 || day > 6 {
		day = domain.DefaultWeekStartDay
	}
	cap := societe.DayCapacityMinutes
	if cap <= 0 || cap > 1440 {
		cap = domain.DefaultDayCapacityMinutes
	}
	policy := societe.WeekSubmitPolicy
	if policy != "block" && policy != "warn" && policy != "none" {
		policy = domain.DefaultWeekSubmitPolicy
	}
	return ports.UserCalendarSettings{
		WeekStartDay:       day,
		DayCapacityMinutes: cap,
		WeekSubmitPolicy:   policy,
	}, nil
}

type userService struct {
	repo                ports.OrganizationRepository
	hasher              ports.PasswordHasher
	tokens              ports.TokenIssuer
	entitlement         ports.EntitlementReader
	cache               cache.Cache
	keys                cache.KeyBuilder
	clock               func() time.Time
	platformAdminLogins map[string]struct{}
}

func NewUserService(
	repo ports.OrganizationRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenIssuer,
	entitlement ports.EntitlementReader,
	appCache cache.Cache,
	keys cache.KeyBuilder,
	platformAdminLogins []string,
) ports.UserService {
	logins := make(map[string]struct{}, len(platformAdminLogins))
	for _, login := range platformAdminLogins {
		logins[strings.ToUpper(login)] = struct{}{}
	}
	return &userService{
		repo:                repo,
		hasher:              hasher,
		tokens:              tokens,
		entitlement:         entitlement,
		cache:               appCache,
		keys:                keys,
		clock:               time.Now,
		platformAdminLogins: logins,
	}
}

func (s *userService) CreateUser(ctx context.Context, cmd ports.CreateUserCommand) (domain.User, error) {
	login, err := domain.NewLogin(cmd.Login)
	if err != nil {
		return domain.User{}, err
	}
	exists, err := s.repo.ExistsLogin(ctx, cmd.TenantID, string(login))
	if err != nil {
		return domain.User{}, err
	}
	if exists {
		return domain.User{}, domain.ErrLoginAlreadyExists
	}
	limit, err := s.entitlement.GetSeatLimit(ctx, cmd.TenantID)
	if err != nil {
		return domain.User{}, err
	}
	count, err := s.repo.CountActiveUsers(ctx, cmd.TenantID)
	if err != nil {
		return domain.User{}, err
	}
	if limit > 0 && count >= limit {
		return domain.User{}, domain.ErrSeatLimitReached
	}
	hash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return domain.User{}, err
	}
	profile := cmd.Profile
	if profile == "" {
		profile = domain.ProfileCollaborateur
	}
	user := domain.User{
		ID:           uuid.New(),
		TenantID:     cmd.TenantID,
		EquipeID:     cmd.EquipeID,
		Login:        login,
		PasswordHash: hash,
		Profile:      profile,
		Active:       true,
		Period: domain.ActivationPeriod{
			Activation: s.clock().UTC().Truncate(24 * time.Hour),
		},
	}
	return user, s.repo.SaveUser(ctx, user)
}

func (s *userService) Authenticate(ctx context.Context, login, password string) (ports.AuthResult, error) {
	user, err := s.repo.FindUserByLoginGlobal(ctx, login)
	if err != nil {
		return ports.AuthResult{}, domain.ErrInvalidCredentials
	}
	if !user.Active || !user.Period.IsActive(s.clock()) {
		return ports.AuthResult{}, domain.ErrAccountExpired
	}
	if !s.hasher.Verify(user.PasswordHash, password) {
		return ports.AuthResult{}, domain.ErrInvalidCredentials
	}
	pair, err := s.tokens.Issue(s.buildIdentity(user))
	if err != nil {
		return ports.AuthResult{}, err
	}
	return ports.AuthResult{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		UserID:       user.ID,
		TenantID:     user.TenantID,
		Profile:      user.Profile,
	}, nil
}

func (s *userService) RefreshSession(ctx context.Context, refreshToken string) (authx.TokenPair, error) {
	identity, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return authx.TokenPair{}, domain.ErrInvalidCredentials
	}
	user, err := s.repo.FindUserByID(ctx, identity.TenantID, identity.UserID)
	if err != nil {
		return authx.TokenPair{}, domain.ErrInvalidCredentials
	}
	if !user.Active || !user.Period.IsActive(s.clock()) {
		return authx.TokenPair{}, domain.ErrAccountExpired
	}
	return s.tokens.Issue(s.buildIdentity(user))
}

func (s *userService) buildIdentity(user domain.User) authx.Identity {
	identity := authx.Identity{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Profile:  authx.Profile(user.Profile),
	}
	if _, ok := s.platformAdminLogins[strings.ToUpper(string(user.Login))]; ok {
		identity.Roles = []string{authx.RolePlatformAdmin}
	}
	return identity
}

func (s *userService) ListUsers(ctx context.Context, tenant kernel.TenantID) ([]ports.UserSummary, error) {
	users, err := s.repo.ListUsers(ctx, tenant)
	if err != nil {
		return nil, err
	}
	out := make([]ports.UserSummary, 0, len(users))
	for _, u := range users {
		out = append(out, userToSummary(u))
	}
	return out, nil
}

func (s *userService) GetUser(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (ports.UserDetail, error) {
	detail, err := s.repo.FindUserDetailByID(ctx, tenant, id)
	if err != nil {
		return ports.UserDetail{}, domain.ErrUserNotFound
	}
	return detail, nil
}

func (s *userService) UpdateUser(ctx context.Context, cmd ports.UpdateUserCommand) (ports.UserSummary, error) {
	user, err := s.repo.FindUserByID(ctx, cmd.TenantID, cmd.UserID)
	if err != nil {
		return ports.UserSummary{}, domain.ErrUserNotFound
	}
	if cmd.Profile != nil {
		user.Profile = *cmd.Profile
	}
	if cmd.Password != "" {
		hash, err := s.hasher.Hash(cmd.Password)
		if err != nil {
			return ports.UserSummary{}, err
		}
		user.PasswordHash = hash
	}
	if cmd.Active != nil {
		if !*cmd.Active && cmd.UserID == cmd.ActorUserID {
			return ports.UserSummary{}, domain.ErrCannotModifySelf
		}
		if *cmd.Active && !user.Active {
			limit, err := s.entitlement.GetSeatLimit(ctx, cmd.TenantID)
			if err != nil {
				return ports.UserSummary{}, err
			}
			count, err := s.repo.CountActiveUsers(ctx, cmd.TenantID)
			if err != nil {
				return ports.UserSummary{}, err
			}
			if limit > 0 && count >= limit {
				return ports.UserSummary{}, domain.ErrSeatLimitReached
			}
		}
		user.Active = *cmd.Active
	}
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return ports.UserSummary{}, domain.ErrUserNotFound
	}
	return userToSummary(user), nil
}

func (s *userService) DeactivateUser(ctx context.Context, cmd ports.DeleteUserCommand) error {
	if cmd.UserID == cmd.ActorUserID {
		return domain.ErrCannotModifySelf
	}
	user, err := s.repo.FindUserByID(ctx, cmd.TenantID, cmd.UserID)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if !user.Active {
		return nil
	}
	user.Active = false
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, cmd ports.DeleteUserCommand) error {
	if cmd.UserID == cmd.ActorUserID {
		return domain.ErrCannotModifySelf
	}
	if err := s.repo.SoftDeleteUser(ctx, cmd.TenantID, cmd.UserID, s.clock().UTC()); err != nil {
		return domain.ErrUserNotFound
	}
	return nil
}

func (s *userService) GetReleaseNotesPreferences(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.ReleaseNotesPreferences, error) {
	return s.repo.GetReleaseNotesPreferences(ctx, tenant, userID)
}

func (s *userService) SetReleaseNotesAutoShow(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, enabled bool) error {
	return s.repo.SetReleaseNotesAutoShow(ctx, tenant, userID, enabled)
}

func (s *userService) MarkReleaseNotesSeen(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, version string) error {
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("version is required")
	}
	return s.repo.SetLastSeenVersion(ctx, tenant, userID, version)
}

func userToSummary(u domain.User) ports.UserSummary {
	return ports.UserSummary{
		ID:      u.ID,
		Login:   string(u.Login),
		Prenom:  u.Prenom,
		Nom:     u.Nom,
		Profile: string(u.Profile),
		Active:  u.Active,
	}
}

type clientService struct {
	repo ports.OrganizationRepository
}

func NewClientService(repo ports.OrganizationRepository) ports.ClientService {
	return &clientService{repo: repo}
}

func (s *clientService) CreateClient(ctx context.Context, cmd ports.CreateClientCommand) (domain.Client, error) {
	client := domain.Client{
		ID:            uuid.New(),
		TenantID:      cmd.TenantID,
		RaisonSociale: cmd.RaisonSociale,
		TVA:           cmd.TVA,
	}
	return client, s.repo.SaveClient(ctx, client)
}

func (s *clientService) ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error) {
	return s.repo.ListClients(ctx, tenant)
}

func (s *clientService) GetClient(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Client, error) {
	return s.repo.GetClient(ctx, tenant, id)
}

type argon2Hasher struct{}

func NewArgon2Hasher() ports.PasswordHasher { return &argon2Hasher{} }

func (h *argon2Hasher) Hash(plain string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(plain), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func (h *argon2Hasher) Verify(hash, plain string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 4 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	key := argon2.IDKey([]byte(plain), salt, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(key, expected) == 1
}

func DefaultPermissions() map[string]map[authx.Module]map[authx.Action]bool {
	read := map[authx.Action]bool{authx.ActionRead: true}
	readWrite := map[authx.Action]bool{authx.ActionRead: true, authx.ActionWrite: true}
	readWriteValidate := map[authx.Action]bool{
		authx.ActionRead: true, authx.ActionWrite: true, authx.ActionValidate: true,
	}
	mvpAdmin := map[authx.Module]map[authx.Action]bool{
		"org":           readWriteValidate,
		"cra":           readWriteValidate,
		"tma":           readWriteValidate,
		"conges":        readWriteValidate,
		"budget":        readWriteValidate,
		"workflow":      readWriteValidate,
		"billing":       readWrite,
		"notifications": readWrite,
		"integrations":  readWriteValidate,
		"invoicing":     readWriteValidate,
		"admin":         readWriteValidate,
		"reporting":     read,
		"ssii":          readWriteValidate,
		"ett":           readWriteValidate,
		"support":       readWriteValidate,
		"maintenance":   readWriteValidate,
	}
	return map[string]map[authx.Module]map[authx.Action]bool{
		string(domain.ProfileAdmin): mvpAdmin,
		string(domain.ProfileCollaborateur): {
			"cra":    readWrite,
			"tma":    readWrite,
			"conges": readWrite,
			"budget": read,
		},
		"Chef d'équipe": {
			"org":    read,
			"cra":    readWriteValidate,
			"tma":    readWriteValidate,
			"conges": read,
			"budget": readWrite,
		},
		"Responsable de service": {
			"org":    read,
			"cra":    readWriteValidate,
			"tma":    readWriteValidate,
			"conges": readWriteValidate,
			"budget": readWriteValidate,
		},
	}
}
