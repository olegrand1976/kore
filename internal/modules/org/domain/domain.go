package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrInvalidLogin                            = errors.New("invalid login format")
	ErrLoginAlreadyExists                      = errors.New("login already exists")
	ErrInvalidCredentials                      = errors.New("invalid credentials")
	ErrWeakPassword                            = errors.New("weak password")
	ErrAccountExpired                          = errors.New("account expired")
	ErrServiceWithoutResponsible               = errors.New("service without responsible")
	ErrEquipeWithoutApplication                = errors.New("equipe without application")
	ErrSeatLimitReached                        = errors.New("seat limit reached")
	ErrProfilesRequired                        = errors.New("at least one profile is required")
	ErrInvalidProfile                          = errors.New("invalid profile")
	ErrEquipeNotFound                          = errors.New("equipe not found")
	ErrSiteNotFound                            = errors.New("site not found")
	ErrServiceNotFound                         = errors.New("service not found")
	ErrInvalidSiteLibelle                      = errors.New("site libelle required")
	ErrInvalidServiceLibelle                   = errors.New("service libelle required")
	ErrInvalidEquipeLibelle                    = errors.New("equipe libelle required")
	ErrUserNotFound                            = errors.New("user not found")
	ErrSocieteNotFound                         = errors.New("societe not found")
	ErrClientNotFound                          = errors.New("client not found")
	ErrInvalidClientName                       = errors.New("client legal name is required")
	ErrLogoNotFound                            = errors.New("logo not found")
	ErrApplicationNotFound                     = errors.New("application not found")
	ErrApplicationsMergeInvalid                = errors.New("invalid application merge")
	ErrApplicationsMergeBothTaigaLinked        = errors.New("cannot merge two taiga-linked applications")
	ErrApplicationsMergeActiveSprintConflict   = errors.New("cannot merge applications with two active sprints")
	ErrApplicationsMergeMethodologyConflict    = errors.New("cannot merge applications with different methodology profiles when agile artifacts exist")
	ErrApplicationsMergeDuplicateDefaultBudget = errors.New("cannot merge applications with two default budgets")
	ErrApplicationWithoutShare                 = errors.New("application requires at least one site, service or equipe share")
	ErrInvalidApplicationShare                 = errors.New("invalid application share")
	ErrInvalidModeFacturation                  = errors.New("invalid mode facturation")
	ErrInvalidMethodologyProfile               = errors.New("invalid methodology profile")
	ErrMethodologyProfileLocked                = errors.New("methodology profile locked after agile artifacts exist")
	ErrInvalidApplicationLibelle               = errors.New("application libelle required")
	ErrBudgetNotFound                          = errors.New("budget not found for application")
	ErrBudgetNotAllowedOnCreate                = errors.New("budgetDefautId cannot be set on create")
	ErrCannotModifySelf                        = errors.New("cannot modify own account")
	ErrCannotDemoteSelf                        = errors.New("cannot remove own administrator profile")
	ErrLastAdmin                               = errors.New("cannot remove the last administrator")
	ErrInvalidGeminiModel                      = errors.New("invalid gemini model")
	ErrInvalidPays                             = errors.New("unsupported country code")
	ErrProvisionInputRequired                  = errors.New("tenant name and company name are required")
	ErrSSONotEnabled                           = errors.New("sso not enabled")
	ErrInvalidIDPToken                         = errors.New("invalid idp token")
	ErrIdentityAlreadyLinked                   = errors.New("identity already linked")
	ErrOIDCStateInvalid                        = errors.New("invalid oidc state")
	ErrAccessTokenInvalid                      = errors.New("invalid access token")
	ErrAccessTokenExpired                      = errors.New("expired access token")
	ErrAccessTokenUsed                         = errors.New("used access token")
	ErrInvalidEmail                            = errors.New("invalid email")
	Err2FANotEnabled                           = errors.New("2fa not enabled")
	Err2FAAlreadyEnabled                       = errors.New("2fa already enabled")
	Err2FAInvalidCode                          = errors.New("invalid 2fa code")
	Err2FAChallengeExpired                     = errors.New("2fa challenge expired")
	Err2FAPasswordRequired                     = errors.New("password required")
	Err2FAPolicyForbidden                      = errors.New("2fa policy forbids this action")
	Err2FAEnrollmentRequired                   = errors.New("2fa enrollment required")
	Err2FAEnrollmentTokenInvalid               = errors.New("invalid 2fa enrollment token")
	Err2FARateLimited                          = errors.New("too many 2fa attempts")
)

// loginPattern accepts an optional XXX_ profile/company prefix (legacy seed style)
// or a plain identifier: olivier, jean.dupont, COL_olivier, ADM_admin.
var loginPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]{2,63}$`)

type Login string

func NewLogin(value string) (Login, error) {
	if !loginPattern.MatchString(value) {
		return "", ErrInvalidLogin
	}
	return Login(value), nil
}

// ValidatePassword enforces: ≥8 chars, at least one lower, one upper, one digit.
func ValidatePassword(value string) error {
	if len(value) < 8 {
		return ErrWeakPassword
	}
	var hasLower, hasUpper, hasDigit bool
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLower || !hasUpper || !hasDigit {
		return ErrWeakPassword
	}
	return nil
}

type Profile string

const ProfileAdmin Profile = "Administrateur"
const ProfileCollaborateur Profile = "Collaborateur"
const ProfileChefEquipe Profile = "Chef d'équipe"
const ProfileResponsableService Profile = "Responsable de service"

// KnownProfiles is the closed set of RBAC profiles assignable to a user.
var KnownProfiles = []Profile{
	ProfileAdmin,
	ProfileCollaborateur,
	ProfileChefEquipe,
	ProfileResponsableService,
}

func ValidateProfiles(profiles []Profile) error {
	if len(profiles) == 0 {
		return ErrProfilesRequired
	}
	known := make(map[Profile]struct{}, len(KnownProfiles))
	for _, p := range KnownProfiles {
		known[p] = struct{}{}
	}
	for _, p := range profiles {
		if _, ok := known[p]; !ok {
			return ErrInvalidProfile
		}
	}
	return nil
}

// HasAdminProfile reports whether the user holds the Administrateur profile
// (membership list or legacy primary column).
func (u User) HasAdminProfile() bool {
	if ProfilesContain(u.Profiles, ProfileAdmin) {
		return true
	}
	return u.Profile == ProfileAdmin
}

func ProfilesContain(profiles []Profile, want Profile) bool {
	for _, p := range profiles {
		if p == want {
			return true
		}
	}
	return false
}

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
	ID                     uuid.UUID
	TenantID               kernel.TenantID
	EquipeID               *uuid.UUID
	EquipeIDs              []uuid.UUID
	Login                  Login
	Prenom                 string
	Nom                    string
	Email                  string
	PasswordHash           string
	Profile                Profile
	Profiles               []Profile
	Active                 bool
	Period                 ActivationPeriod
	DeletedAt              *time.Time
	TotpEnabled            bool
	TotpEnrollmentRequired bool
	TotpSecretEncrypted    string
	TotpEnabledAt          *time.Time
}

// SyncPrimaryMemberships keeps denormalized profil / equipe_id in sync with the
// multi-membership slices (JWT primary claim + legacy joins).
// For teams: keeps the current EquipeID when it remains a member; otherwise
// picks EquipeIDs[0]. Never reorders membership just because of UUID sort order.
func (u *User) SyncPrimaryMemberships() {
	if len(u.Profiles) == 0 && u.Profile != "" {
		u.Profiles = []Profile{u.Profile}
	}
	if len(u.Profiles) > 0 {
		u.Profile = PrimaryProfile(u.Profiles)
	} else if u.Profile == "" {
		u.Profile = ProfileCollaborateur
		u.Profiles = []Profile{ProfileCollaborateur}
	}
	switch {
	case len(u.EquipeIDs) > 0:
		if u.EquipeID != nil {
			for _, id := range u.EquipeIDs {
				if id == *u.EquipeID {
					return
				}
			}
		}
		id := u.EquipeIDs[0]
		u.EquipeID = &id
	case u.EquipeIDs != nil:
		// Non-nil empty slice = explicit detach.
		u.EquipeID = nil
	case u.EquipeID != nil:
		// Nil EquipeIDs + primary set = hydrate from legacy column.
		u.EquipeIDs = []uuid.UUID{*u.EquipeID}
	}
}

// PrimaryProfile prefers Administrateur, then manager roles, else first entry.
func PrimaryProfile(profiles []Profile) Profile {
	if len(profiles) == 0 {
		return ProfileCollaborateur
	}
	priority := []Profile{
		ProfileAdmin,
		ProfileResponsableService,
		ProfileChefEquipe,
		ProfileCollaborateur,
	}
	set := make(map[Profile]struct{}, len(profiles))
	for _, p := range profiles {
		set[p] = struct{}{}
	}
	for _, p := range priority {
		if _, ok := set[p]; ok {
			return p
		}
	}
	return profiles[0]
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
	ID                   uuid.UUID       `json:"id"`
	TenantID             kernel.TenantID `json:"tenantId"`
	RaisonSociale        string          `json:"raisonSociale"`
	Logo                 string          `json:"logo,omitempty"`
	Devise               string          `json:"devise"`
	Pays                 string          `json:"pays"`
	WeekStartDay         int             `json:"weekStartDay"`
	DayCapacityMinutes   int             `json:"dayCapacityMinutes"`
	CraMailAuto          bool            `json:"craMailAuto"`
	CraMailRecipients    []string        `json:"craMailRecipients,omitempty"`
	WeekSubmitPolicy     string          `json:"weekSubmitPolicy"`
	CraGateMode          string          `json:"craGateMode"`
	TaskTypesEnabled     []string        `json:"taskTypesEnabled,omitempty"`
	TotpDefaultEnabled   bool            `json:"totpDefaultEnabled"`
	TotpUserConfigurable bool            `json:"totpUserConfigurable"`
	Adresse              string          `json:"adresse,omitempty"`
	AdresseNumero        string          `json:"adresseNumero,omitempty"`
	AdresseBoite         string          `json:"adresseBoite,omitempty"`
	CodePostal           string          `json:"codePostal,omitempty"`
	Ville                string          `json:"ville,omitempty"`
	Siret                string          `json:"siret,omitempty"`
	URLTenant            string          `json:"urlTenant,omitempty"`
	SeedProtected        bool            `json:"seedProtected"`
	DefaultTJMCents      int64           `json:"defaultTjmCents"`
}

// NormalizeSocietePays returns the canonical ISO country code for a société.
// ok is false when the value is non-empty but not in the supported whitelist.
func NormalizeSocietePays(pays string) (normalized string, ok bool) {
	switch strings.ToUpper(strings.TrimSpace(pays)) {
	case "BE":
		return "BE", true
	case "FR", "":
		return "FR", true
	case "MG", "MD": // MD was a mistaken non-ISO alias for Madagascar
		return "MG", true
	case "MA":
		return "MA", true
	case "TN":
		return "TN", true
	case "CA":
		return "CA", true
	default:
		return "", false
	}
}

// FormatSocieteAddress builds a single-line postal address for display/PDF.
func FormatSocieteAddress(s Societe) string {
	return formatPostalAddress(s.Adresse, s.AdresseNumero, s.AdresseBoite, s.CodePostal, s.Ville, s.Pays)
}

func formatPostalAddress(street, numero, boite, cp, ville, paysRaw string) string {
	street = strings.TrimSpace(street)
	numero = strings.TrimSpace(numero)
	boite = strings.TrimSpace(boite)
	cp = strings.TrimSpace(cp)
	ville = strings.TrimSpace(ville)
	pays := strings.ToUpper(strings.TrimSpace(paysRaw))
	switch pays {
	case "FR":
		pays = "France"
	case "BE":
		pays = "Belgique"
	case "MG":
		pays = "Madagascar"
	case "MA":
		pays = "Maroc"
	case "TN":
		pays = "Tunisie"
	case "CA":
		pays = "Canada"
	}

	line1 := street
	if numero != "" {
		if line1 != "" {
			line1 += " " + numero
		} else {
			line1 = numero
		}
	}
	if boite != "" {
		if line1 != "" {
			line1 += " / " + boite
		} else {
			line1 = boite
		}
	}

	line2 := strings.TrimSpace(cp + " " + ville)
	parts := make([]string, 0, 3)
	if line1 != "" {
		parts = append(parts, line1)
	}
	if line2 != "" {
		parts = append(parts, line2)
	}
	if pays != "" {
		parts = append(parts, pays)
	}
	return strings.Join(parts, ", ")
}

const DefaultWeekStartDay = 1 // Monday (0=Sunday … 6=Saturday)
const DefaultDayCapacityMinutes = 480
const DefaultWeekSubmitPolicy = "warn"
const DefaultCraGateMode = "warn"

var DefaultTaskTypesEnabled = []string{"manual", "interne", "formation", "mission"}

func EffectiveTaskTypesEnabled(types []string) []string {
	if len(types) == 0 {
		return append([]string(nil), DefaultTaskTypesEnabled...)
	}
	return types
}

// DefaultServiceType est le type appliqué à un service créé sans type explicite.
const DefaultServiceType = "interne"

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
	Libelle       string
	Type          string
	ResponsableID *uuid.UUID
}

type Equipe struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      kernel.TenantID `json:"tenantId"`
	ApplicationID uuid.UUID       `json:"applicationId"`
	Libelle       string          `json:"libelle"`
	ResponsableID *uuid.UUID      `json:"responsableId"`
}

type SiteSummary struct {
	ID        uuid.UUID `json:"id"`
	SocieteID uuid.UUID `json:"societeId"`
	Libelle   string    `json:"libelle"`
	Pays      string    `json:"pays,omitempty"`
}

type ServiceSummary struct {
	ID            uuid.UUID  `json:"id"`
	SiteID        uuid.UUID  `json:"siteId"`
	SiteLabel     string     `json:"siteLabel,omitempty"`
	SocieteID     uuid.UUID  `json:"societeId,omitempty"`
	Libelle       string     `json:"libelle,omitempty"`
	Type          string     `json:"type,omitempty"`
	ResponsableID *uuid.UUID `json:"responsableId,omitempty"`
}

// Mode facturation values for an Application (spec §4.3 Non / Forfait / Réel).
const (
	ModeFacturationNon        = "non"
	ModeFacturationForfait    = "forfait"
	ModeFacturationTempsPasse = "temps_passe" // « Réel »
	DefaultModeFacturation    = ModeFacturationTempsPasse
)

type MethodologyProfile string

const (
	MethodologyPSA         MethodologyProfile = "psa"
	MethodologyAgileScrum  MethodologyProfile = "agile_scrum"
	MethodologyAgileKanban MethodologyProfile = "agile_kanban"
)

func DefaultMethodologyProfile() MethodologyProfile {
	return MethodologyPSA
}

func NormalizeMethodologyProfile(raw string) (MethodologyProfile, error) {
	if raw == "" {
		return DefaultMethodologyProfile(), nil
	}
	switch MethodologyProfile(raw) {
	case MethodologyPSA, MethodologyAgileScrum, MethodologyAgileKanban:
		return MethodologyProfile(raw), nil
	default:
		return "", ErrInvalidMethodologyProfile
	}
}

func (p MethodologyProfile) IsAgile() bool {
	return p == MethodologyAgileScrum || p == MethodologyAgileKanban
}

func NormalizeModeFacturation(raw string) (string, error) {
	if raw == "" {
		return DefaultModeFacturation, nil
	}
	switch raw {
	case ModeFacturationNon, ModeFacturationForfait, ModeFacturationTempsPasse:
		return raw, nil
	default:
		return "", ErrInvalidModeFacturation
	}
}

type Application struct {
	ID                 uuid.UUID          `json:"id"`
	TenantID           kernel.TenantID    `json:"tenantId"`
	Libelle            string             `json:"libelle"`
	Proprietaire       string             `json:"proprietaire,omitempty"`
	ModeFacturation    string             `json:"modeFacturation,omitempty"`
	MethodologyProfile MethodologyProfile `json:"methodologyProfile"`
	UOActivee          bool               `json:"uoActivee"`
	ChefUtilisateurID  *uuid.UUID         `json:"chefUtilisateurId,omitempty"`
	BudgetDefautID     *uuid.UUID         `json:"budgetDefautId,omitempty"`
	Active             bool               `json:"active"`
	DefaultTJMCents    int64              `json:"defaultTjmCents"`
	SiteIDs            []uuid.UUID        `json:"siteIds,omitempty"`
	ServiceIDs         []uuid.UUID        `json:"serviceIds,omitempty"`
	EquipeIDs          []uuid.UUID        `json:"equipeIds,omitempty"`
}

// HasShares reports whether the application has at least one org attachment.
func (a Application) HasShares() bool {
	return len(a.SiteIDs) > 0 || len(a.ServiceIDs) > 0 || len(a.EquipeIDs) > 0
}

// DedupeUUIDs returns unique non-nil UUIDs preserving first-seen order.
func DedupeUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
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

type ClientContact struct {
	ID        uuid.UUID `json:"id"`
	Nom       string    `json:"nom"`
	Prenom    string    `json:"prenom"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Telephone string    `json:"telephone"`
}

// NormalizeClientContacts assigns missing IDs and trims fields.
func NormalizeClientContacts(contacts []ClientContact) []ClientContact {
	out, _ := EnsureClientContactIDs(contacts)
	return out
}

// SanitizeClientContacts trims fields without generating IDs (read path).
func SanitizeClientContacts(contacts []ClientContact) []ClientContact {
	if len(contacts) == 0 {
		return []ClientContact{}
	}
	out := make([]ClientContact, 0, len(contacts))
	for _, c := range contacts {
		out = append(out, ClientContact{
			ID:        c.ID,
			Nom:       strings.TrimSpace(c.Nom),
			Prenom:    strings.TrimSpace(c.Prenom),
			Email:     strings.TrimSpace(c.Email),
			Role:      strings.TrimSpace(c.Role),
			Telephone: strings.TrimSpace(c.Telephone),
		})
	}
	return out
}

// EnsureClientContactIDs normalizes contacts and reports whether any ID was generated.
func EnsureClientContactIDs(contacts []ClientContact) ([]ClientContact, bool) {
	if len(contacts) == 0 {
		return []ClientContact{}, false
	}
	out := make([]ClientContact, 0, len(contacts))
	changed := false
	for _, c := range contacts {
		id := c.ID
		if id == uuid.Nil {
			id = uuid.New()
			changed = true
		}
		out = append(out, ClientContact{
			ID:        id,
			Nom:       strings.TrimSpace(c.Nom),
			Prenom:    strings.TrimSpace(c.Prenom),
			Email:     strings.TrimSpace(c.Email),
			Role:      strings.TrimSpace(c.Role),
			Telephone: strings.TrimSpace(c.Telephone),
		})
	}
	return out, changed
}

// ClientContactDisplayName returns "Prenom Nom" or a fallback.
func ClientContactDisplayName(c ClientContact) string {
	name := strings.TrimSpace(strings.TrimSpace(c.Prenom) + " " + strings.TrimSpace(c.Nom))
	if name != "" {
		return name
	}
	if email := strings.TrimSpace(c.Email); email != "" {
		return email
	}
	return c.ID.String()
}

type Client struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      kernel.TenantID `json:"tenantId"`
	RaisonSociale string          `json:"raisonSociale"`
	TVA           string          `json:"tva"`
	Pays          string          `json:"pays,omitempty"`
	Adresse       string          `json:"adresse,omitempty"`
	AdresseNumero string          `json:"adresseNumero,omitempty"`
	AdresseBoite  string          `json:"adresseBoite,omitempty"`
	CodePostal    string          `json:"codePostal,omitempty"`
	Ville         string          `json:"ville,omitempty"`
	Siret         string          `json:"siret,omitempty"`
	Contacts      []ClientContact `json:"contacts"`
	Archived      bool            `json:"archived"`
	CreatedAt     time.Time       `json:"createdAt"`
}

// FormatClientAddress builds a single-line postal address for display/PDF.
func FormatClientAddress(c Client) string {
	return formatPostalAddress(c.Adresse, c.AdresseNumero, c.AdresseBoite, c.CodePostal, c.Ville, c.Pays)
}

type Tenant struct {
	ID   uuid.UUID
	Name string
}
