package http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kore/kore/internal/modules/org/app"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/httpx"
)

const (
	signupRateLimitWindow = time.Minute
	signupRateLimitMax    = 5
	platformDefaultSeats  = 50
)

func RegisterPlatformRoutes(
	r chi.Router,
	platform ports.PlatformService,
	tokens *authx.TokenIssuer,
	entitlements authx.EntitlementReader,
) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Use(requirePlatformAdmin())
		pr.Get("/platform/overview", platformOverview(platform))
		pr.Get("/platform/settings", platformGetSettings(platform))
		pr.Put("/platform/settings", platformUpdateSettings(platform))
		pr.Post("/platform/tenants", platformProvisionTenant(platform))
	})
}

// RegisterPublicSignupRoutes exposes unauthenticated org self-serve signup.
func RegisterPublicSignupRoutes(
	r chi.Router,
	platform ports.PlatformService,
	appCache cache.Cache,
	keys cache.KeyBuilder,
	enabled bool,
) {
	handler := publicSignup(platform, enabled)
	r.With(signupRateLimit(appCache, keys)).Post("/public/signup", handler)
}

func requirePlatformAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := authx.FromContext(r.Context())
			if !ok || !authx.IsPlatformAdmin(identity) {
				httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "platform admin required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func platformOverview(platform ports.PlatformService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if platform == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "unavailable")
			return
		}
		overview, err := platform.GetOverview(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "internal error")
			return
		}
		httpx.WriteData(w, http.StatusOK, overview)
	}
}

func platformGetSettings(platform ports.PlatformService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if platform == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "unavailable")
			return
		}
		settings, err := platform.GetSettings(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "internal error")
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}

func platformUpdateSettings(platform ports.PlatformService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if platform == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "unavailable")
			return
		}
		var req struct {
			GeminiModel string `json:"geminiModel"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		settings, err := platform.UpdateSettings(r.Context(), ports.UpdatePlatformSettingsCommand{
			GeminiModel: req.GeminiModel,
			ActorUserID: identity.UserID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidGeminiModel) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "internal error")
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}

func platformProvisionTenant(platform ports.PlatformService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if platform == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "unavailable")
			return
		}
		var req struct {
			TenantName    string   `json:"tenantName"`
			RaisonSociale string   `json:"raisonSociale"`
			Pays          string   `json:"pays"`
			AdminLogin    string   `json:"adminLogin"`
			AdminEmail    string   `json:"adminEmail"`
			AdminPassword string   `json:"adminPassword"`
			Seats         int      `json:"seats"`
			Modules       []string `json:"modules"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		seats := req.Seats
		if seats <= 0 {
			seats = platformDefaultSeats
		}
		result, err := platform.ProvisionTenant(r.Context(), ports.ProvisionTenantCommand{
			TenantName:    req.TenantName,
			RaisonSociale: req.RaisonSociale,
			Pays:          req.Pays,
			AdminLogin:    req.AdminLogin,
			AdminEmail:    req.AdminEmail,
			AdminPassword: req.AdminPassword,
			Seats:         seats,
			Modules:       app.SanitizeTrialModules(req.Modules),
		})
		if err != nil {
			writeProvisionError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusCreated, result)
	}
}

func publicSignup(platform ports.PlatformService, enabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !enabled {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "signup disabled")
			return
		}
		if platform == nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "unavailable")
			return
		}
		var req struct {
			RaisonSociale string `json:"raisonSociale"`
			Pays          string `json:"pays"`
			AdminLogin    string `json:"adminLogin"`
			AdminEmail    string `json:"adminEmail"`
			AdminPassword string `json:"adminPassword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		result, err := platform.ProvisionTenant(r.Context(), ports.ProvisionTenantCommand{
			TenantName:    req.RaisonSociale,
			RaisonSociale: req.RaisonSociale,
			Pays:          req.Pays,
			AdminLogin:    req.AdminLogin,
			AdminEmail:    req.AdminEmail,
			AdminPassword: req.AdminPassword,
			Seats:         10,
			Modules:       app.DefaultTrialModules(),
		})
		if err != nil {
			writeProvisionError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusCreated, result)
	}
}

func writeProvisionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrLoginAlreadyExists):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidPays):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrInvalidLogin),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrProvisionInputRequired):
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "internal error")
	}
}

func signupRateLimit(appCache cache.Cache, keys cache.KeyBuilder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appCache == nil || keys == nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "rate limit unavailable")
				return
			}
			ip := signupClientIP(r)
			key := keys.PublicKey("org", "ratelimit", "signup", ip)
			var count int
			found, err := appCache.Get(r.Context(), key, &count)
			if err != nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "rate limit unavailable")
				return
			}
			if found && count >= signupRateLimitMax {
				httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrCodeTooManyRequests, "too many requests")
				return
			}
			count++
			if err := appCache.Set(r.Context(), key, count, signupRateLimitWindow); err != nil {
				httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "rate limit unavailable")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// signupClientIP uses RemoteAddr after chi RealIP middleware (trusted proxy).
func signupClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
