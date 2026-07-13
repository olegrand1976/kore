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
	ID           uuid.UUID
	TenantID     kernel.TenantID
	EquipeID     *uuid.UUID
	Login        Login
	PasswordHash string
	Profile      Profile
	Active       bool
	Period       ActivationPeriod
	DeletedAt    *time.Time
}

type Societe struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      kernel.TenantID `json:"tenantId"`
	RaisonSociale string          `json:"raisonSociale"`
	Logo          string          `json:"logo,omitempty"`
	Devise        string          `json:"devise"`
	Pays          string          `json:"pays"`
	Adresse       string          `json:"adresse,omitempty"`
	Siret         string          `json:"siret,omitempty"`
	URLTenant     string          `json:"urlTenant,omitempty"`
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
	ID        uuid.UUID
	TenantID  kernel.TenantID
	ServiceID uuid.UUID
	Libelle   string
}

type Client struct {
	ID            uuid.UUID
	TenantID      kernel.TenantID
	RaisonSociale string
	TVA           string
	Archived      bool
}

type Tenant struct {
	ID   uuid.UUID
	Name string
}
