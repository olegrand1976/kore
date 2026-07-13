package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/admin/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, svc ports.AdminService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/admin/parameters", getParameters(svc, authorizer))
		pr.Put("/admin/parameters", upsertParameters(svc, authorizer))
		pr.Get("/admin/templates", listTemplates(svc, authorizer))
		pr.Post("/admin/templates", createTemplate(svc, authorizer))
		pr.Get("/admin/templates/{id}", getTemplate(svc, authorizer))
		pr.Get("/admin/phone-directory", listPhoneDirectory(svc, authorizer))
		pr.Post("/admin/phone-directory", createPhoneEntry(svc, authorizer))
	})
}

func getParameters(svc ports.AdminService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "admin", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			code = "default"
		}
		identity, _ := authx.FromContext(r.Context())
		ps, err := svc.GetParameters(r.Context(), identity.TenantID, code)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, ps)
	}
}

func upsertParameters(svc ports.AdminService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "admin", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Code    string         `json:"code"`
			Payload map[string]any `json:"payload"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if req.Code == "" {
			req.Code = "default"
		}
		identity, _ := authx.FromContext(r.Context())
		ps, err := svc.UpsertParameters(r.Context(), ports.UpsertParametersCommand{
			TenantID: identity.TenantID,
			Code:     req.Code,
			Payload:  req.Payload,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, ps)
	}
}

func listTemplates(svc ports.AdminService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "admin", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.ListTemplates(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createTemplate(svc ports.AdminService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "admin", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Type    string         `json:"type"`
			Name    string         `json:"name"`
			Content map[string]any `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		t, err := svc.CreateTemplate(r.Context(), ports.CreateTemplateCommand{
			TenantID: identity.TenantID,
			Type:     req.Type,
			Name:     req.Name,
			Content:  req.Content,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, t)
	}
}

func getTemplate(svc ports.AdminService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "admin", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		t, err := svc.GetTemplate(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, t)
	}
}

func listPhoneDirectory(svc ports.AdminService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "admin", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.ListPhoneDirectory(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createPhoneEntry(svc ports.AdminService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "admin", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			UserID     *uuid.UUID `json:"userId"`
			Label      string     `json:"label"`
			Phone      string     `json:"phone"`
			Visibility string     `json:"visibility"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		e, err := svc.CreatePhoneEntry(r.Context(), ports.CreatePhoneEntryCommand{
			TenantID:   identity.TenantID,
			UserID:     req.UserID,
			Label:      req.Label,
			Phone:      req.Phone,
			Visibility: req.Visibility,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, e)
	}
}
