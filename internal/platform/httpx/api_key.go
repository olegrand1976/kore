package httpx

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
)

const (
	apiKeyHeader       = "X-Api-Key"
	apiKeyRateLimitMax = 1000
	apiKeyRateWindow   = time.Hour
)

// ApiKeyLookup resolves API keys stored as SHA-256 hashes.
type ApiKeyLookup interface {
	GetApiKeyByHash(ctx context.Context, keyHash string) (domain.ApiKey, error)
	TouchApiKeyUsed(ctx context.Context, key domain.ApiKey) error
}

type apiKeyLookupFunc struct {
	lookup func(ctx context.Context, keyHash string) (domain.ApiKey, error)
	touch  func(ctx context.Context, key domain.ApiKey) error
}

func NewApiKeyLookup(
	lookup func(ctx context.Context, keyHash string) (domain.ApiKey, error),
	touch func(ctx context.Context, key domain.ApiKey) error,
) ApiKeyLookup {
	return &apiKeyLookupFunc{lookup: lookup, touch: touch}
}

func (f *apiKeyLookupFunc) GetApiKeyByHash(ctx context.Context, keyHash string) (domain.ApiKey, error) {
	return f.lookup(ctx, keyHash)
}

func (f *apiKeyLookupFunc) TouchApiKeyUsed(ctx context.Context, key domain.ApiKey) error {
	if f.touch == nil {
		return nil
	}
	return f.touch(ctx, key)
}

// ApiKeyMiddleware validates X-Api-Key and applies Redis-backed rate limiting.
func ApiKeyMiddleware(lookup ApiKeyLookup, appCache cache.Cache, keys cache.KeyBuilder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := strings.TrimSpace(r.Header.Get(apiKeyHeader))
			if rawKey == "" {
				if auth := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(auth), "bearer kore_") {
					rawKey = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
				}
			}
			if rawKey == "" {
				WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "missing api key")
				return
			}
			sum := sha256.Sum256([]byte(rawKey))
			keyHash := fmt.Sprintf("%x", sum)
			apiKey, err := lookup.GetApiKeyByHash(r.Context(), keyHash)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid api key")
				return
			}
			if apiKey.IsRevoked() {
				WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "api key revoked")
				return
			}
			if appCache != nil && keys != nil {
				rateKey := keys.Key(apiKey.TenantID, "integrations", "apikey-ratelimit", apiKey.ID.String())
				var count int
				found, err := appCache.Get(r.Context(), rateKey, &count)
				if err == nil && found && count >= apiKeyRateLimitMax {
					WriteError(w, http.StatusTooManyRequests, ErrCodeTooManyRequests, "api key rate limit exceeded")
					return
				}
				count++
				_ = appCache.Set(r.Context(), rateKey, count, apiKeyRateWindow)
			}
			identity := authx.Identity{
				UserID:   apiKey.ID,
				TenantID: apiKey.TenantID,
				Profile:  authx.ProfileAdmin,
				Roles:    []string{"api_key"},
			}
			ctx := authx.WithIdentity(r.Context(), identity)
			_ = lookup.TouchApiKeyUsed(ctx, apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PublicAPIStack combines API key auth without JWT entitlement checks.
func PublicAPIStack(lookup ApiKeyLookup, appCache cache.Cache, keys cache.KeyBuilder) func(http.Handler) http.Handler {
	return ApiKeyMiddleware(lookup, appCache, keys)
}

// TouchApiKeyNoop is a no-op touch helper when last_used_at update is optional.
func TouchApiKeyNoop(_ context.Context, _ domain.ApiKey) error { return nil }
