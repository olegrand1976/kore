package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func registerRequestSettingsRoutes(r interface {
	Get(pattern string, handlerFn http.HandlerFunc)
	Put(pattern string, handlerFn http.HandlerFunc)
}, svc ports.RequestSettingsService, authorizer authx.Authorizer) {
	r.Get("/request-settings", getRequestSettings(svc))
	r.Put("/admin/request-settings", updateRequestSettings(svc, authorizer))
}

func getRequestSettings(svc ports.RequestSettingsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		settings, err := svc.Get(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}

func updateRequestSettings(svc ports.RequestSettingsService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ChannelsEnabled domain.ChannelsEnabled `json:"channelsEnabled"`
			GuidesEnabled   bool                   `json:"guidesEnabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		settings, err := svc.Update(r.Context(), ports.UpdateRequestSettingsCommand{
			TenantID:        identity.TenantID,
			ChannelsEnabled: req.ChannelsEnabled,
			GuidesEnabled:   req.GuidesEnabled,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidRequestChannels) {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, settings)
	}
}
