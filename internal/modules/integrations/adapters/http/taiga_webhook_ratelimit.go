package http

import (
	"context"
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

func allowTaigaWebhookRequest(ctx context.Context, appCache cache.Cache, keys cache.KeyBuilder, r *http.Request) bool {
	if appCache == nil || keys == nil {
		return true
	}
	ip := taigaWebhookClientIP(r)
	key := keys.PublicKey("integrations", "ratelimit", "taiga-webhook", ip)
	var count int
	found, err := appCache.Get(ctx, key, &count)
	if err != nil {
		return true
	}
	if found && count >= taigaWebhookRateLimitMax {
		return false
	}
	count++
	_ = appCache.Set(ctx, key, count, taigaWebhookRateLimitWindow)
	return true
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

func writeTaigaWebhookRateLimited(w http.ResponseWriter) {
	httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrCodeTooManyRequests, "too many requests")
}
