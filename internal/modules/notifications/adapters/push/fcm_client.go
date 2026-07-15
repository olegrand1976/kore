package push

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kore/kore/internal/modules/notifications/ports"
)

const (
	fcmScope         = "https://www.googleapis.com/auth/firebase.messaging"
	oauthTokenURL    = "https://oauth2.googleapis.com/token"
	metadataTokenURL = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"
	fcmMessageFmt    = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
)

type FCMConfig struct {
	ProjectID       string
	CredentialsPath string
}

type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	ProjectID   string `json:"project_id"`
}

type cachedAccessToken struct {
	value     string
	expiresAt time.Time
}

type FCMClient struct {
	projectID string
	sa        *serviceAccount
	useMeta   bool
	http      *http.Client
	mu        sync.Mutex
	token     cachedAccessToken
}

func NewFCMClient(cfg FCMConfig) (*FCMClient, error) {
	projectID := strings.TrimSpace(cfg.ProjectID)
	creds := strings.TrimSpace(cfg.CredentialsPath)
	if creds == "" {
		creds = strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	}

	client := &FCMClient{
		projectID: projectID,
		http:      &http.Client{Timeout: 15 * time.Second},
	}

	if creds != "" {
		raw, err := loadCredentialBytes(creds)
		if err != nil {
			return nil, fmt.Errorf("fcm credentials: %w", err)
		}
		var sa serviceAccount
		if err := json.Unmarshal(raw, &sa); err != nil {
			return nil, fmt.Errorf("fcm credentials json: %w", err)
		}
		if sa.ClientEmail == "" || sa.PrivateKey == "" {
			return nil, fmt.Errorf("fcm credentials: missing client_email or private_key")
		}
		client.sa = &sa
		if projectID == "" {
			client.projectID = strings.TrimSpace(sa.ProjectID)
		}
	} else {
		client.useMeta = true
	}

	if client.projectID == "" {
		return nil, fmt.Errorf("fcm: project id required")
	}
	return client, nil
}

func loadCredentialBytes(pathOrJSON string) ([]byte, error) {
	pathOrJSON = strings.TrimSpace(pathOrJSON)
	if pathOrJSON == "" {
		return nil, fmt.Errorf("empty credentials")
	}
	if strings.HasPrefix(pathOrJSON, "{") {
		return []byte(pathOrJSON), nil
	}
	return os.ReadFile(pathOrJSON)
}

func (c *FCMClient) Send(ctx context.Context, tokens []string, msg ports.PushMessage) error {
	if len(tokens) == 0 {
		return nil
	}
	accessToken, err := c.accessToken(ctx)
	if err != nil {
		return fmt.Errorf("fcm auth: %w", err)
	}

	var invalid []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if err := c.sendOne(ctx, accessToken, token, msg); err != nil {
			if isInvalidFCMToken(err) {
				invalid = append(invalid, token)
				continue
			}
			return err
		}
	}
	if len(invalid) > 0 {
		return &invalidTokensError{tokens: invalid}
	}
	return nil
}

func (c *FCMClient) sendOne(ctx context.Context, accessToken, deviceToken string, msg ports.PushMessage) error {
	message := map[string]any{
		"token": deviceToken,
		"notification": map[string]string{
			"title": msg.Title,
			"body":  msg.Body,
		},
	}
	if len(msg.Data) > 0 {
		message["data"] = msg.Data
	}
	payload, err := json.Marshal(map[string]any{"message": message})
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf(fcmMessageFmt, c.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("fcm http: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("fcm status %d: %s", resp.StatusCode, string(raw))
}

func (c *FCMClient) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token.value != "" && time.Now().Before(c.token.expiresAt.Add(-30*time.Second)) {
		return c.token.value, nil
	}
	var (
		value string
		ttl   time.Duration
		err   error
	)
	if c.sa != nil {
		value, ttl, err = c.tokenFromServiceAccount(ctx)
	} else if c.useMeta {
		value, ttl, err = c.tokenFromMetadata(ctx)
	} else {
		return "", fmt.Errorf("fcm: no credentials configured")
	}
	if err != nil {
		return "", err
	}
	c.token = cachedAccessToken{value: value, expiresAt: time.Now().Add(ttl)}
	return value, nil
}

func (c *FCMClient) tokenFromServiceAccount(ctx context.Context) (string, time.Duration, error) {
	key, err := parseRSAPrivateKey(c.sa.PrivateKey)
	if err != nil {
		return "", 0, err
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":   c.sa.ClientEmail,
		"sub":   c.sa.ClientEmail,
		"aud":   oauthTokenURL,
		"iat":   now.Unix(),
		"exp":   now.Add(55 * time.Minute).Unix(),
		"scope": fcmScope,
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return "", 0, err
	}
	return exchangeGoogleToken(ctx, c.http, signed)
}

func (c *FCMClient) tokenFromMetadata(ctx context.Context) (string, time.Duration, error) {
	reqURL, err := url.Parse(metadataTokenURL)
	if err != nil {
		return "", 0, err
	}
	q := reqURL.Query()
	q.Set("scopes", fcmScope)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("metadata token: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", 0, fmt.Errorf("metadata token status %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", 0, err
	}
	if out.AccessToken == "" {
		return "", 0, fmt.Errorf("metadata token: empty access_token")
	}
	ttl := time.Duration(out.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = 55 * time.Minute
	}
	return out.AccessToken, ttl, nil
}

func exchangeGoogleToken(ctx context.Context, client *http.Client, jwtAssertion string) (string, time.Duration, error) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", jwtAssertion)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", 0, fmt.Errorf("oauth token status %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", 0, err
	}
	ttl := time.Duration(out.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = 55 * time.Minute
	}
	return out.AccessToken, ttl, nil
}

func parseRSAPrivateKey(pemKey string) (*rsa.PrivateKey, error) {
	pemKey = strings.ReplaceAll(pemKey, `\n`, "\n")
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("invalid private key pem")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return key.(*rsa.PrivateKey), nil
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not rsa")
	}
	return rsaKey, nil
}

type invalidTokensError struct {
	tokens []string
}

func (e *invalidTokensError) Error() string {
	return fmt.Sprintf("fcm: %d invalid token(s)", len(e.tokens))
}

func (e *invalidTokensError) InvalidTokens() []string {
	return e.tokens
}

func isInvalidFCMToken(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	var payload struct {
		Error struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if idx := strings.Index(msg, "{"); idx >= 0 {
		if json.Unmarshal([]byte(msg[idx:]), &payload) == nil {
			switch strings.ToUpper(payload.Error.Status) {
			case "NOT_FOUND", "INVALID_ARGUMENT", "UNREGISTERED":
				return true
			}
		}
	}
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "not_found") ||
		strings.Contains(lower, "unregistered") ||
		strings.Contains(lower, "invalid_argument") ||
		strings.Contains(lower, "registration token")
}
