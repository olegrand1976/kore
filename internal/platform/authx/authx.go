package authx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

type Profile string

const (
	ProfileAdmin        Profile = "Administrateur"
	ProfileUtilisateur  Profile = "Utilisateur"
	ProfileCollaborateur Profile = "Collaborateur"
)

type Module string

type Action string

const (
	ActionRead     Action = "L"
	ActionWrite    Action = "E"
	ActionValidate Action = "V"
)

type Identity struct {
	UserID   uuid.UUID
	TenantID kernel.TenantID
	Profile  Profile
	Roles    []string
	JTI      string
}

type contextKey struct{}

var ErrUnauthorized = errors.New("unauthorized")
var ErrForbidden = errors.New("forbidden")
var ErrPaymentRequired = errors.New("payment required")

type EntitlementReader interface {
	IsModuleEnabled(ctx context.Context, tenantID kernel.TenantID, module Module) (bool, error)
	GetSeatLimit(ctx context.Context, tenantID kernel.TenantID) (int, error)
}

type Authorizer interface {
	Can(ctx context.Context, module Module, action Action) bool
}

type TokenIssuer struct {
	signingKey []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenIssuer(signingKey string, accessTTL, refreshTTL time.Duration) *TokenIssuer {
	return &TokenIssuer{
		signingKey: []byte(signingKey),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessJTI    string
	RefreshJTI   string
}

func (t *TokenIssuer) Issue(identity Identity) (TokenPair, error) {
	accessJTI := uuid.NewString()
	refreshJTI := uuid.NewString()
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"sub":       identity.UserID.String(),
		"tenant_id": identity.TenantID.String(),
		"profile":   string(identity.Profile),
		"roles":     identity.Roles,
		"jti":       accessJTI,
		"typ":       "access",
		"iat":       now.Unix(),
		"exp":       now.Add(t.accessTTL).Unix(),
	}
	refreshClaims := jwt.MapClaims{
		"sub":       identity.UserID.String(),
		"tenant_id": identity.TenantID.String(),
		"jti":       refreshJTI,
		"typ":       "refresh",
		"iat":       now.Unix(),
		"exp":       now.Add(t.refreshTTL).Unix(),
	}

	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(t.signingKey)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(t.signingKey)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		AccessJTI:    accessJTI,
		RefreshJTI:   refreshJTI,
	}, nil
}

func (t *TokenIssuer) ParseAccessToken(token string) (Identity, error) {
	return t.parseToken(token, "access")
}

func (t *TokenIssuer) ParseRefreshToken(token string) (Identity, error) {
	return t.parseToken(token, "refresh")
}

func (t *TokenIssuer) parseToken(token, expectedType string) (Identity, error) {
	parsed, err := jwt.Parse(token, func(tok *jwt.Token) (any, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return t.signingKey, nil
	})
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrUnauthorized, err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return Identity{}, ErrUnauthorized
	}
	if typ, _ := claims["typ"].(string); typ != expectedType {
		return Identity{}, ErrUnauthorized
	}
	userID, err := uuid.Parse(fmt.Sprint(claims["sub"]))
	if err != nil {
		return Identity{}, ErrUnauthorized
	}
	tenantID, err := kernel.ParseTenantID(fmt.Sprint(claims["tenant_id"]))
	if err != nil {
		return Identity{}, ErrUnauthorized
	}
	profile := Profile(fmt.Sprint(claims["profile"]))
	var roles []string
	if raw, ok := claims["roles"].([]any); ok {
		for _, r := range raw {
			roles = append(roles, fmt.Sprint(r))
		}
	}
	jti, _ := claims["jti"].(string)
	return Identity{
		UserID:   userID,
		TenantID: tenantID,
		Profile:  profile,
		Roles:    roles,
		JTI:      jti,
	}, nil
}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, identity)
}

func FromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(contextKey{}).(Identity)
	return identity, ok
}

func MustFromContext(ctx context.Context) Identity {
	identity, ok := FromContext(ctx)
	if !ok {
		panic("identity missing from context")
	}
	return identity
}

func BearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

type RBACAuthorizer struct {
	permissions map[string]map[Module]map[Action]bool
}

func NewRBACAuthorizer(permissions map[string]map[Module]map[Action]bool) *RBACAuthorizer {
	return &RBACAuthorizer{permissions: permissions}
}

func (a *RBACAuthorizer) Can(ctx context.Context, module Module, action Action) bool {
	identity, ok := FromContext(ctx)
	if !ok {
		return false
	}
	modPerms, ok := a.permissions[string(identity.Profile)]
	if !ok {
		return false
	}
	actions, ok := modPerms[module]
	if !ok {
		return false
	}
	return actions[action]
}
