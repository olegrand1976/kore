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
	AccessToken           string         `json:"accessToken"`
	RefreshToken          string         `json:"refreshToken"`
	UserID                uuid.UUID      `json:"userId"`
	TenantID              kernel.TenantID `json:"tenantId"`
	Profile               domain.Profile `json:"profile"`
	Requires2FA           bool           `json:"requires2FA,omitempty"`
	ChallengeToken        string         `json:"challengeToken,omitempty"`
	Requires2FAEnrollment bool           `json:"requires2FAEnrollment,omitempty"`
	EnrollmentToken       string         `json:"enrollmentToken,omitempty"`
	BackupCodes           []string       `json:"backupCodes,omitempty"`
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
	GetReleaseNotesPreferences(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ReleaseNotesPreferences, error)
	SetReleaseNotesAutoShow(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, enabled bool) error
	SetLastSeenVersion(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, version string) error
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
	ResolveSocieteIDForEquipe(ctx context.Context, tenant kernel.TenantID, equipeID uuid.UUID) (uuid.UUID, error)
	ListSocietesCraMailAuto(ctx context.Context) ([]CraMailReminderTarget, error)
	SaveIdentityProvider(ctx context.Context, idp domain.IdentityProvider) error
	GetIdentityProvider(ctx context.Context, tenant kernel.TenantID) (domain.IdentityProvider, error)
	ListIdentityProviders(ctx context.Context, tenant kernel.TenantID) ([]domain.IdentityProvider, error)
	LinkUserIdentity(ctx context.Context, link domain.UserIdentityLink) error
	FindUserIdentityBySubject(ctx context.Context, tenant kernel.TenantID, idpID uuid.UUID, subject string) (domain.UserIdentityLink, error)
	FindUserByEmail(ctx context.Context, tenant kernel.TenantID, email string) (domain.User, error)
	FindTenantIDsByEmail(ctx context.Context, email string) ([]kernel.TenantID, error)
	SaveAccessToken(ctx context.Context, tokenHash string, tenant kernel.TenantID, email, kind string, expiresAt time.Time) error
	ConsumeAccessToken(ctx context.Context, tokenHash string, now time.Time) (AccessTokenRow, bool, error)
	UpdateUserTotp(ctx context.Context, u domain.User) error
	SaveTotpBackupCodes(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, codeHashes []string) error
	ConsumeTotpBackupCode(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, codeHash string, usedAt time.Time) (bool, error)
	DeleteTotpBackupCodes(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) error
	ListUnusedTotpBackupCodeHashes(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]string, error)
	MarkTotpEnrollmentRequiredForSocieteUsers(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) (int, error)
	ClearTotpEnrollmentRequiredForSocieteUsers(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) error
}

type AccessTokenRow struct {
	TenantID  kernel.TenantID
	Email     string
	Kind      string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	TokenHash string
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

type UpdateSocieteSettingsCommand struct {
	TenantID             kernel.TenantID
	SocieteID            uuid.UUID
	WeekStartDay         *int
	DayCapacityMinutes   *int
	CraMailAuto          *bool
	CraMailRecipients    *[]string
	WeekSubmitPolicy     *string
	TaskTypesEnabled     *[]string
	TotpDefaultEnabled   *bool
	TotpUserConfigurable *bool
}

type CraMailReminderTarget struct {
	TenantID   kernel.TenantID
	SocieteID  uuid.UUID
	Pays       string
	Recipients []string
}

type UserCalendarSettings struct {
	WeekStartDay       int      `json:"weekStartDay"`
	DayCapacityMinutes int      `json:"dayCapacityMinutes"`
	WeekSubmitPolicy   string   `json:"weekSubmitPolicy"`
	TaskTypesEnabled   []string `json:"taskTypesEnabled"`
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
	UpdateSocieteSettings(ctx context.Context, cmd UpdateSocieteSettingsCommand) (domain.Societe, error)
	CalendarSettingsForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (UserCalendarSettings, error)
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

type TotpPolicy struct {
	DefaultEnabled   bool
	UserConfigurable bool
}

type TotpStatus struct {
	Enabled            bool       `json:"enabled"`
	EnrollmentRequired bool       `json:"enrollmentRequired"`
	UserConfigurable   bool       `json:"userConfigurable"`
	OrgDefaultEnabled  bool       `json:"orgDefaultEnabled"`
	EnabledAt          *time.Time `json:"enabledAt,omitempty"`
	PasswordLogin      bool       `json:"passwordLogin"`
}

type TotpSetupResult struct {
	OtpauthURL    string `json:"otpauthUrl"`
	Secret        string `json:"secret"`
	QrCodeDataURL string `json:"qrCodeDataUrl"`
}

type TotpConfirmResult struct {
	BackupCodes []string `json:"backupCodes"`
}

type TotpApplyPolicyResult struct {
	UsersMarked int `json:"usersMarked"`
}

type ReleaseNotesPreferences struct {
	LastSeenVersion *string `json:"lastSeenVersion"`
	AutoShowEnabled bool    `json:"autoShowEnabled"`
}

type Confirm2FACommand struct {
	TenantID        kernel.TenantID
	UserID          uuid.UUID
	Code            string
	Password        string
	EnrollmentToken string
	SkipPassword    bool
}

type Disable2FACommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	Password string
	Code     string
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
	GetReleaseNotesPreferences(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ReleaseNotesPreferences, error)
	SetReleaseNotesAutoShow(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, enabled bool) error
	MarkReleaseNotesSeen(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, version string) error
	Get2FAStatus(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (TotpStatus, error)
	Setup2FA(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (TotpSetupResult, error)
	Confirm2FA(ctx context.Context, cmd Confirm2FACommand) (TotpConfirmResult, error)
	Disable2FA(ctx context.Context, cmd Disable2FACommand) error
	Verify2FAChallenge(ctx context.Context, challengeToken, code string) (AuthResult, error)
	Verify2FAEnrollment(ctx context.Context, enrollmentToken, code, password string) (AuthResult, error)
	Setup2FAWithEnrollmentToken(ctx context.Context, enrollmentToken string) (TotpSetupResult, error)
	ApplyTotpPolicyOnSociete(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, defaultEnabled bool) (TotpApplyPolicyResult, error)
	ClearTotpEnrollmentRequiredOnSociete(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) error
}

type ClientService interface {
	CreateClient(ctx context.Context, cmd CreateClientCommand) (domain.Client, error)
	GetClient(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Client, error)
	ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error)
}

type TenantAccessResolveResult struct {
	TenantID kernel.TenantID `json:"tenantId"`
	Kind     string          `json:"kind"`
}

type TenantAccessRepository interface {
	FindTenantIDsByEmail(ctx context.Context, email string) ([]kernel.TenantID, error)
	SaveAccessToken(ctx context.Context, tokenHash string, tenant kernel.TenantID, email, kind string, expiresAt time.Time) error
	ConsumeAccessToken(ctx context.Context, tokenHash string, now time.Time) (AccessTokenRow, bool, error)
}

type TenantAccessService interface {
	RequestTenantDiscovery(ctx context.Context, email string, baseLoginURL string) error
	CreateInvitation(ctx context.Context, tenant kernel.TenantID, email string, baseLoginURL string) error
	Resolve(ctx context.Context, token string) (TenantAccessResolveResult, error)
}

type TransactionalEmailSender interface {
	SendTenantAccessEmail(ctx context.Context, to string, subject string, body string) error
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
	Status(ctx context.Context, tenant kernel.TenantID) (OIDCStatus, error)
}

type OIDCStatus struct {
	Enabled      bool   `json:"enabled"`
	ProviderName string `json:"providerName"`
}

type IdentityProviderService interface {
	Configure(ctx context.Context, cmd ConfigureIdPCommand) (domain.IdentityProvider, error)
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.IdentityProvider, error)
}
