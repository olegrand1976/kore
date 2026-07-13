package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

type tenantDiscoveryRequest struct {
	Email string `json:"email"`
}

func tenantDiscoveryRequestHandler(svc ports.TenantAccessService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tenantDiscoveryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		base := baseLoginURL(r)
		_ = svc.RequestTenantDiscovery(r.Context(), req.Email, base)
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

func createInvitationHandler(svc ports.TenantAccessService, authorizer authx.Authorizer) http.HandlerFunc {
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
		base := baseLoginURL(r)
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

func baseLoginURL(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("x-public-base-url")); v != "" {
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
			return strings.TrimRight(v, "/") + "/login"
		}
	}
	proto := r.Header.Get("x-forwarded-proto")
	if proto == "" {
		proto = "http"
	}
	host := r.Host
	if host == "" {
		host = "localhost"
	}
	return proto + "://" + host + "/login"
}

