package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func registerLeaveTypeConfigRoutes(
	r chi.Router,
	configs ports.LeaveTypeConfigService,
	authorizer authx.Authorizer,
) {
	r.Get("/leave-type-configs/mine", listMyLeaveTypeConfigs(configs))
	r.Get("/leave-type-configs", listLeaveTypeConfigs(configs, authorizer))
	r.Post("/leave-type-configs/reset", resetLeaveTypeConfigs(configs, authorizer))
	r.Post("/leave-type-configs", createLeaveTypeConfig(configs, authorizer))
	r.Put("/leave-type-configs/{id}", updateLeaveTypeConfig(configs, authorizer))
	r.Delete("/leave-type-configs/{id}", deleteLeaveTypeConfig(configs, authorizer))
}

func listMyLeaveTypeConfigs(configs ports.LeaveTypeConfigService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		items, err := configs.ListForUser(r.Context(), identity.TenantID, identity.UserID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func listLeaveTypeConfigs(configs ports.LeaveTypeConfigService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		societeID, err := uuid.Parse(r.URL.Query().Get("societeId"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid societeId")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		activeOnly := !authorizer.Can(r.Context(), "org", authx.ActionWrite)
		items, err := configs.List(r.Context(), identity.TenantID, societeID, activeOnly)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createLeaveTypeConfig(configs ports.LeaveTypeConfigService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			SocieteID     uuid.UUID `json:"societeId"`
			Code          string    `json:"code"`
			Label         string    `json:"label"`
			TracksBalance bool      `json:"tracksBalance"`
			Active        bool      `json:"active"`
			SortOrder     int       `json:"sortOrder"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		created, err := configs.Create(r.Context(), ports.CreateLeaveTypeConfigCommand{
			TenantID:      identity.TenantID,
			SocieteID:     req.SocieteID,
			Code:          req.Code,
			Label:         req.Label,
			TracksBalance: req.TracksBalance,
			Active:        req.Active,
			SortOrder:     req.SortOrder,
		})
		if err != nil {
			if errors.Is(err, domain.ErrLeaveTypeCodeExists) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, created)
	}
}

func updateLeaveTypeConfig(configs ports.LeaveTypeConfigService, authorizer authx.Authorizer) http.HandlerFunc {
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
			Label         string `json:"label"`
			TracksBalance bool   `json:"tracksBalance"`
			Active        bool   `json:"active"`
			SortOrder     int    `json:"sortOrder"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		updated, err := configs.Update(r.Context(), ports.UpdateLeaveTypeConfigCommand{
			TenantID:      identity.TenantID,
			ID:            id,
			Label:         req.Label,
			TracksBalance: req.TracksBalance,
			Active:        req.Active,
			SortOrder:     req.SortOrder,
		})
		if err != nil {
			if errors.Is(err, domain.ErrLeaveTypeNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, updated)
	}
}

func deleteLeaveTypeConfig(configs ports.LeaveTypeConfigService, authorizer authx.Authorizer) http.HandlerFunc {
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
		identity, _ := authx.FromContext(r.Context())
		err = configs.Delete(r.Context(), identity.TenantID, id)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrLeaveTypeNotFound):
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			case errors.Is(err, domain.ErrLeaveTypeInUse):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func resetLeaveTypeConfigs(configs ports.LeaveTypeConfigService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			SocieteID uuid.UUID `json:"societeId"`
			Confirm   bool      `json:"confirm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if !req.Confirm {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "confirm required")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := configs.ResetDefaults(r.Context(), ports.ResetLeaveTypeConfigsCommand{
			TenantID:  identity.TenantID,
			SocieteID: req.SocieteID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrUnsupportedCountry) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}
