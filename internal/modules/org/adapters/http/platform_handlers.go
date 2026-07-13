package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kore/kore/internal/modules/org/domain"
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
		pr.Get("/platform/settings", platformGetSettings(platform))
		pr.Put("/platform/settings", platformUpdateSettings(platform))
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

func platformGetSettings(platform ports.PlatformService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		settings, err := platform.GetSettings(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}

func platformUpdateSettings(platform ports.PlatformService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}
