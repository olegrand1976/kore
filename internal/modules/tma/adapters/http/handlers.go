package http

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, tma ports.TMAService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/demands", listDemands(tma, authorizer))
		pr.Get("/demands/export.xml", exportXML(tma))
		pr.Get("/demands/{id}", getDemand(tma, authorizer))
		pr.Get("/demands/{id}/analysis", getAnalysis(tma, authorizer))
		pr.Post("/demands", createDemand(tma, authorizer))
		pr.Post("/demands/{id}/validate-creation", validateCreation(tma, authorizer))
		pr.Post("/demands/{id}/assign", assignDemand(tma, authorizer))
		pr.Post("/demands/{id}/take-over", takeOverDemand(tma, authorizer))
		pr.Post("/demands/{id}/analysis", addAnalysis(tma, authorizer))
		pr.Post("/demands/{id}/resolve", resolveDemand(tma, authorizer))
		pr.Post("/demands/{id}/reopen", reopenDemand(tma, authorizer))
	})
}

func listDemands(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		filter := ports.ExportFilter{
			TenantID:    identity.TenantID,
			VisibleOnly: !authorizer.Can(r.Context(), "tma", authx.ActionValidate),
		}
		if appID := r.URL.Query().Get("applicationId"); appID != "" {
			parsed, err := uuid.Parse(appID)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
				return
			}
			filter.ApplicationID = &parsed
		}
		items, err := tma.List(r.Context(), identity.TenantID, filter)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func demandVisibleTo(d domain.Demand, authorizer authx.Authorizer, r *http.Request) bool {
	if d.Visible {
		return true
	}
	return authorizer.Can(r.Context(), "tma", authx.ActionValidate)
}

func getDemand(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		d, err := tma.Get(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		if !demandVisibleTo(d, authorizer, r) {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "demand not found")
			return
		}
		httpx.WriteData(w, http.StatusOK, d)
	}
}

func getAnalysis(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		d, err := tma.Get(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		if !demandVisibleTo(d, authorizer, r) {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "demand not found")
			return
		}
		dossier, err := tma.GetAnalysis(r.Context(), identity.TenantID, id)
		if err != nil {
			if errors.Is(err, domain.ErrAnalysisNotFound) {
				httpx.WriteData(w, http.StatusOK, domain.AnalysisDossier{DemandID: id, TenantID: identity.TenantID})
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, dossier)
	}
}

func createDemand(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ApplicationID    uuid.UUID `json:"applicationId"`
			Subject          string    `json:"subject"`
			RequiresChefGate bool      `json:"requiresChefGate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		d, err := tma.CreateDemand(r.Context(), ports.CreateDemandCommand{
			TenantID:         identity.TenantID,
			ApplicationID:    req.ApplicationID,
			AuthorID:         identity.UserID,
			Subject:          req.Subject,
			RequiresChefGate: req.RequiresChefGate,
		})
		if err != nil {
			if errors.Is(err, domain.ErrDefaultBudgetRequired) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, d)
	}
}

func validateCreation(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = tma.ValidateCreation(r.Context(), ports.ChefUtilisateurCommand{
			TenantID: identity.TenantID,
			ID:       id,
			ActorID:  identity.UserID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrTransitionNotAllowed) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "validated"})
	}
}

func assignDemand(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionValidate) {
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
		err = tma.Assign(r.Context(), ports.AssignCommand{
			TenantID:   identity.TenantID,
			ID:         id,
			AssigneeID: req.AssigneeID,
			ActorID:    identity.UserID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrDemandNotVisible) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "assigned"})
	}
}

func takeOverDemand(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = tma.TakeOver(r.Context(), identity.TenantID, id, identity.UserID)
		if err != nil {
			if errors.Is(err, domain.ErrDemandNotVisible) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "taken_over"})
	}
}

func addAnalysis(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			Functional   string `json:"functional"`
			Technical    string `json:"technical"`
			Risks        string `json:"risks"`
			TestScenario string `json:"testScenario"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = tma.AddAnalysis(r.Context(), ports.AnalysisCommand{
			TenantID:     identity.TenantID,
			DemandID:     id,
			Functional:   req.Functional,
			Technical:    req.Technical,
			Risks:        req.Risks,
			TestScenario: req.TestScenario,
		})
		if err != nil {
			if errors.Is(err, domain.ErrDemandNotVisible) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, map[string]string{"status": "analysis_saved"})
	}
}

func resolveDemand(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = tma.Resolve(r.Context(), identity.TenantID, id, identity.UserID)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrDemandNotVisible):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			case errors.Is(err, domain.ErrDemandAlreadyResolved):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "resolved"})
	}
}

func reopenDemand(tma ports.TMAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "tma", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = tma.Reopen(r.Context(), ports.ReworkCommand{
			TenantID: identity.TenantID,
			ID:       id,
			Reason:   req.Reason,
			ActorID:  identity.UserID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrDemandNotVisible) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "reopened"})
	}
}

type xmlExport struct {
	Rows []domain.XmlExportRow `xml:"row"`
}

func exportXML(tma ports.TMAService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		filter := ports.ExportFilter{TenantID: identity.TenantID}
		rows, err := tma.ExportXML(r.Context(), filter)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		payload, err := xml.Marshal(xmlExport{Rows: rows})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(append([]byte(xml.Header), payload...))
	}
}
