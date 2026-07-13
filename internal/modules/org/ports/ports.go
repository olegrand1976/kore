package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type LeaveTypeBootstrapper interface {
	BootstrapDefaults(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) error
}

type CreateSocieteCommand struct {
	TenantID      kernel.TenantID
	RaisonSociale string
	Devise        string
	Pays          string
}

type CreateSiteCommand struct {
	TenantID  kernel.TenantID
	SocieteID uuid.UUID
	Libelle   string
	Pays      string
}

type CreateServiceCommand struct {
	TenantID      kernel.TenantID
	SiteID        uuid.UUID
	ResponsableID uuid.UUID
}

type CreateApplicationCommand struct {
	TenantID  kernel.TenantID
	ServiceID uuid.UUID
	Libelle   string
}

type CreateUserCommand struct {
	TenantID kernel.TenantID
	Login    string
	Password string
	Profile  domain.Profile
	EquipeID *uuid.UUID
}

type UpdateUserCommand struct {
	TenantID    kernel.TenantID
	UserID      uuid.UUID
	ActorUserID uuid.UUID
	Profile     *domain.Profile
	Password    string
	Active      *bool
}

type DeleteUserCommand struct {
	TenantID    kernel.TenantID
	UserID      uuid.UUID
	ActorUserID uuid.UUID
}

type CreateClientCommand struct {
	TenantID      kernel.TenantID
	RaisonSociale string
	TVA           string
}

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	UserID       uuid.UUID
	TenantID     kernel.TenantID
	Profile      domain.Profile
}

type OrganizationRepository interface {
	SaveTenant(ctx context.Context, tenant domain.Tenant) error
	GetTenant(ctx context.Context, id kernel.TenantID) (domain.Tenant, error)
	SaveSociete(ctx context.Context, s domain.Societe) error
	UpdateSociete(ctx context.Context, s domain.Societe) error
	ListSocietes(ctx context.Context, tenant kernel.TenantID) ([]domain.Societe, error)
	GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Societe, error)
	SaveSite(ctx context.Context, s domain.Site) error
	SaveService(ctx context.Context, s domain.Service) error
	SaveApplication(ctx context.Context, a domain.Application) error
	ListApplications(ctx context.Context, tenant kernel.TenantID) ([]domain.Application, error)
	GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Application, error)
	SaveUser(ctx context.Context, u domain.User) error
	FindUserByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.User, error)
	FindUserDetailByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (UserDetail, error)
	UpdateUser(ctx context.Context, u domain.User) error
	SoftDeleteUser(ctx context.Context, tenant kernel.TenantID, id uuid.UUID, deletedAt time.Time) error
	FindUserByLogin(ctx context.Context, tenant kernel.TenantID, login string) (domain.User, error)
	FindUserByLoginGlobal(ctx context.Context, login string) (domain.User, error)
	ExistsLogin(ctx context.Context, tenant kernel.TenantID, login string) (bool, error)
	CountActiveUsers(ctx context.Context, tenant kernel.TenantID) (int, error)
	ListUsers(ctx context.Context, tenant kernel.TenantID) ([]domain.User, error)
	SaveClient(ctx context.Context, c domain.Client) error
	GetClient(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Client, error)
	ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error)
	GetPermissions(ctx context.Context) (map[string]map[authx.Module]map[authx.Action]bool, error)
	ResolveUserEmails(ctx context.Context, tenant kernel.TenantID, userIDs []uuid.UUID) ([]string, error)
	ResolveSocieteIDForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (uuid.UUID, error)
	SaveIdentityProvider(ctx context.Context, idp domain.IdentityProvider) error
	GetIdentityProvider(ctx context.Context, tenant kernel.TenantID) (domain.IdentityProvider, error)
	ListIdentityProviders(ctx context.Context, tenant kernel.TenantID) ([]domain.IdentityProvider, error)
	LinkUserIdentity(ctx context.Context, link domain.UserIdentityLink) error
	FindUserIdentityBySubject(ctx context.Context, tenant kernel.TenantID, idpID uuid.UUID, subject string) (domain.UserIdentityLink, error)
	FindUserByEmail(ctx context.Context, tenant kernel.TenantID, email string) (domain.User, error)
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hash, plain string) bool
}

type TokenIssuer interface {
	Issue(identity authx.Identity) (authx.TokenPair, error)
	ParseRefreshToken(token string) (authx.Identity, error)
}

type EntitlementReader interface {
	IsModuleEnabled(ctx context.Context, tenantID kernel.TenantID, module authx.Module) (bool, error)
	GetSeatLimit(ctx context.Context, tenantID kernel.TenantID) (int, error)
}

type UpdateSocieteBrandingCommand struct {
	TenantID      kernel.TenantID
	SocieteID     uuid.UUID
	RaisonSociale string
	Logo          string
	Adresse       string
	Siret         string
	URLTenant     string
}

type OrganizationService interface {
	CreateSociete(ctx context.Context, cmd CreateSocieteCommand) (domain.Societe, error)
	CreateSite(ctx context.Context, cmd CreateSiteCommand) (domain.Site, error)
	CreateService(ctx context.Context, cmd CreateServiceCommand) (domain.Service, error)
	CreateApplication(ctx context.Context, cmd CreateApplicationCommand) (domain.Application, error)
	ListApplications(ctx context.Context, tenant kernel.TenantID) ([]domain.Application, error)
	GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Application, error)
	ListSocietes(ctx context.Context, tenant kernel.TenantID) ([]domain.Societe, error)
	GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Societe, error)
	UpdateSocieteBranding(ctx context.Context, cmd UpdateSocieteBrandingCommand) (domain.Societe, error)
}

type UserSummary struct {
	ID      uuid.UUID `json:"id"`
	Login   string    `json:"login"`
	Prenom  string    `json:"prenom"`
	Nom     string    `json:"nom"`
	Profile string    `json:"profil"`
	Active  bool      `json:"active"`
}

type UserDetail struct {
	ID             uuid.UUID  `json:"id"`
	Login          string     `json:"login"`
	Prenom         string     `json:"prenom"`
	Nom            string     `json:"nom"`
	Email          string     `json:"email,omitempty"`
	Profile        string     `json:"profil"`
	Active         bool       `json:"active"`
	Langue         string     `json:"langue"`
	TypeCompte     string     `json:"typeCompte"`
	CraRequis      bool       `json:"craRequis"`
	SalarieETT     bool       `json:"salarieETT"`
	EquipeID       *uuid.UUID `json:"equipeId,omitempty"`
	EquipeLibelle  string     `json:"equipeLibelle,omitempty"`
	DateActivation string     `json:"dateActivation"`
	DateExpiration *string    `json:"dateExpiration,omitempty"`
}

type UserService interface {
	CreateUser(ctx context.Context, cmd CreateUserCommand) (domain.User, error)
	Authenticate(ctx context.Context, login, password string) (AuthResult, error)
	RefreshSession(ctx context.Context, refreshToken string) (authx.TokenPair, error)
	ListUsers(ctx context.Context, tenant kernel.TenantID) ([]UserSummary, error)
	GetUser(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (UserDetail, error)
	UpdateUser(ctx context.Context, cmd UpdateUserCommand) (UserSummary, error)
	DeactivateUser(ctx context.Context, cmd DeleteUserCommand) error
	DeleteUser(ctx context.Context, cmd DeleteUserCommand) error
}

type ClientService interface {
	CreateClient(ctx context.Context, cmd CreateClientCommand) (domain.Client, error)
	GetClient(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Client, error)
	ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error)
}

type TenantUsageSummary struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	SocieteName        string     `json:"societeName"`
	CreatedAt          time.Time  `json:"createdAt"`
	SubscriptionStatus string     `json:"subscriptionStatus"`
	SeatLimit          int        `json:"seatLimit"`
	ActiveUsers        int        `json:"activeUsers"`
	SeatUsagePct       float64    `json:"seatUsagePct"`
	ModulesEnabled     int        `json:"modulesEnabled"`
	CraCount           int        `json:"craCount"`
	TmaCount           int        `json:"tmaCount"`
	TmaOpen            int        `json:"tmaOpen"`
	BudgetCount        int        `json:"budgetCount"`
	LeaveCount         int        `json:"leaveCount"`
	AIRequests30d      int        `json:"aiRequests30d"`
	LastActivityAt     *time.Time `json:"lastActivityAt"`
	ActiveLast30d      bool       `json:"activeLast30d"`
}

type PlatformOverviewSummary struct {
	TotalTenants     int            `json:"totalTenants"`
	ActiveTenants30d int            `json:"activeTenants30d"`
	TotalActiveUsers int            `json:"totalActiveUsers"`
	TotalSeatLimit   int            `json:"totalSeatLimit"`
	TenantsByStatus  map[string]int `json:"tenantsByStatus"`
}

type PlatformOverview struct {
	Summary PlatformOverviewSummary `json:"summary"`
	Tenants []TenantUsageSummary    `json:"tenants"`
}

const DefaultGeminiModel = "gemini-3.5-flash"

type PlatformSettings struct {
	GeminiModel string     `json:"geminiModel"`
	LLMProvider string     `json:"llmProvider"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type UpdatePlatformSettingsCommand struct {
	GeminiModel string
	ActorUserID uuid.UUID
}

type PlatformRepository interface {
	ListTenantsUsage(ctx context.Context) ([]TenantUsageSummary, error)
	GetPlatformSettings(ctx context.Context) (PlatformSettings, error)
	SavePlatformSettings(ctx context.Context, geminiModel string, updatedBy uuid.UUID, updatedAt time.Time) error
}

type PlatformService interface {
	GetOverview(ctx context.Context) (PlatformOverview, error)
	GetSettings(ctx context.Context) (PlatformSettings, error)
	UpdateSettings(ctx context.Context, cmd UpdatePlatformSettingsCommand) (PlatformSettings, error)
	CurrentGeminiModel(ctx context.Context) string
}

type ConfigureIdPCommand struct {
	ID             uuid.UUID
	TenantID       kernel.TenantID
	Name           string
	Issuer         string
	ClientID       string
	ClientSecret   string
	JWKSURI        string
	Scopes         string
	DefaultProfile domain.Profile
	Enabled        bool
}

type OIDCAuthorizeCommand struct {
	TenantID      kernel.TenantID
	RedirectURI   string
	CodeChallenge string
	State         string
}

type OIDCCallbackCommand struct {
	TenantID     kernel.TenantID
	Code         string
	RedirectURI  string
	CodeVerifier string
	State        string
}

type OIDCService interface {
	AuthorizeURL(ctx context.Context, cmd OIDCAuthorizeCommand) (string, error)
	HandleCallback(ctx context.Context, cmd OIDCCallbackCommand) (AuthResult, error)
}

type IdentityProviderService interface {
	Configure(ctx context.Context, cmd ConfigureIdPCommand) (domain.IdentityProvider, error)
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.IdentityProvider, error)
}
