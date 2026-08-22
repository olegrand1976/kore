package http

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/httpx"
)

const (
	taigaWebhookRateLimitWindow = time.Minute
	taigaWebhookRateLimitMax    = 60
)

func taigaWebhookRateLimit(appCache cache.Cache, keys cache.KeyBuilder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appCache == nil || keys == nil {
				next.ServeHTTP(w, r)
				return
			}
			ip := taigaWebhookClientIP(r)
			key := keys.PublicKey("integrations", "ratelimit", "taiga-webhook", ip)
			var count int
			found, err := appCache.Get(r.Context(), key, &count)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if found && count >= taigaWebhookRateLimitMax {
				httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrCodeTooManyRequests, "too many requests")
				return
			}
			count++
			_ = appCache.Set(r.Context(), key, count, taigaWebhookRateLimitWindow)
			next.ServeHTTP(w, r)
		})
	}
}

func taigaWebhookClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.IndexByte(xff, ','); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
