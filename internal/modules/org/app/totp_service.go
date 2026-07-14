package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/cryptox"
	"github.com/kore/kore/pkg/kernel"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
)

const (
	totpChallengeTTL    = 5 * time.Minute
	totpEnrollmentTTL   = 30 * time.Minute
	totpSetupPendingTTL = 10 * time.Minute
	totpBackupCodeCount = 8
	totpIssuer          = "Kore"
	totpRateLimitMax    = 5
	totpRateLimitWindow = 15 * time.Minute
)

type totpChallengePayload struct {
	UserID   uuid.UUID       `json:"userId"`
	TenantID kernel.TenantID `json:"tenantId"`
}

type totpEnrollmentPayload struct {
	UserID   uuid.UUID       `json:"userId"`
	TenantID kernel.TenantID `json:"tenantId"`
}

type totpSetupPendingPayload struct {
	SecretEncrypted string `json:"secretEncrypted"`
}

type totpRateLimitPayload struct {
	Count int `json:"count"`
}

func (s *userService) resolveTotpPolicy(ctx context.Context, user domain.User) (ports.TotpPolicy, error) {
	societeID, err := s.repo.ResolveSocieteIDForUser(ctx, user.TenantID, user.ID)
	if err != nil {
		return ports.TotpPolicy{}, err
	}
	societe, err := s.repo.GetSociete(ctx, user.TenantID, societeID)
	if err != nil {
		return ports.TotpPolicy{}, err
	}
	return ports.TotpPolicy{
		DefaultEnabled:   societe.TotpDefaultEnabled,
		UserConfigurable: societe.TotpUserConfigurable,
	}, nil
}

func (s *userService) resolveTotpPolicyOrDefault(ctx context.Context, user domain.User) ports.TotpPolicy {
	policy, err := s.resolveTotpPolicy(ctx, user)
	if err != nil {
		return ports.TotpPolicy{UserConfigurable: true}
	}
	return policy
}

func (s *userService) checkRateLimit(ctx context.Context, scope, id string) error {
	if id == "" {
		return domain.Err2FAInvalidCode
	}
	key := s.keys.PublicKey("org", "2fa-ratelimit", scope, id)
	var payload totpRateLimitPayload
	found, err := s.cache.Get(ctx, key, &payload)
	if err != nil {
		return err
	}
	if found && payload.Count >= totpRateLimitMax {
		return domain.Err2FARateLimited
	}
	payload.Count++
	return s.cache.Set(ctx, key, payload, totpRateLimitWindow)
}

func totpQRDataURL(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 200)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

func (s *userService) mustEnrollment(user domain.User, policy ports.TotpPolicy) bool {
	if user.TotpEnabled {
		return false
	}
	return user.TotpEnrollmentRequired || policy.DefaultEnabled
}

func (s *userService) canSetupTotp(policy ports.TotpPolicy) bool {
	if policy.UserConfigurable {
		return true
	}
	return policy.DefaultEnabled
}

func (s *userService) canDisableTotp(policy ports.TotpPolicy) bool {
	return policy.UserConfigurable
}

func (s *userService) issueAuthResult(user domain.User) (ports.AuthResult, error) {
	pair, err := s.tokens.Issue(s.buildIdentity(user))
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

func (s *userService) Authenticate(ctx context.Context, login, password string) (ports.AuthResult, error) {
	user, err := s.repo.FindUserByLoginGlobal(ctx, login)
	if err != nil {
		return ports.AuthResult{}, domain.ErrInvalidCredentials
	}
	if !user.Active || !user.Period.IsActive(s.clock()) {
		return ports.AuthResult{}, domain.ErrAccountExpired
	}
	if user.PasswordHash == "" || !s.hasher.Verify(user.PasswordHash, password) {
		return ports.AuthResult{}, domain.ErrInvalidCredentials
	}

	policy := s.resolveTotpPolicyOrDefault(ctx, user)

	if user.TotpEnabled {
		token, err := s.createChallengeToken(ctx, user)
		if err != nil {
			return ports.AuthResult{}, err
		}
		return ports.AuthResult{
			UserID:         user.ID,
			TenantID:       user.TenantID,
			Profile:        user.Profile,
			Requires2FA:    true,
			ChallengeToken: token,
		}, nil
	}

	if s.mustEnrollment(user, policy) && s.canSetupTotp(policy) {
		token, err := s.createEnrollmentToken(ctx, user)
		if err != nil {
			return ports.AuthResult{}, err
		}
		return ports.AuthResult{
			UserID:                user.ID,
			TenantID:              user.TenantID,
			Profile:               user.Profile,
			Requires2FAEnrollment: true,
			EnrollmentToken:       token,
		}, nil
	}

	return s.issueAuthResult(user)
}

func (s *userService) createChallengeToken(ctx context.Context, user domain.User) (string, error) {
	token := uuid.NewString()
	key := s.keys.PublicKey("org", "2fa-challenge", token)
	payload := totpChallengePayload{UserID: user.ID, TenantID: user.TenantID}
	if err := s.cache.Set(ctx, key, payload, totpChallengeTTL); err != nil {
		return "", err
	}
	return token, nil
}

func (s *userService) createEnrollmentToken(ctx context.Context, user domain.User) (string, error) {
	token := uuid.NewString()
	key := s.keys.PublicKey("org", "2fa-enrollment", token)
	payload := totpEnrollmentPayload{UserID: user.ID, TenantID: user.TenantID}
	if err := s.cache.Set(ctx, key, payload, totpEnrollmentTTL); err != nil {
		return "", err
	}
	return token, nil
}

func (s *userService) loadChallengeToken(ctx context.Context, token string) (totpChallengePayload, error) {
	key := s.keys.PublicKey("org", "2fa-challenge", token)
	var payload totpChallengePayload
	found, err := s.cache.Get(ctx, key, &payload)
	if err != nil || !found {
		return totpChallengePayload{}, domain.Err2FAChallengeExpired
	}
	return payload, nil
}

func (s *userService) deleteChallengeToken(ctx context.Context, token string) {
	key := s.keys.PublicKey("org", "2fa-challenge", token)
	_ = s.cache.Delete(ctx, key)
}

func (s *userService) loadEnrollmentToken(ctx context.Context, token string) (totpEnrollmentPayload, error) {
	key := s.keys.PublicKey("org", "2fa-enrollment", token)
	var payload totpEnrollmentPayload
	found, err := s.cache.Get(ctx, key, &payload)
	if err != nil || !found {
		return totpEnrollmentPayload{}, domain.Err2FAEnrollmentTokenInvalid
	}
	return payload, nil
}

func (s *userService) deleteEnrollmentToken(ctx context.Context, token string) {
	key := s.keys.PublicKey("org", "2fa-enrollment", token)
	_ = s.cache.Delete(ctx, key)
}

func (s *userService) Verify2FAChallenge(ctx context.Context, challengeToken, code string) (ports.AuthResult, error) {
	if err := s.checkRateLimit(ctx, "challenge", challengeToken); err != nil {
		return ports.AuthResult{}, err
	}
	payload, err := s.loadChallengeToken(ctx, challengeToken)
	if err != nil {
		return ports.AuthResult{}, err
	}
	user, err := s.repo.FindUserByID(ctx, payload.TenantID, payload.UserID)
	if err != nil {
		return ports.AuthResult{}, domain.Err2FAChallengeExpired
	}
	if !user.TotpEnabled {
		return ports.AuthResult{}, domain.Err2FANotEnabled
	}
	ok, err := s.validateTotpCode(ctx, user, code)
	if err != nil {
		return ports.AuthResult{}, err
	}
	if !ok {
		return ports.AuthResult{}, domain.Err2FAInvalidCode
	}
	s.deleteChallengeToken(ctx, challengeToken)
	return s.issueAuthResult(user)
}

func (s *userService) Verify2FAEnrollment(ctx context.Context, enrollmentToken, code, password string) (ports.AuthResult, error) {
	if err := s.checkRateLimit(ctx, "enrollment", enrollmentToken); err != nil {
		return ports.AuthResult{}, err
	}
	payload, err := s.loadEnrollmentToken(ctx, enrollmentToken)
	if err != nil {
		return ports.AuthResult{}, err
	}
	confirmResult, err := s.Confirm2FA(ctx, ports.Confirm2FACommand{
		TenantID:        payload.TenantID,
		UserID:          payload.UserID,
		Code:            code,
		Password:        password,
		EnrollmentToken: enrollmentToken,
		SkipPassword:    true,
	})
	if err != nil {
		return ports.AuthResult{}, err
	}
	s.deleteEnrollmentToken(ctx, enrollmentToken)
	user, err := s.repo.FindUserByID(ctx, payload.TenantID, payload.UserID)
	if err != nil {
		return ports.AuthResult{}, err
	}
	authResult, err := s.issueAuthResult(user)
	if err != nil {
		return ports.AuthResult{}, err
	}
	authResult.BackupCodes = confirmResult.BackupCodes
	return authResult, nil
}

func (s *userService) Get2FAStatus(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.TotpStatus, error) {
	user, err := s.repo.FindUserByID(ctx, tenant, userID)
	if err != nil {
		return ports.TotpStatus{}, err
	}
	policy := s.resolveTotpPolicyOrDefault(ctx, user)
	return ports.TotpStatus{
		Enabled:            user.TotpEnabled,
		EnrollmentRequired: s.mustEnrollment(user, policy),
		UserConfigurable:   policy.UserConfigurable,
		OrgDefaultEnabled:  policy.DefaultEnabled,
		EnabledAt:          user.TotpEnabledAt,
		PasswordLogin:      user.PasswordHash != "",
	}, nil
}

func (s *userService) Setup2FA(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.TotpSetupResult, error) {
	user, err := s.repo.FindUserByID(ctx, tenant, userID)
	if err != nil {
		return ports.TotpSetupResult{}, err
	}
	return s.beginTotpSetup(ctx, user)
}

func (s *userService) Setup2FAWithEnrollmentToken(ctx context.Context, enrollmentToken string) (ports.TotpSetupResult, error) {
	if err := s.checkRateLimit(ctx, "enrollment-setup", enrollmentToken); err != nil {
		return ports.TotpSetupResult{}, err
	}
	payload, err := s.loadEnrollmentToken(ctx, enrollmentToken)
	if err != nil {
		return ports.TotpSetupResult{}, err
	}
	user, err := s.repo.FindUserByID(ctx, payload.TenantID, payload.UserID)
	if err != nil {
		return ports.TotpSetupResult{}, err
	}
	return s.beginTotpSetup(ctx, user)
}

func (s *userService) beginTotpSetup(ctx context.Context, user domain.User) (ports.TotpSetupResult, error) {
	if user.PasswordHash == "" {
		return ports.TotpSetupResult{}, domain.Err2FAPolicyForbidden
	}
	if user.TotpEnabled {
		return ports.TotpSetupResult{}, domain.Err2FAAlreadyEnabled
	}
	policy := s.resolveTotpPolicyOrDefault(ctx, user)
	if !s.canSetupTotp(policy) && !s.mustEnrollment(user, policy) {
		return ports.TotpSetupResult{}, domain.Err2FAPolicyForbidden
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: string(user.Login),
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return ports.TotpSetupResult{}, err
	}

	encrypted, err := cryptox.Encrypt(s.totpKey, []byte(key.Secret()))
	if err != nil {
		return ports.TotpSetupResult{}, err
	}

	pendingKey := s.keys.Key(user.TenantID, "org", "2fa-pending", user.ID.String())
	if err := s.cache.Set(ctx, pendingKey, totpSetupPendingPayload{SecretEncrypted: encrypted}, totpSetupPendingTTL); err != nil {
		return ports.TotpSetupResult{}, err
	}

	qrDataURL, err := totpQRDataURL(key.URL())
	if err != nil {
		return ports.TotpSetupResult{}, err
	}

	return ports.TotpSetupResult{
		OtpauthURL:    key.URL(),
		Secret:        key.Secret(),
		QrCodeDataURL: qrDataURL,
	}, nil
}

func (s *userService) Confirm2FA(ctx context.Context, cmd ports.Confirm2FACommand) (ports.TotpConfirmResult, error) {
	user, err := s.repo.FindUserByID(ctx, cmd.TenantID, cmd.UserID)
	if err != nil {
		return ports.TotpConfirmResult{}, err
	}
	if user.TotpEnabled {
		return ports.TotpConfirmResult{}, domain.Err2FAAlreadyEnabled
	}
	if !cmd.SkipPassword {
		if cmd.Password == "" {
			return ports.TotpConfirmResult{}, domain.Err2FAPasswordRequired
		}
		if !s.hasher.Verify(user.PasswordHash, cmd.Password) {
			return ports.TotpConfirmResult{}, domain.ErrInvalidCredentials
		}
	}

	pendingKey := s.keys.Key(user.TenantID, "org", "2fa-pending", user.ID.String())
	var pending totpSetupPendingPayload
	found, err := s.cache.Get(ctx, pendingKey, &pending)
	if err != nil || !found {
		return ports.TotpConfirmResult{}, domain.Err2FAChallengeExpired
	}

	secret, err := s.decryptSecret(pending.SecretEncrypted)
	if err != nil {
		return ports.TotpConfirmResult{}, err
	}
	if !totp.Validate(cmd.Code, secret) {
		return ports.TotpConfirmResult{}, domain.Err2FAInvalidCode
	}

	now := s.clock().UTC()
	user.TotpEnabled = true
	user.TotpEnrollmentRequired = false
	user.TotpSecretEncrypted = pending.SecretEncrypted
	user.TotpEnabledAt = &now
	if err := s.repo.UpdateUserTotp(ctx, user); err != nil {
		return ports.TotpConfirmResult{}, err
	}
	_ = s.cache.Delete(ctx, pendingKey)

	codes, hashes, err := s.generateBackupCodes()
	if err != nil {
		return ports.TotpConfirmResult{}, err
	}
	if err := s.repo.SaveTotpBackupCodes(ctx, user.TenantID, user.ID, hashes); err != nil {
		return ports.TotpConfirmResult{}, err
	}

	return ports.TotpConfirmResult{BackupCodes: codes}, nil
}

func (s *userService) Disable2FA(ctx context.Context, cmd ports.Disable2FACommand) error {
	user, err := s.repo.FindUserByID(ctx, cmd.TenantID, cmd.UserID)
	if err != nil {
		return err
	}
	if !user.TotpEnabled {
		return domain.Err2FANotEnabled
	}
	policy := s.resolveTotpPolicyOrDefault(ctx, user)
	if !s.canDisableTotp(policy) {
		return domain.Err2FAPolicyForbidden
	}
	if cmd.Password == "" {
		return domain.Err2FAPasswordRequired
	}
	if !s.hasher.Verify(user.PasswordHash, cmd.Password) {
		return domain.ErrInvalidCredentials
	}
	ok, err := s.validateTotpCode(ctx, user, cmd.Code)
	if err != nil {
		return err
	}
	if !ok {
		return domain.Err2FAInvalidCode
	}

	user.TotpEnabled = false
	user.TotpSecretEncrypted = ""
	user.TotpEnabledAt = nil
	if policy.DefaultEnabled {
		user.TotpEnrollmentRequired = true
	} else {
		user.TotpEnrollmentRequired = false
	}
	if err := s.repo.UpdateUserTotp(ctx, user); err != nil {
		return err
	}
	return s.repo.DeleteTotpBackupCodes(ctx, user.TenantID, user.ID)
}

func (s *userService) ApplyTotpPolicyOnSociete(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, defaultEnabled bool) (ports.TotpApplyPolicyResult, error) {
	if !defaultEnabled {
		if err := s.repo.ClearTotpEnrollmentRequiredForSocieteUsers(ctx, tenant, societeID); err != nil {
			return ports.TotpApplyPolicyResult{}, err
		}
		return ports.TotpApplyPolicyResult{}, nil
	}
	count, err := s.repo.MarkTotpEnrollmentRequiredForSocieteUsers(ctx, tenant, societeID)
	if err != nil {
		return ports.TotpApplyPolicyResult{}, err
	}
	return ports.TotpApplyPolicyResult{UsersMarked: count}, nil
}

func (s *userService) ClearTotpEnrollmentRequiredOnSociete(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) error {
	return s.repo.ClearTotpEnrollmentRequiredForSocieteUsers(ctx, tenant, societeID)
}

func (s *userService) validateTotpCode(ctx context.Context, user domain.User, code string) (bool, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return false, nil
	}
	if user.TotpSecretEncrypted != "" {
		secret, err := s.decryptSecret(user.TotpSecretEncrypted)
		if err != nil {
			return false, err
		}
		if totp.Validate(code, secret) {
			return true, nil
		}
	}
	return s.tryBackupCode(ctx, user, code)
}

func (s *userService) tryBackupCode(ctx context.Context, user domain.User, code string) (bool, error) {
	normalized := normalizeBackupCode(code)
	if normalized == "" {
		return false, nil
	}
	hashes, err := s.repo.ListUnusedTotpBackupCodeHashes(ctx, user.TenantID, user.ID)
	if err != nil {
		return false, err
	}
	for _, hash := range hashes {
		if !s.hasher.Verify(hash, normalized) {
			continue
		}
		consumed, err := s.repo.ConsumeTotpBackupCode(ctx, user.TenantID, user.ID, hash, s.clock().UTC())
		if err != nil {
			return false, err
		}
		return consumed, nil
	}
	return false, nil
}

func (s *userService) decryptSecret(encrypted string) (string, error) {
	raw, err := cryptox.Decrypt(s.totpKey, encrypted)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (s *userService) generateBackupCodes() ([]string, []string, error) {
	codes := make([]string, 0, totpBackupCodeCount)
	hashes := make([]string, 0, totpBackupCodeCount)
	for i := 0; i < totpBackupCodeCount; i++ {
		code, err := randomBackupCode()
		if err != nil {
			return nil, nil, err
		}
		hash, err := s.hasher.Hash(normalizeBackupCode(code))
		if err != nil {
			return nil, nil, err
		}
		codes = append(codes, code)
		hashes = append(hashes, hash)
	}
	return codes, hashes, nil
}

func randomBackupCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	for i := 0; i < 8; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[n.Int64()])
	}
	return b.String(), nil
}

func normalizeBackupCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(code, "-", "")))
}

func (s *userService) applyTotpPolicyOnCreate(ctx context.Context, user *domain.User) error {
	var societeID uuid.UUID
	var err error
	if user.EquipeID != nil {
		societeID, err = s.repo.ResolveSocieteIDForEquipe(ctx, user.TenantID, *user.EquipeID)
	} else {
		societeID, err = s.repo.ResolveSocieteIDForUser(ctx, user.TenantID, user.ID)
	}
	if err != nil {
		societes, listErr := s.repo.ListSocietes(ctx, user.TenantID)
		if listErr != nil || len(societes) == 0 {
			return nil
		}
		if societes[0].TotpDefaultEnabled {
			user.TotpEnrollmentRequired = true
		}
		return nil
	}
	societe, err := s.repo.GetSociete(ctx, user.TenantID, societeID)
	if err != nil {
		return nil
	}
	if societe.TotpDefaultEnabled {
		user.TotpEnrollmentRequired = true
	}
	return nil
}

func NewTotpEncryptionKey(cfgKey, jwtKey string, devSeedEnabled bool) ([]byte, error) {
	if cfgKey != "" {
		return cryptox.KeyFromBase64(cfgKey)
	}
	if devSeedEnabled {
		return cryptox.DevKeyFromJWTSigningKey(jwtKey), nil
	}
	return nil, fmt.Errorf("TOTP_ENCRYPTION_KEY is required when DEV_SEED_ENABLED is false")
}
