package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/adapters/oidc"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
)

type oidcStatePayload struct {
	TenantID      string `json:"tenantId"`
	RedirectURI   string `json:"redirectUri"`
	CodeChallenge string `json:"codeChallenge"`
	Nonce         string `json:"nonce"`
}

type tokenGateway interface {
	ExchangeCode(ctx context.Context, issuer, clientID, clientSecret, redirectURI, code, codeVerifier string) (oidc.TokenResponse, error)
	ValidateIDToken(ctx context.Context, idToken, issuer, jwksURI, clientID string) (oidc.IDTokenClaims, error)
}

type oidcService struct {
	repo        ports.OrganizationRepository
	tokens      ports.TokenIssuer
	entitlement ports.EntitlementReader
	hasher      ports.PasswordHasher
	gateway     tokenGateway
	cache       cache.Cache
	keys        cache.KeyBuilder
	clock       func() time.Time
}

func NewOIDCService(
	repo ports.OrganizationRepository,
	tokens ports.TokenIssuer,
	entitlement ports.EntitlementReader,
	hasher ports.PasswordHasher,
	appCache cache.Cache,
	keyBuilder cache.KeyBuilder,
) ports.OIDCService {
	return &oidcService{
		repo:        repo,
		tokens:      tokens,
		entitlement: entitlement,
		hasher:      hasher,
		gateway:     oidc.NewGateway(),
		cache:       appCache,
		keys:        keyBuilder,
		clock:       time.Now,
	}
}

func (s *oidcService) AuthorizeURL(ctx context.Context, cmd ports.OIDCAuthorizeCommand) (string, error) {
	idp, err := s.repo.GetIdentityProvider(ctx, cmd.TenantID)
	if err != nil || !idp.Enabled {
		return "", domain.ErrSSONotEnabled
	}
	state := cmd.State
	if state == "" {
		state, err = oidc.GenerateState()
		if err != nil {
			return "", err
		}
	}
	nonce, err := oidc.GenerateState()
	if err != nil {
		return "", err
	}
	payload := oidcStatePayload{
		TenantID:      cmd.TenantID.String(),
		RedirectURI:   cmd.RedirectURI,
		CodeChallenge: cmd.CodeChallenge,
		Nonce:         nonce,
	}
	stateKey := s.keys.Key(cmd.TenantID, "org", "oidc-state", state)
	if err := s.cache.Set(ctx, stateKey, payload, 10*time.Minute); err != nil {
		return "", err
	}
	scopes := idp.Scopes
	if scopes == "" {
		scopes = "openid profile email"
	}
	return oidc.BuildAuthorizeURL(idp.Issuer, idp.ClientID, cmd.RedirectURI, scopes, state, cmd.CodeChallenge)
}

func (s *oidcService) Status(ctx context.Context, tenant kernel.TenantID) (ports.OIDCStatus, error) {
	idp, err := s.repo.GetIdentityProvider(ctx, tenant)
	if err != nil || !idp.Enabled {
		return ports.OIDCStatus{Enabled: false}, nil
	}
	return ports.OIDCStatus{Enabled: true, ProviderName: idp.Name}, nil
}

func (s *oidcService) HandleCallback(ctx context.Context, cmd ports.OIDCCallbackCommand) (ports.AuthResult, error) {
	stateKey := s.keys.Key(cmd.TenantID, "org", "oidc-state", cmd.State)
	var stored oidcStatePayload
	found, err := s.cache.Get(ctx, stateKey, &stored)
	if err != nil || !found {
		return ports.AuthResult{}, domain.ErrOIDCStateInvalid
	}
	_ = s.cache.Delete(ctx, stateKey)

	idp, err := s.repo.GetIdentityProvider(ctx, cmd.TenantID)
	if err != nil || !idp.Enabled {
		return ports.AuthResult{}, domain.ErrSSONotEnabled
	}

	tokenResp, err := s.gateway.ExchangeCode(ctx, idp.Issuer, idp.ClientID, idp.ClientSecret, cmd.RedirectURI, cmd.Code, cmd.CodeVerifier)
	if err != nil {
		return ports.AuthResult{}, fmt.Errorf("%w: %v", domain.ErrInvalidIDPToken, err)
	}

	var claims oidc.IDTokenClaims
	if idp.JWKSURI != "" {
		claims, err = s.gateway.ValidateIDToken(ctx, tokenResp.IDToken, idp.Issuer, idp.JWKSURI, idp.ClientID)
	} else {
		claims, err = oidc.ParseIDTokenPayload(tokenResp.IDToken)
	}
	if err != nil || claims.Subject == "" {
		if err != nil {
			return ports.AuthResult{}, fmt.Errorf("%w: %v", domain.ErrInvalidIDPToken, err)
		}
		return ports.AuthResult{}, domain.ErrInvalidIDPToken
	}

	user, err := s.resolveUser(ctx, cmd.TenantID, idp, claims.Subject, claims.Email)
	if err != nil {
		return ports.AuthResult{}, err
	}
	if !user.Active || !user.Period.IsActive(s.clock()) {
		return ports.AuthResult{}, domain.ErrAccountExpired
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

func (s *oidcService) resolveUser(ctx context.Context, tenant kernel.TenantID, idp domain.IdentityProvider, subject, email string) (domain.User, error) {
	link, err := s.repo.FindUserIdentityBySubject(ctx, tenant, idp.ID, subject)
	if err == nil {
		return s.repo.FindUserByID(ctx, tenant, link.UserID)
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if email != "" {
		if existing, err := s.repo.FindUserByEmail(ctx, tenant, email); err == nil {
			if err := s.repo.LinkUserIdentity(ctx, domain.UserIdentityLink{
				ID:       uuid.New(),
				TenantID: tenant,
				UserID:   existing.ID,
				IdPID:    idp.ID,
				Subject:  subject,
				Email:    email,
			}); err != nil {
				return domain.User{}, err
			}
			return existing, nil
		}
	}

	limit, err := s.entitlement.GetSeatLimit(ctx, tenant)
	if err != nil {
		return domain.User{}, err
	}
	count, err := s.repo.CountActiveUsers(ctx, tenant)
	if err != nil {
		return domain.User{}, err
	}
	if limit > 0 && count >= limit {
		return domain.User{}, domain.ErrSeatLimitReached
	}

	login, err := s.generateJITLogin(ctx, tenant, email)
	if err != nil {
		return domain.User{}, err
	}
	hash, err := s.hasher.Hash(randomPassword())
	if err != nil {
		return domain.User{}, err
	}
	profile := idp.DefaultProfile
	if profile == "" {
		profile = domain.ProfileCollaborateur
	}
	user := domain.User{
		ID:           uuid.New(),
		TenantID:     tenant,
		Login:        login,
		Email:        email,
		PasswordHash: hash,
		Profile:      profile,
		Active:       true,
		Period: domain.ActivationPeriod{
			Activation: s.clock().UTC().Truncate(24 * time.Hour),
		},
	}
	if err := s.repo.SaveUser(ctx, user); err != nil {
		return domain.User{}, err
	}
	if err := s.repo.LinkUserIdentity(ctx, domain.UserIdentityLink{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   user.ID,
		IdPID:    idp.ID,
		Subject:  subject,
		Email:    email,
	}); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *oidcService) generateJITLogin(ctx context.Context, tenant kernel.TenantID, email string) (domain.Login, error) {
	prefix := "SSO"
	if email != "" {
		local := strings.Split(email, "@")[0]
		local = strings.ToLower(strings.NewReplacer(".", "_", "-", "_").Replace(local))
		if len(local) >= 3 {
			prefix = strings.ToUpper(local[:3])
		}
	}
	for i := 0; i < 100; i++ {
		suffix := email
		if suffix == "" {
			suffix = uuid.New().String()[:8]
		}
		suffix = strings.NewReplacer("@", "_", ".", "_").Replace(suffix)
		candidate := fmt.Sprintf("%s_%s", prefix, suffix)
		if len(candidate) > 40 {
			candidate = candidate[:40]
		}
		login, err := domain.NewLogin(candidate)
		if err != nil {
			continue
		}
		exists, err := s.repo.ExistsLogin(ctx, tenant, string(login))
		if err != nil {
			return "", err
		}
		if !exists {
			return login, nil
		}
		prefix = "SSO"
	}
	return "", domain.ErrLoginAlreadyExists
}

func randomPassword() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

type identityProviderService struct {
	repo ports.OrganizationRepository
}

func NewIdentityProviderService(repo ports.OrganizationRepository) ports.IdentityProviderService {
	return &identityProviderService{repo: repo}
}

func (s *identityProviderService) Configure(ctx context.Context, cmd ports.ConfigureIdPCommand) (domain.IdentityProvider, error) {
	existing, err := s.repo.GetIdentityProvider(ctx, cmd.TenantID)
	idp := domain.IdentityProvider{
		ID:             cmd.ID,
		TenantID:       cmd.TenantID,
		Name:           cmd.Name,
		Issuer:         strings.TrimRight(cmd.Issuer, "/"),
		ClientID:       cmd.ClientID,
		ClientSecret:   cmd.ClientSecret,
		JWKSURI:        cmd.JWKSURI,
		Scopes:         cmd.Scopes,
		DefaultProfile: cmd.DefaultProfile,
		Enabled:        cmd.Enabled,
	}
	if err == nil && existing.ID != uuid.Nil {
		idp.ID = existing.ID
		if cmd.ClientSecret == "" {
			idp.ClientSecret = existing.ClientSecret
		}
	} else if idp.ID == uuid.Nil {
		idp.ID = uuid.New()
	}
	if idp.Scopes == "" {
		idp.Scopes = "openid profile email"
	}
	if idp.DefaultProfile == "" {
		idp.DefaultProfile = domain.ProfileCollaborateur
	}
	return idp, s.repo.SaveIdentityProvider(ctx, idp)
}

func (s *identityProviderService) List(ctx context.Context, tenant kernel.TenantID) ([]domain.IdentityProvider, error) {
	return s.repo.ListIdentityProviders(ctx, tenant)
}
