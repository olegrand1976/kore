package http

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/httpx"
)

const (
	passwordResetRateLimitWindow = time.Minute
	passwordResetRateLimitMax    = 5
)

type tenantDiscoveryRequest struct {
	Email string `json:"email"`
}

func tenantDiscoveryRequestHandler(svc ports.TenantAccessService, configuredPublicBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tenantDiscoveryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		base := baseLoginURL(configuredPublicBaseURL)
		if err := svc.RequestTenantDiscovery(r.Context(), req.Email, base); err != nil {
			log.Printf("tenant discovery email failed for %q: %v", req.Email, err)
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"sent": true})
	}
}

func tenantDiscoveryResolveHandler(svc ports.TenantAccessService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		res, err := svc.Resolve(r.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrAccessTokenInvalid),
				errors.Is(err, domain.ErrAccessTokenExpired),
				errors.Is(err, domain.ErrAccessTokenUsed):
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, res)
	}
}

func invitationResolveHandler(svc ports.TenantAccessService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		res, err := svc.Resolve(r.Context(), token)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "invalid invite")
			return
		}
		httpx.WriteData(w, http.StatusOK, res)
	}
}

type createInvitationRequest struct {
	Email string `json:"email"`
}

func createInvitationHandler(svc ports.TenantAccessService, authorizer authx.Authorizer, configuredPublicBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req createInvitationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		base := baseLoginURL(configuredPublicBaseURL)
		if err := svc.CreateInvitation(r.Context(), identity.TenantID, req.Email, base); err != nil {
			if errors.Is(err, domain.ErrInvalidEmail) {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"sent": true})
	}
}

func baseLoginURL(configuredPublicBaseURL string) string {
	return resolvePublicBaseURL(configuredPublicBaseURL) + "/login"
}

func baseResetPasswordURL(configuredPublicBaseURL string) string {
	return resolvePublicBaseURL(configuredPublicBaseURL) + "/reset-password"
}

// resolvePublicBaseURL prefers the configured PUBLIC_BASE_URL (trusted) over request headers.
func resolvePublicBaseURL(configured string) string {
	if v := strings.TrimSpace(configured); v != "" {
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
			return strings.TrimRight(v, "/")
		}
	}
	return "http://localhost:3001"
}

type passwordResetRequestBody struct {
	Email string `json:"email"`
}

func passwordResetRequestHandler(svc ports.TenantAccessService, configuredPublicBaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req passwordResetRequestBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		base := baseResetPasswordURL(configuredPublicBaseURL)
		if err := svc.RequestPasswordReset(r.Context(), req.Email, base); err != nil {
			log.Printf("password reset email failed for %q: %v", req.Email, err)
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"sent": true})
	}
}

type passwordResetConfirmBody struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func passwordResetConfirmHandler(svc ports.TenantAccessService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req passwordResetConfirmBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if err := svc.ConfirmPasswordReset(r.Context(), req.Token, req.NewPassword); err != nil {
			switch {
			case errors.Is(err, domain.ErrWeakPassword):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			case errors.Is(err, domain.ErrAccessTokenInvalid),
				errors.Is(err, domain.ErrAccessTokenExpired),
				errors.Is(err, domain.ErrAccessTokenUsed),
				errors.Is(err, domain.ErrUserNotFound),
				errors.Is(err, domain.ErrAccountExpired):
				httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func passwordResetRateLimit(appCache cache.Cache, keys cache.KeyBuilder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appCache == nil || keys == nil {
				next.ServeHTTP(w, r)
				return
			}
			ip := passwordResetClientIP(r)
			key := keys.PublicKey("org", "ratelimit", "password-reset", ip)
			var count int
			found, err := appCache.Get(r.Context(), key, &count)
			if err != nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "rate limit unavailable")
				return
			}
			if found && count >= passwordResetRateLimitMax {
				httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrCodeTooManyRequests, "too many requests")
				return
			}
			count++
			if err := appCache.Set(r.Context(), key, count, passwordResetRateLimitWindow); err != nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "rate limit unavailable")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func passwordResetClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
