package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/internal/modules/maintenance/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, svc ports.MaintenanceService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/work-requests", listWorkRequests(svc, authorizer))
		pr.Post("/work-requests", createWorkRequest(svc, authorizer))
		pr.Get("/work-requests/{id}", getWorkRequest(svc, authorizer))
		pr.Post("/work-requests/{id}/assign", assignWorkRequest(svc, authorizer))
		pr.Post("/work-requests/{id}/progress", progressWorkRequest(svc, authorizer))
		pr.Post("/work-requests/{id}/complete", completeWorkRequest(svc, authorizer))
	})
}

func listWorkRequests(svc ports.MaintenanceService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "maintenance", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.List(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createWorkRequest(svc ports.MaintenanceService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "maintenance", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ApplicationID uuid.UUID `json:"applicationId"`
			Subject       string    `json:"subject"`
			Description   string    `json:"description"`
			Priority      string    `json:"priority"`
			DueAt         *string   `json:"dueAt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		var dueAt *time.Time
		if req.DueAt != nil && *req.DueAt != "" {
			parsed, err := time.Parse(time.RFC3339, *req.DueAt)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid dueAt")
				return
			}
			dueAt = &parsed
		}
		identity, _ := authx.FromContext(r.Context())
		wr, err := svc.Create(r.Context(), ports.CreateWorkRequestCommand{
			TenantID:      identity.TenantID,
			ApplicationID: req.ApplicationID,
			Subject:       req.Subject,
			Description:   req.Description,
			Priority:      req.Priority,
			DueAt:         dueAt,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, wr)
	}
}

func getWorkRequest(svc ports.MaintenanceService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "maintenance", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		wr, err := svc.Get(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, wr)
	}
}

func assignWorkRequest(svc ports.MaintenanceService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "maintenance", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			AssigneeID uuid.UUID `json:"assigneeId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		wr, err := svc.Assign(r.Context(), ports.AssignCommand{
			TenantID:   identity.TenantID,
			RequestID:  id,
			AssigneeID: req.AssigneeID,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, wr)
	}
}

func progressWorkRequest(svc ports.MaintenanceService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "maintenance", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			ConsumptionDays float64 `json:"consumptionDays"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		wr, err := svc.Progress(r.Context(), ports.ProgressCommand{
			TenantID:        identity.TenantID,
			RequestID:       id,
			ConsumptionDays: req.ConsumptionDays,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, wr)
	}
}

func completeWorkRequest(svc ports.MaintenanceService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "maintenance", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		wr, err := svc.Complete(r.Context(), identity.TenantID, id)
		if err != nil {
			writeMaintenanceError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, wr)
	}
}

func writeMaintenanceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidWorkState):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, cradomain.ErrCRAAlreadyValidated):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeCRAAlreadyValidated, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
