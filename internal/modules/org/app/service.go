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
		ID:            uuid.New(),
		TenantID:      cmd.TenantID,
		RaisonSociale: cmd.RaisonSociale,
		Devise:        cmd.Devise,
	}
	if societe.Devise == "" {
		societe.Devise = "EUR"
	}
	return societe, s.repo.SaveSociete(ctx, societe)
}

func (s *organizationService) CreateSite(ctx context.Context, cmd ports.CreateSiteCommand) (domain.Site, error) {
	site := domain.Site{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		SocieteID: cmd.SocieteID,
		Libelle:   cmd.Libelle,
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

type userService struct {
	repo        ports.OrganizationRepository
	hasher      ports.PasswordHasher
	tokens      ports.TokenIssuer
	entitlement ports.EntitlementReader
	cache       cache.Cache
	keys        cache.KeyBuilder
	clock       func() time.Time
}

func NewUserService(
	repo ports.OrganizationRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenIssuer,
	entitlement ports.EntitlementReader,
	appCache cache.Cache,
	keys cache.KeyBuilder,
) ports.UserService {
	return &userService{
		repo:        repo,
		hasher:      hasher,
		tokens:      tokens,
		entitlement: entitlement,
		cache:       appCache,
		keys:        keys,
		clock:       time.Now,
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
	pair, err := s.tokens.Issue(authx.Identity{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Profile:  authx.Profile(user.Profile),
	})
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

func (s *userService) ListUsers(ctx context.Context, tenant kernel.TenantID) ([]ports.UserSummary, error) {
	users, err := s.repo.ListUsers(ctx, tenant)
	if err != nil {
		return nil, err
	}
	out := make([]ports.UserSummary, 0, len(users))
	for _, u := range users {
		out = append(out, ports.UserSummary{
			ID:      u.ID,
			Login:   string(u.Login),
			Profile: string(u.Profile),
			Active:  u.Active,
		})
	}
	return out, nil
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
	return map[string]map[authx.Module]map[authx.Action]bool{
		string(domain.ProfileAdmin): {
			"org":  {authx.ActionRead: true, authx.ActionWrite: true, authx.ActionValidate: true},
			"cra":  {authx.ActionRead: true, authx.ActionWrite: true, authx.ActionValidate: true},
			"tma":  {authx.ActionRead: true, authx.ActionWrite: true, authx.ActionValidate: true},
			"billing": {authx.ActionRead: true, authx.ActionWrite: true},
			"notifications": {authx.ActionRead: true, authx.ActionWrite: true},
		},
		string(domain.ProfileCollaborateur): {
			"cra": {authx.ActionRead: true, authx.ActionWrite: true},
		},
	}
}
