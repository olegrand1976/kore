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
	serviceType := cmd.Type
	if serviceType == "" {
		serviceType = domain.DefaultServiceType
	}
	service := domain.Service{
		ID:            uuid.New(),
		TenantID:      cmd.TenantID,
		SiteID:        cmd.SiteID,
		Libelle:       cmd.Libelle,
		Type:          serviceType,
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
		Active:    true,
	}
	return app, s.repo.SaveApplication(ctx, app)
}

func (s *organizationService) UpdateApplication(ctx context.Context, cmd ports.UpdateApplicationCommand) (domain.Application, error) {
	app, err := s.repo.GetApplication(ctx, cmd.TenantID, cmd.ApplicationID)
	if err != nil {
		return domain.Application{}, domain.ErrApplicationNotFound
	}
	if cmd.Libelle != nil {
		app.Libelle = *cmd.Libelle
	}
	if cmd.Active != nil {
		app.Active = *cmd.Active
	}
	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		return domain.Application{}, err
	}
	return app, nil
}

func (s *organizationService) SetApplicationActive(ctx context.Context, cmd ports.SetApplicationActiveCommand) (domain.Application, error) {
	return s.UpdateApplication(ctx, ports.UpdateApplicationCommand{
		TenantID:      cmd.TenantID,
		ApplicationID: cmd.ApplicationID,
		Active:        &cmd.Active,
	})
}

func (s *organizationService) CreateEquipe(ctx context.Context, cmd ports.CreateEquipeCommand) (domain.Equipe, error) {
	if cmd.ApplicationID == uuid.Nil {
		return domain.Equipe{}, domain.ErrEquipeWithoutApplication
	}
	equipe := domain.Equipe{
		ID:            uuid.New(),
		TenantID:      cmd.TenantID,
		ApplicationID: cmd.ApplicationID,
		Libelle:       cmd.Libelle,
		ResponsableID: cmd.ResponsableID,
	}
	return equipe, s.repo.SaveEquipe(ctx, equipe)
}

func (s *organizationService) ListSites(ctx context.Context, tenant kernel.TenantID) ([]domain.SiteSummary, error) {
	return s.repo.ListSites(ctx, tenant)
}

func (s *organizationService) ListApplications(ctx context.Context, tenant kernel.TenantID, filter ports.ApplicationListFilter) ([]domain.Application, error) {
	return s.repo.ListApplications(ctx, tenant, filter)
}

func (s *organizationService) ListEquipes(ctx context.Context, tenant kernel.TenantID) ([]domain.Equipe, error) {
	return s.repo.ListEquipes(ctx, tenant)
}

func (s *organizationService) ListServices(ctx context.Context, tenant kernel.TenantID) ([]domain.ServiceSummary, error) {
	return s.repo.ListServices(ctx, tenant)
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
	if cmd.CraMailRecipients != nil {
		societe.CraMailRecipients = normalizeMailRecipients(*cmd.CraMailRecipients)
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
	if cmd.CraGateMode != nil {
		mode := strings.TrimSpace(*cmd.CraGateMode)
		switch mode {
		case "block", "warn":
			societe.CraGateMode = mode
		default:
			return domain.Societe{}, fmt.Errorf("craGateMode must be block or warn")
		}
	}
	if cmd.TaskTypesEnabled != nil {
		types, err := normalizeTaskTypes(*cmd.TaskTypesEnabled)
		if err != nil {
			return domain.Societe{}, err
		}
		societe.TaskTypesEnabled = types
	}
	prevDefault := societe.TotpDefaultEnabled
	if cmd.TotpDefaultEnabled != nil {
		societe.TotpDefaultEnabled = *cmd.TotpDefaultEnabled
	}
	if cmd.TotpUserConfigurable != nil {
		societe.TotpUserConfigurable = *cmd.TotpUserConfigurable
	}
	if err := s.repo.UpdateSociete(ctx, societe); err != nil {
		return domain.Societe{}, err
	}
	if cmd.TotpDefaultEnabled != nil && societe.TotpDefaultEnabled != prevDefault {
		if societe.TotpDefaultEnabled {
			if _, err := s.repo.MarkTotpEnrollmentRequiredForSocieteUsers(ctx, cmd.TenantID, cmd.SocieteID); err != nil {
				return domain.Societe{}, err
			}
		} else {
			if err := s.repo.ClearTotpEnrollmentRequiredForSocieteUsers(ctx, cmd.TenantID, cmd.SocieteID); err != nil {
				return domain.Societe{}, err
			}
		}
	}
	return societe, nil
}

func normalizeMailRecipients(recipients []string) []string {
	seen := make(map[string]struct{}, len(recipients))
	out := make([]string, 0, len(recipients))
	for _, raw := range recipients {
		email := strings.ToLower(strings.TrimSpace(raw))
		if email == "" {
			continue
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, email)
	}
	return out
}

var allowedTaskTypes = map[string]struct{}{
	"manual": {}, "interne": {}, "formation": {}, "mission": {},
}

func normalizeTaskTypes(types []string) ([]string, error) {
	if len(types) == 0 {
		return []string{}, nil
	}
	seen := make(map[string]struct{}, len(types))
	out := make([]string, 0, len(types))
	for _, raw := range types {
		code := strings.ToLower(strings.TrimSpace(raw))
		if code == "" {
			continue
		}
		if _, ok := allowedTaskTypes[code]; !ok {
			return nil, fmt.Errorf("invalid task type %q", raw)
		}
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	return out, nil
}

func (s *organizationService) CalendarSettingsForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.UserCalendarSettings, error) {
	defaults := ports.UserCalendarSettings{
		WeekStartDay:       domain.DefaultWeekStartDay,
		DayCapacityMinutes: domain.DefaultDayCapacityMinutes,
		WeekSubmitPolicy:   domain.DefaultWeekSubmitPolicy,
		CraGateMode:        domain.DefaultCraGateMode,
		TaskTypesEnabled:   domain.EffectiveTaskTypesEnabled(nil),
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
	gateMode := societe.CraGateMode
	if gateMode != "block" && gateMode != "warn" {
		gateMode = domain.DefaultCraGateMode
	}
	return ports.UserCalendarSettings{
		WeekStartDay:       day,
		DayCapacityMinutes: cap,
		WeekSubmitPolicy:   policy,
		CraGateMode:        gateMode,
		TaskTypesEnabled:   domain.EffectiveTaskTypesEnabled(societe.TaskTypesEnabled),
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
	totpKey             []byte
}

func NewUserService(
	repo ports.OrganizationRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenIssuer,
	entitlement ports.EntitlementReader,
	appCache cache.Cache,
	keys cache.KeyBuilder,
	platformAdminLogins []string,
	totpKey []byte,
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
		totpKey:             totpKey,
	}
}

func (s *userService) CreateUser(ctx context.Context, cmd ports.CreateUserCommand) (domain.User, error) {
	login, err := domain.NewLogin(cmd.Login)
	if err != nil {
		return domain.User{}, err
	}
	if err := domain.ValidatePassword(cmd.Password); err != nil {
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
	profiles := uniqueProfiles(cmd.Profiles)
	if len(profiles) == 0 && cmd.Profile != "" {
		profiles = []domain.Profile{cmd.Profile}
	}
	if err := domain.ValidateProfiles(profiles); err != nil {
		return domain.User{}, err
	}
	equipeIDs := uniqueUUIDs(cmd.EquipeIDs)
	if len(equipeIDs) == 0 && cmd.EquipeID != nil {
		equipeIDs = []uuid.UUID{*cmd.EquipeID}
	}
	if err := s.assertEquipesInTenant(ctx, cmd.TenantID, equipeIDs); err != nil {
		return domain.User{}, err
	}
	user := domain.User{
		ID:           uuid.New(),
		TenantID:     cmd.TenantID,
		Login:        login,
		PasswordHash: hash,
		Profiles:     profiles,
		EquipeIDs:    equipeIDs,
		Active:       true,
		Period: domain.ActivationPeriod{
			Activation: s.clock().UTC().Truncate(24 * time.Hour),
		},
	}
	user.SyncPrimaryMemberships()
	if err := s.applyTotpPolicyOnCreate(ctx, &user); err != nil {
		return domain.User{}, err
	}
	return user, s.repo.SaveUser(ctx, user)
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
	profiles := make([]authx.Profile, 0, len(user.Profiles))
	for _, p := range user.Profiles {
		profiles = append(profiles, authx.Profile(p))
	}
	if len(profiles) == 0 && user.Profile != "" {
		profiles = []authx.Profile{authx.Profile(user.Profile)}
	}
	identity := authx.Identity{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Profile:  authx.Profile(user.Profile),
		Profiles: profiles,
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
	if cmd.Profiles != nil || cmd.Profile != nil {
		if cmd.UserID == cmd.ActorUserID {
			return ports.UserSummary{}, domain.ErrCannotModifySelf
		}
		var profiles []domain.Profile
		if cmd.Profiles != nil {
			profiles = uniqueProfiles(*cmd.Profiles)
		} else {
			profiles = []domain.Profile{*cmd.Profile}
		}
		if err := domain.ValidateProfiles(profiles); err != nil {
			return ports.UserSummary{}, err
		}
		user.Profiles = profiles
		user.SyncPrimaryMemberships()
	}
	if cmd.Password != "" {
		if err := domain.ValidatePassword(cmd.Password); err != nil {
			return ports.UserSummary{}, err
		}
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
	if cmd.EquipeIDs != nil {
		user.EquipeIDs = uniqueUUIDs(*cmd.EquipeIDs)
		if user.EquipeIDs == nil {
			user.EquipeIDs = []uuid.UUID{}
		}
		if err := s.assertEquipesInTenant(ctx, cmd.TenantID, user.EquipeIDs); err != nil {
			return ports.UserSummary{}, err
		}
		user.SyncPrimaryMemberships()
	} else if cmd.EquipeID != nil {
		if *cmd.EquipeID == nil {
			user.EquipeIDs = []uuid.UUID{}
		} else {
			user.EquipeIDs = []uuid.UUID{**cmd.EquipeID}
			if err := s.assertEquipesInTenant(ctx, cmd.TenantID, user.EquipeIDs); err != nil {
				return ports.UserSummary{}, err
			}
		}
		user.SyncPrimaryMemberships()
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
	profiles := make([]string, 0, len(u.Profiles))
	for _, p := range u.Profiles {
		profiles = append(profiles, string(p))
	}
	if len(profiles) == 0 && u.Profile != "" {
		profiles = []string{string(u.Profile)}
	}
	return ports.UserSummary{
		ID:        u.ID,
		Login:     string(u.Login),
		Prenom:    u.Prenom,
		Nom:       u.Nom,
		Profile:   string(u.Profile),
		Profiles:  profiles,
		Active:    u.Active,
		EquipeID:  u.EquipeID,
		EquipeIDs: u.EquipeIDs,
	}
}

func uniqueProfiles(in []domain.Profile) []domain.Profile {
	seen := make(map[domain.Profile]struct{}, len(in))
	out := make([]domain.Profile, 0, len(in))
	for _, p := range in {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func uniqueUUIDs(in []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(in))
	out := make([]uuid.UUID, 0, len(in))
	for _, id := range in {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *userService) assertEquipesInTenant(ctx context.Context, tenant kernel.TenantID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	equipes, err := s.repo.ListEquipes(ctx, tenant)
	if err != nil {
		return err
	}
	known := make(map[uuid.UUID]struct{}, len(equipes))
	for _, e := range equipes {
		known[e.ID] = struct{}{}
	}
	for _, id := range ids {
		if _, ok := known[id]; !ok {
			return domain.ErrEquipeNotFound
		}
	}
	return nil
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
			"org":       read,
			"cra":       readWriteValidate,
			"tma":       readWriteValidate,
			"conges":    read,
			"budget":    readWrite,
			"reporting": read,
		},
		"Responsable de service": {
			"org":       read,
			"cra":       readWriteValidate,
			"tma":       readWriteValidate,
			"conges":    readWriteValidate,
			"budget":    readWriteValidate,
			"reporting": read,
		},
	}
}
