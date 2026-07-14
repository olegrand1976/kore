package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/internal/platform/logging"
)

type ErrorCode string

const (
	ErrCodeUnauthorized           ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden              ErrorCode = "FORBIDDEN"
	ErrCodeNotFound               ErrorCode = "NOT_FOUND"
	ErrCodeValidation             ErrorCode = "VALIDATION_ERROR"
	ErrCodeConflict               ErrorCode = "CONFLICT"
	ErrCodePaymentRequired        ErrorCode = "PAYMENT_REQUIRED"
	ErrCodeTooManyRequests        ErrorCode = "TOO_MANY_REQUESTS"
	ErrCodeInternal               ErrorCode = "INTERNAL_ERROR"
	ErrCodeCRAAlreadyValidated    ErrorCode = "CRA_ALREADY_VALIDATED"
	ErrCodeCommercialInfoRequired ErrorCode = "COMMERCIAL_INFO_REQUIRED"
	ErrCodeDayCapacityExceeded    ErrorCode = "DAY_CAPACITY_EXCEEDED"
	ErrCodeCRAConflictAbsence     ErrorCode = "CRA_CONFLICT_ABSENCE"
	ErrCodeWeekIncomplete         ErrorCode = "WEEK_INCOMPLETE"
)

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type Envelope struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type Router struct {
	chi.Router
	log *logging.Logger
}

type Dependencies struct {
	Logger            *logging.Logger
	Pool              *db.Pool
	Cache             cache.Cache
	TokenIssuer       *authx.TokenIssuer
	EntitlementReader authx.EntitlementReader
	Authorizer        authx.Authorizer
}

func NewRouter(deps Dependencies) *Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	router := &Router{Router: r, log: deps.Logger}
	return router
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code ErrorCode, message string) {
	WriteJSON(w, status, Envelope{Error: &APIError{Code: code, Message: message}})
}

func WriteData(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, Envelope{Data: data})
}

func MapDomainError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, authx.ErrUnauthorized):
		WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, err.Error())
		return true
	case errors.Is(err, authx.ErrForbidden):
		WriteError(w, http.StatusForbidden, ErrCodeForbidden, err.Error())
		return true
	case errors.Is(err, authx.ErrPaymentRequired):
		WriteError(w, http.StatusPaymentRequired, ErrCodePaymentRequired, err.Error())
		return true
	default:
		return false
	}
}

func (r *Router) MountHealth(pool *db.Pool, pingRedis func(r *http.Request) error) {
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/ready", func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if err := pool.Ping(ctx); err != nil {
			WriteError(w, http.StatusServiceUnavailable, ErrCodeInternal, "database unavailable")
			return
		}
		if pingRedis != nil {
			if err := pingRedis(req); err != nil {
				WriteError(w, http.StatusServiceUnavailable, ErrCodeInternal, "redis unavailable")
				return
			}
		}
		WriteData(w, http.StatusOK, map[string]string{"status": "ready"})
	})
}

func AuthStack(issuer *authx.TokenIssuer, entitlements authx.EntitlementReader) func(http.Handler) http.Handler {
	auth := AuthMiddleware(issuer)
	ent := EntitlementMiddleware(entitlements)
	return func(next http.Handler) http.Handler {
		return auth(ent(next))
	}
}

func AuthMiddleware(issuer *authx.TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := authx.BearerToken(r)
			if token == "" {
				WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "missing bearer token")
				return
			}
			identity, err := issuer.ParseAccessToken(token)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid token")
				return
			}
			ctx := authx.WithIdentity(r.Context(), identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequirePermission(authz authx.Authorizer, module authx.Module, action authx.Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authz.Can(r.Context(), module, action) {
				WriteError(w, http.StatusForbidden, ErrCodeForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func PublicOnly(next http.Handler) http.Handler {
	return next
}
