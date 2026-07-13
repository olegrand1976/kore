package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, svc ports.SSIIService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/missions", listMissions(svc, authorizer))
		pr.Post("/missions", createMission(svc, authorizer))
		pr.Get("/missions/{id}", getMission(svc, authorizer))
		pr.Post("/missions/{id}/stop", stopMission(svc, authorizer))
		pr.Put("/missions/{id}/end-date", updateEndDate(svc, authorizer))
	})
}

func listMissions(svc ports.SSIIService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		canRead := authorizer.Can(r.Context(), "ssii", authx.ActionRead)
		canReadCra := authorizer.Can(r.Context(), "cra", authx.ActionRead)
		if !canRead && !canReadCra {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.ListSummaries(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createMission(svc ports.SSIIService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ssii", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ClientID        uuid.UUID   `json:"clientId"`
			StartDate       time.Time   `json:"startDate"`
			EndDate         *time.Time  `json:"endDate"`
			TJMAmount       int64       `json:"tjmAmount"`
			Currency        string      `json:"currency"`
			Technologies    []string    `json:"technologies"`
			ClientContact   string      `json:"clientContact"`
			CollaboratorIDs []uuid.UUID `json:"collaboratorIds"`
			CountryCode     string      `json:"countryCode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		m, err := svc.Create(r.Context(), ports.CreateMissionCommand{
			TenantID:        identity.TenantID,
			ClientID:        req.ClientID,
			StartDate:       req.StartDate,
			EndDate:         req.EndDate,
			TJMAmount:       req.TJMAmount,
			Currency:        req.Currency,
			Technologies:    req.Technologies,
			ClientContact:   req.ClientContact,
			CollaboratorIDs: req.CollaboratorIDs,
			CountryCode:     req.CountryCode,
		})
		if err != nil {
			if errors.Is(err, domain.ErrMissionWithoutCollaborator) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, "MISSION_WITHOUT_COLLABORATOR", err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, m)
	}
}

func getMission(svc ports.SSIIService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		canRead := authorizer.Can(r.Context(), "ssii", authx.ActionRead)
		canReadCra := authorizer.Can(r.Context(), "cra", authx.ActionRead)
		if !canRead && !canReadCra {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		m, err := svc.GetDetail(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, m)
	}
}

func stopMission(svc ports.SSIIService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ssii", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		m, err := svc.Stop(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, m)
	}
}

func updateEndDate(svc ports.SSIIService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ssii", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			EndDate time.Time `json:"endDate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		m, err := svc.UpdateEndDate(r.Context(), ports.UpdateEndDateCommand{
			TenantID:  identity.TenantID,
			MissionID: id,
			EndDate:   req.EndDate,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, m)
	}
}
