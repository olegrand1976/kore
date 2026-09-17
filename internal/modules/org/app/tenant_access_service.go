package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type AccessTokenKind string

const (
	AccessTokenKindInvite        AccessTokenKind = "invite"
	AccessTokenKindDiscovery     AccessTokenKind = "discovery"
	AccessTokenKindPasswordReset AccessTokenKind = "password_reset"
)

const passwordResetTTL = time.Hour

// PasswordResetUserStore looks up and updates users during password reset confirm.
type PasswordResetUserStore interface {
	FindUserByEmail(ctx context.Context, tenant kernel.TenantID, email string) (domain.User, error)
	UpdateUser(ctx context.Context, u domain.User) error
}

type TenantAccessService struct {
	repo   ports.TenantAccessRepository
	users  PasswordResetUserStore
	hasher ports.PasswordHasher
	mailer ports.TransactionalEmailSender
	clock  func() time.Time
}

func NewTenantAccessService(repo ports.TenantAccessRepository, mailer ports.TransactionalEmailSender) *TenantAccessService {
	return &TenantAccessService{
		repo:   repo,
		mailer: mailer,
		clock:  time.Now,
	}
}

// WithPasswordReset enables forgot-password request/confirm (hasher + user store required).
func (s *TenantAccessService) WithPasswordReset(users PasswordResetUserStore, hasher ports.PasswordHasher) *TenantAccessService {
	s.users = users
	s.hasher = hasher
	return s
}

func (s *TenantAccessService) RequestTenantDiscovery(ctx context.Context, email string, baseLoginURL string) error {
	email = normalizeEmail(email)
	if email == "" {
		return nil
	}
	tenants, err := s.repo.FindTenantIDsByEmail(ctx, email)
	if err != nil || len(tenants) == 0 {
		// Avoid leaking whether the email exists.
		return nil
	}
	tenant := tenants[0]
	token, tokenHash, err := newToken()
	if err != nil {
		return err
	}
	expiresAt := s.clock().Add(24 * time.Hour)
	if err := s.repo.SaveAccessToken(ctx, tokenHash, tenant, email, string(AccessTokenKindDiscovery), expiresAt); err != nil {
		return err
	}
	link, err := withQuery(baseLoginURL, "discover", token)
	if err != nil {
		return err
	}
	subject := "Kore — Retrouver votre organisation"
	body := "Pour retrouver votre organisation, ouvrez ce lien :\n\n" + link + "\n\nCe lien expire sous 24h."
	return s.mailer.SendTenantAccessEmail(ctx, email, subject, body)
}

func (s *TenantAccessService) CreateInvitation(ctx context.Context, tenant kernel.TenantID, email string, baseLoginURL string) error {
	email = normalizeEmail(email)
	if email == "" {
		return domain.ErrInvalidEmail
	}
	token, tokenHash, err := newToken()
	if err != nil {
		return err
	}
	expiresAt := s.clock().Add(24 * time.Hour)
	if err := s.repo.SaveAccessToken(ctx, tokenHash, tenant, email, string(AccessTokenKindInvite), expiresAt); err != nil {
		return err
	}
	link, err := withQuery(baseLoginURL, "invite", token)
	if err != nil {
		return err
	}
	subject := "Kore — Invitation"
	body := "Vous avez été invité à rejoindre Kore.\n\nOuvrez ce lien pour vous connecter :\n\n" + link + "\n\nCe lien expire sous 24h."
	return s.mailer.SendTenantAccessEmail(ctx, email, subject, body)
}

func (s *TenantAccessService) Resolve(ctx context.Context, token string) (ports.TenantAccessResolveResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return ports.TenantAccessResolveResult{}, domain.ErrAccessTokenInvalid
	}
	tokenHash := hashToken(token)
	now := s.clock()
	row, ok, err := s.repo.ConsumeAccessToken(ctx, tokenHash, now)
	if err != nil {
		return ports.TenantAccessResolveResult{}, err
	}
	if !ok {
		return ports.TenantAccessResolveResult{}, accessTokenFailure(row, now)
	}
	return ports.TenantAccessResolveResult{TenantID: row.TenantID, Kind: row.Kind}, nil
}

// RequestPasswordReset emails a one-time reset link for each tenant matching the email.
// Always returns nil when the email is empty or unknown (anti-enumeration).
func (s *TenantAccessService) RequestPasswordReset(ctx context.Context, email string, baseResetURL string) error {
	email = normalizeEmail(email)
	if email == "" {
		return nil
	}
	tenants, err := s.repo.FindTenantIDsByEmail(ctx, email)
	if err != nil || len(tenants) == 0 {
		return nil
	}
	now := s.clock()
	for _, tenant := range tenants {
		if err := s.repo.InvalidateUnusedAccessTokens(ctx, tenant, email, string(AccessTokenKindPasswordReset), now); err != nil {
			return err
		}
		token, tokenHash, err := newToken()
		if err != nil {
			return err
		}
		expiresAt := now.Add(passwordResetTTL)
		if err := s.repo.SaveAccessToken(ctx, tokenHash, tenant, email, string(AccessTokenKindPasswordReset), expiresAt); err != nil {
			return err
		}
		link, err := withQuery(baseResetURL, "token", token)
		if err != nil {
			return err
		}
		subject := "Kore — Réinitialisation du mot de passe"
		body := "Pour choisir un nouveau mot de passe, ouvrez ce lien :\n\n" + link + "\n\nCe lien expire sous 1h."
		if err := s.mailer.SendTenantAccessEmail(ctx, email, subject, body); err != nil {
			return err
		}
	}
	return nil
}

// ConfirmPasswordReset validates a password_reset token, updates the password, then consumes the token.
// Peek-then-update-then-consume avoids burning invite/discovery tokens and keeps the link reusable if update fails.
func (s *TenantAccessService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if s.users == nil || s.hasher == nil {
		return errors.New("password reset is not configured")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.ErrAccessTokenInvalid
	}
	if err := domain.ValidatePassword(newPassword); err != nil {
		return err
	}
	tokenHash := hashToken(token)
	now := s.clock()
	row, found, err := s.repo.FindAccessToken(ctx, tokenHash)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrAccessTokenInvalid
	}
	if row.Kind != string(AccessTokenKindPasswordReset) {
		return domain.ErrAccessTokenInvalid
	}
	if row.UsedAt != nil {
		return domain.ErrAccessTokenUsed
	}
	if !row.ExpiresAt.After(now) {
		return domain.ErrAccessTokenExpired
	}
	user, err := s.users.FindUserByEmail(ctx, row.TenantID, row.Email)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if !user.Active || !user.Period.IsActive(now) {
		return domain.ErrAccountExpired
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	if err := s.users.UpdateUser(ctx, user); err != nil {
		return err
	}
	_, ok, err := s.repo.ConsumeAccessTokenOfKind(ctx, tokenHash, string(AccessTokenKindPasswordReset), now)
	if err != nil {
		return err
	}
	if !ok {
		// Password already updated; treat concurrent consume as success.
		return nil
	}
	return nil
}

func accessTokenFailure(row ports.AccessTokenRow, now time.Time) error {
	if row.TokenHash != "" {
		if row.UsedAt != nil {
			return domain.ErrAccessTokenUsed
		}
		if !row.ExpiresAt.After(now) {
			return domain.ErrAccessTokenExpired
		}
	}
	return domain.ErrAccessTokenInvalid
}

func normalizeEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !strings.Contains(email, "@") {
		return ""
	}
	return email
}

func newToken() (token string, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	tokenHash = hashToken(token)
	return token, tokenHash, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func withQuery(baseLoginURL string, key string, value string) (string, error) {
	u, err := url.Parse(baseLoginURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errors.New("base login url must be absolute")
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
