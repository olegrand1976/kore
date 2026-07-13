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
	AccessTokenKindInvite    AccessTokenKind = "invite"
	AccessTokenKindDiscovery AccessTokenKind = "discovery"
)

type TenantAccessService struct {
	repo   ports.TenantAccessRepository
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
		if row.TokenHash != "" {
			if row.UsedAt != nil {
				return ports.TenantAccessResolveResult{}, domain.ErrAccessTokenUsed
			}
			if !row.ExpiresAt.After(now) {
				return ports.TenantAccessResolveResult{}, domain.ErrAccessTokenExpired
			}
		}
		return ports.TenantAccessResolveResult{}, domain.ErrAccessTokenInvalid
	}
	return ports.TenantAccessResolveResult{TenantID: row.TenantID, Kind: row.Kind}, nil
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
