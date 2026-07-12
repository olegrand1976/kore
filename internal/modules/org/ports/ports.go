package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type CreateSocieteCommand struct {
	TenantID      kernel.TenantID
	RaisonSociale string
	Devise        string
}

type CreateSiteCommand struct {
	TenantID  kernel.TenantID
	SocieteID uuid.UUID
	Libelle   string
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
	ListSocietes(ctx context.Context, tenant kernel.TenantID) ([]domain.Societe, error)
	SaveSite(ctx context.Context, s domain.Site) error
	SaveService(ctx context.Context, s domain.Service) error
	SaveApplication(ctx context.Context, a domain.Application) error
	SaveUser(ctx context.Context, u domain.User) error
	FindUserByLogin(ctx context.Context, tenant kernel.TenantID, login string) (domain.User, error)
	FindUserByLoginGlobal(ctx context.Context, login string) (domain.User, error)
	ExistsLogin(ctx context.Context, tenant kernel.TenantID, login string) (bool, error)
	CountActiveUsers(ctx context.Context, tenant kernel.TenantID) (int, error)
	SaveClient(ctx context.Context, c domain.Client) error
	ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error)
	GetPermissions(ctx context.Context) (map[string]map[authx.Module]map[authx.Action]bool, error)
	ResolveUserEmails(ctx context.Context, tenant kernel.TenantID, userIDs []uuid.UUID) ([]string, error)
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hash, plain string) bool
}

type TokenIssuer interface {
	Issue(identity authx.Identity) (authx.TokenPair, error)
}

type EntitlementReader interface {
	IsModuleEnabled(ctx context.Context, tenantID kernel.TenantID, module authx.Module) (bool, error)
	GetSeatLimit(ctx context.Context, tenantID kernel.TenantID) (int, error)
}

type OrganizationService interface {
	CreateSociete(ctx context.Context, cmd CreateSocieteCommand) (domain.Societe, error)
	CreateSite(ctx context.Context, cmd CreateSiteCommand) (domain.Site, error)
	CreateService(ctx context.Context, cmd CreateServiceCommand) (domain.Service, error)
	CreateApplication(ctx context.Context, cmd CreateApplicationCommand) (domain.Application, error)
	ListSocietes(ctx context.Context, tenant kernel.TenantID) ([]domain.Societe, error)
}

type UserService interface {
	CreateUser(ctx context.Context, cmd CreateUserCommand) (domain.User, error)
	Authenticate(ctx context.Context, login, password string) (AuthResult, error)
}

type ClientService interface {
	CreateClient(ctx context.Context, cmd CreateClientCommand) (domain.Client, error)
	ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error)
}

type Clock interface {
	Now() time.Time
}
