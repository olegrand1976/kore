package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/app"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterTaigaRoutes(r chi.Router, taiga *app.TaigaService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader, appCache cache.Cache, keys cache.KeyBuilder, webhookSecret, defaultTenantID string) {
	if taiga == nil {
		return
	}
	r.With(taigaWebhookRateLimit(appCache, keys)).Post("/integrations/taiga/webhook", taigaWebhook(taiga, webhookSecret, defaultTenantID))
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/integrations/taiga/links/by-demand/{id}", findTaigaLinkByDemand(taiga, authorizer))
		pr.Post("/integrations/taiga/user-mappings", upsertTaigaUserMapping(taiga, authorizer))
	})
}

func taigaWebhook(taiga *app.TaigaService, webhookSecret, defaultTenantID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if webhookSecret == "" {
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrCodeInternal, "taiga webhook not configured")
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if !verifyTaigaWebhookAuth(r, body, webhookSecret) {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "invalid webhook secret")
			return
		}
		tenantRaw := r.Header.Get("X-Kore-Tenant-ID")
		if tenantRaw == "" {
			tenantRaw = defaultTenantID
		}
		if tenantRaw == "" {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "missing tenant")
			return
		}
		tenantID, err := uuid.Parse(tenantRaw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid tenant")
			return
		}
		if err := taiga.HandleWebhook(r.Context(), kernel.NewTenantID(tenantID), body); err != nil {
			switch {
			case errors.Is(err, domain.ErrInvalidKoreDemandID), errors.Is(err, domain.ErrKoreDemandNotFound):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			default:
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func findTaigaLinkByDemand(taiga *app.TaigaService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		demandID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid demand id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		link, err := taiga.FindByKoreDemand(r.Context(), identity.TenantID, demandID)
		if errors.Is(err, domain.ErrExternalLinkNotFound) {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "link not found")
			return
		}
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, link)
	}
}

func upsertTaigaUserMapping(taiga *app.TaigaService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "integrations", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			TaigaUserID   int       `json:"taigaUserId"`
			TaigaUsername string    `json:"taigaUsername"`
			KoreUserID    uuid.UUID `json:"koreUserId"`
			MatchMethod   string    `json:"matchMethod"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		mapping, err := taiga.UpsertUserMapping(r.Context(), app.UpsertUserMappingCommand{
			TenantID:      identity.TenantID,
			TaigaUserID:   req.TaigaUserID,
			TaigaUsername: req.TaigaUsername,
			KoreUserID:    req.KoreUserID,
			MatchMethod:   req.MatchMethod,
		})
		if errors.Is(err, domain.ErrInvalidTaigaUserID) ||
			errors.Is(err, domain.ErrInvalidKoreUserID) ||
			errors.Is(err, domain.ErrInvalidMatchMethod) {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, mapping)
	}
}
