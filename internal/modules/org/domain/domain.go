package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrInvalidLogin              = errors.New("invalid login format")
	ErrLoginAlreadyExists        = errors.New("login already exists")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrAccountExpired            = errors.New("account expired")
	ErrServiceWithoutResponsible = errors.New("service without responsible")
	ErrSeatLimitReached          = errors.New("seat limit reached")
	ErrUserNotFound              = errors.New("user not found")
	ErrCannotModifySelf          = errors.New("cannot modify own account")
	ErrInvalidGeminiModel        = errors.New("invalid gemini model")
	ErrSSONotEnabled             = errors.New("sso not enabled")
	ErrInvalidIDPToken           = errors.New("invalid idp token")
	ErrIdentityAlreadyLinked     = errors.New("identity already linked")
	ErrOIDCStateInvalid          = errors.New("invalid oidc state")
	ErrAccessTokenInvalid        = errors.New("invalid access token")
	ErrAccessTokenExpired        = errors.New("expired access token")
	ErrAccessTokenUsed           = errors.New("used access token")
	ErrInvalidEmail              = errors.New("invalid email")
	Err2FANotEnabled             = errors.New("2fa not enabled")
	Err2FAAlreadyEnabled         = errors.New("2fa already enabled")
	Err2FAInvalidCode            = errors.New("invalid 2fa code")
	Err2FAChallengeExpired       = errors.New("2fa challenge expired")
	Err2FAPasswordRequired       = errors.New("password required")
	Err2FAPolicyForbidden        = errors.New("2fa policy forbids this action")
	Err2FAEnrollmentRequired     = errors.New("2fa enrollment required")
	Err2FAEnrollmentTokenInvalid = errors.New("invalid 2fa enrollment token")
	Err2FARateLimited            = errors.New("too many 2fa attempts")
)

var loginPattern = regexp.MustCompile(`^[A-Z]{3}_[a-z0-9_]+$`)

type Login string

func NewLogin(value string) (Login, error) {
	if !loginPattern.MatchString(value) {
		return "", ErrInvalidLogin
	}
	return Login(value), nil
}

type Profile string

const ProfileAdmin Profile = "Administrateur"
const ProfileCollaborateur Profile = "Collaborateur"

type ActivationPeriod struct {
	Activation time.Time
	Expiration *time.Time
}

func (a ActivationPeriod) IsActive(now time.Time) bool {
	if now.Before(a.Activation) {
		return false
	}
	if a.Expiration != nil && now.After(*a.Expiration) {
		return false
	}
	return true
}

type User struct {
	ID                    uuid.UUID
	TenantID              kernel.TenantID
	EquipeID              *uuid.UUID
	Login                 Login
	Prenom                string
	Nom                   string
	Email                 string
	PasswordHash          string
	Profile               Profile
	Active                bool
	Period                ActivationPeriod
	DeletedAt             *time.Time
	TotpEnabled           bool
	TotpEnrollmentRequired bool
	TotpSecretEncrypted   string
	TotpEnabledAt         *time.Time
}

type IdentityProvider struct {
	ID             uuid.UUID
	TenantID       kernel.TenantID
	Name           string
	Issuer         string
	ClientID       string
	ClientSecret   string
	JWKSURI        string
	Scopes         string
	DefaultProfile Profile
	Enabled        bool
}

type UserIdentityLink struct {
	ID       uuid.UUID
	TenantID kernel.TenantID
	UserID   uuid.UUID
	IdPID    uuid.UUID
	Subject  string
	Email    string
}

type Societe struct {
	ID                 uuid.UUID       `json:"id"`
	TenantID           kernel.TenantID `json:"tenantId"`
	RaisonSociale      string          `json:"raisonSociale"`
	Logo               string          `json:"logo,omitempty"`
	Devise             string          `json:"devise"`
	Pays               string          `json:"pays"`
	WeekStartDay       int             `json:"weekStartDay"`
	DayCapacityMinutes int             `json:"dayCapacityMinutes"`
	CraMailAuto        bool            `json:"craMailAuto"`
	CraMailRecipients  []string        `json:"craMailRecipients,omitempty"`
	WeekSubmitPolicy      string   `json:"weekSubmitPolicy"`
	TaskTypesEnabled      []string `json:"taskTypesEnabled,omitempty"`
	TotpDefaultEnabled    bool     `json:"totpDefaultEnabled"`
	TotpUserConfigurable  bool   `json:"totpUserConfigurable"`
	Adresse               string `json:"adresse,omitempty"`
	Siret              string          `json:"siret,omitempty"`
	URLTenant          string          `json:"urlTenant,omitempty"`
}

const DefaultWeekStartDay = 1 // Monday (0=Sunday … 6=Saturday)
const DefaultDayCapacityMinutes = 480
const DefaultWeekSubmitPolicy = "warn"

var DefaultTaskTypesEnabled = []string{"manual", "interne", "formation", "mission"}

func EffectiveTaskTypesEnabled(types []string) []string {
	if len(types) == 0 {
		return append([]string(nil), DefaultTaskTypesEnabled...)
	}
	return types
}

type Site struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	SocieteID uuid.UUID
	Libelle   string
	Pays      string
}

type Service struct {
	ID            uuid.UUID
	TenantID      kernel.TenantID
	SiteID        uuid.UUID
	ResponsableID *uuid.UUID
}

type Application struct {
	ID              uuid.UUID       `json:"id"`
	TenantID        kernel.TenantID `json:"tenantId"`
	ServiceID       uuid.UUID       `json:"serviceId"`
	Libelle         string          `json:"libelle"`
	Proprietaire    string          `json:"proprietaire,omitempty"`
	ModeFacturation string          `json:"modeFacturation,omitempty"`
	UOActivee       bool            `json:"uoActivee"`
}

type ClientContact struct {
	Nom       string `json:"nom"`
	Prenom    string `json:"prenom"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Telephone string `json:"telephone"`
}

type Client struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      kernel.TenantID `json:"tenantId"`
	RaisonSociale string          `json:"raisonSociale"`
	TVA           string          `json:"tva"`
	Contacts      []ClientContact `json:"contacts"`
	Archived      bool            `json:"archived"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type Tenant struct {
	ID   uuid.UUID
	Name string
}
