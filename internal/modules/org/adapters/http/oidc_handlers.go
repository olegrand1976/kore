package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterOIDCRoutes(
	r chi.Router,
	oidc ports.OIDCService,
	idp ports.IdentityProviderService,
	authorizer authx.Authorizer,
) {
	r.Get("/auth/oidc/authorize", oidcAuthorizeHandler(oidc))
	r.Post("/auth/oidc/callback", oidcCallbackHandler(oidc))

	r.Group(func(pr chi.Router) {
		pr.Get("/admin/identity-providers", listIdentityProviders(idp, authorizer))
		pr.Put("/admin/identity-providers/{id}", configureIdentityProvider(idp, authorizer))
	})
}

func oidcAuthorizeHandler(oidc ports.OIDCService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantStr := r.URL.Query().Get("tenant")
		tenantID, err := uuid.Parse(tenantStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid tenant")
			return
		}
		url, err := oidc.AuthorizeURL(r.Context(), ports.OIDCAuthorizeCommand{
			TenantID:      kernel.NewTenantID(tenantID),
			RedirectURI:   r.URL.Query().Get("redirect_uri"),
			CodeChallenge: r.URL.Query().Get("code_challenge"),
			State:         r.URL.Query().Get("state"),
		})
		if err != nil {
			writeOIDCError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"authorizeUrl": url})
	}
}

func oidcCallbackHandler(oidc ports.OIDCService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TenantID     string `json:"tenantId"`
			Code         string `json:"code"`
			RedirectURI  string `json:"redirectUri"`
			CodeVerifier string `json:"codeVerifier"`
			State        string `json:"state"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		tenantID, err := uuid.Parse(req.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid tenant")
			return
		}
		result, err := oidc.HandleCallback(r.Context(), ports.OIDCCallbackCommand{
			TenantID:     kernel.NewTenantID(tenantID),
			Code:         req.Code,
			RedirectURI:  req.RedirectURI,
			CodeVerifier: req.CodeVerifier,
			State:        req.State,
		})
		if err != nil {
			writeOIDCError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func listIdentityProviders(idp ports.IdentityProviderService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := idp.List(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		out := make([]map[string]any, 0, len(items))
		for _, item := range items {
			out = append(out, map[string]any{
				"id":             item.ID,
				"name":           item.Name,
				"issuer":         item.Issuer,
				"clientId":       item.ClientID,
				"jwksUri":        item.JWKSURI,
				"scopes":         item.Scopes,
				"defaultProfile": item.DefaultProfile,
				"enabled":        item.Enabled,
			})
		}
		httpx.WriteData(w, http.StatusOK, out)
	}
}

func configureIdentityProvider(idp ports.IdentityProviderService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			Name           string         `json:"name"`
			Issuer         string         `json:"issuer"`
			ClientID       string         `json:"clientId"`
			ClientSecret   string         `json:"clientSecret"`
			JWKSURI        string         `json:"jwksUri"`
			Scopes         string         `json:"scopes"`
			DefaultProfile domain.Profile `json:"defaultProfile"`
			Enabled        bool           `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := idp.Configure(r.Context(), ports.ConfigureIdPCommand{
			ID:             id,
			TenantID:       identity.TenantID,
			Name:           req.Name,
			Issuer:         req.Issuer,
			ClientID:       req.ClientID,
			ClientSecret:   req.ClientSecret,
			JWKSURI:        req.JWKSURI,
			Scopes:         req.Scopes,
			DefaultProfile: req.DefaultProfile,
			Enabled:        req.Enabled,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"id":             item.ID,
			"name":           item.Name,
			"issuer":         item.Issuer,
			"clientId":       item.ClientID,
			"jwksUri":        item.JWKSURI,
			"scopes":         item.Scopes,
			"defaultProfile": item.DefaultProfile,
			"enabled":        item.Enabled,
		})
	}
}

func writeOIDCError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrSSONotEnabled):
		httpx.WriteError(w, http.StatusForbidden, "SSO_NOT_ENABLED", err.Error())
	case errors.Is(err, domain.ErrInvalidIDPToken):
		httpx.WriteError(w, http.StatusUnauthorized, "INVALID_IDP_TOKEN", err.Error())
	case errors.Is(err, domain.ErrIdentityAlreadyLinked):
		httpx.WriteError(w, http.StatusConflict, "IDENTITY_ALREADY_LINKED", err.Error())
	case errors.Is(err, domain.ErrOIDCStateInvalid):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrSeatLimitReached):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	case errors.Is(err, domain.ErrAccountExpired):
		httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
