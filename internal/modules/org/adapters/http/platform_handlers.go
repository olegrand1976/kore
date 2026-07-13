package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
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
	})
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
		overview, err := platform.GetOverview(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, overview)
	}
}
