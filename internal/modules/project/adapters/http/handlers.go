package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/internal/modules/project/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, svc ports.ProjectService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/project/applications", listAgileApplications(svc, authorizer))
		pr.Get("/applications/{applicationId}/epics", listEpics(svc, authorizer))
		pr.Post("/applications/{applicationId}/epics", createEpic(svc, authorizer))
		pr.Patch("/applications/{applicationId}/epics/{id}", updateEpic(svc, authorizer))
		pr.Get("/applications/{applicationId}/sprints", listSprints(svc, authorizer))
		pr.Post("/applications/{applicationId}/sprints", createSprint(svc, authorizer))
		pr.Post("/applications/{applicationId}/sprints/{id}/start", startSprint(svc, authorizer))
		pr.Post("/applications/{applicationId}/sprints/{id}/close", closeSprint(svc, authorizer))
		pr.Post("/applications/{applicationId}/sprints/{id}/plan", planSprint(svc, authorizer))
		pr.Get("/applications/{applicationId}/backlog", listBacklog(svc, authorizer))
		pr.Patch("/applications/{applicationId}/backlog/reorder", reorderBacklog(svc, authorizer))
		pr.Get("/applications/{applicationId}/kanban-config", getKanbanConfig(svc, authorizer))
		pr.Put("/applications/{applicationId}/kanban-config", saveKanbanConfig(svc, authorizer))
		pr.Get("/applications/{applicationId}/sprints/{id}/burndown", getBurndown(svc, authorizer))
		pr.Get("/applications/{applicationId}/velocity", getVelocity(svc, authorizer))
	})
}

func canReadProject(ctx context.Context, authorizer authx.Authorizer) bool {
	return authorizer.Can(ctx, "project", authx.ActionRead)
}

func canWriteProject(ctx context.Context, authorizer authx.Authorizer) bool {
	return authorizer.Can(ctx, "project", authx.ActionWrite)
}

func canValidateProject(ctx context.Context, authorizer authx.Authorizer) bool {
	return authorizer.Can(ctx, "project", authx.ActionValidate)
}

func parseAppID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "applicationId"))
}

func listAgileApplications(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.ListAgileApplications(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func listEpics(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.ListEpics(r.Context(), identity.TenantID, appID)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createEpic(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canWriteProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Priority    string `json:"priority"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.CreateEpic(r.Context(), ports.CreateEpicCommand{
			TenantID:      identity.TenantID,
			ApplicationID: appID,
			Title:         req.Title,
			Description:   req.Description,
			Priority:      req.Priority,
		})
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusCreated, item)
	}
}

func updateEpic(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canWriteProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		epicID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid epic id")
			return
		}
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		cmd := ports.UpdateEpicCommand{
			TenantID:      authx.MustFromContext(r.Context()).TenantID,
			ApplicationID: appID,
			EpicID:        epicID,
		}
		if v, ok := raw["title"]; ok {
			var s string
			_ = json.Unmarshal(v, &s)
			cmd.Title = &s
		}
		if v, ok := raw["description"]; ok {
			var s string
			_ = json.Unmarshal(v, &s)
			cmd.Description = &s
		}
		if v, ok := raw["status"]; ok {
			var s string
			_ = json.Unmarshal(v, &s)
			st := domain.EpicStatus(s)
			cmd.Status = &st
		}
		if v, ok := raw["priority"]; ok {
			var s string
			_ = json.Unmarshal(v, &s)
			cmd.Priority = &s
		}
		if v, ok := raw["targetSprintId"]; ok {
			var rawID *string
			if err := json.Unmarshal(v, &rawID); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid targetSprintId")
				return
			}
			if rawID == nil || *rawID == "" {
				var nilID *uuid.UUID
				cmd.TargetSprint = &nilID
			} else {
				parsed, err := uuid.Parse(*rawID)
				if err != nil {
					httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid targetSprintId")
					return
				}
				id := parsed
				idPtr := &id
				cmd.TargetSprint = &idPtr
			}
		}
		item, err := svc.UpdateEpic(r.Context(), cmd)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func listSprints(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.ListSprints(r.Context(), identity.TenantID, appID)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createSprint(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canWriteProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		var req struct {
			Name           string `json:"name"`
			Goal           string `json:"goal"`
			StartDate      string `json:"startDate"`
			EndDate        string `json:"endDate"`
			CapacityPoints *int16 `json:"capacityPoints"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid startDate")
			return
		}
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid endDate")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.CreateSprint(r.Context(), ports.CreateSprintCommand{
			TenantID:       identity.TenantID,
			ApplicationID:  appID,
			Name:           req.Name,
			Goal:           req.Goal,
			StartDate:      start,
			EndDate:        end,
			CapacityPoints: req.CapacityPoints,
		})
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusCreated, item)
	}
}

func startSprint(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canValidateProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, sprintID, ok := parseAppAndSprint(r, w)
		if !ok {
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.StartSprint(r.Context(), identity.TenantID, appID, sprintID)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func closeSprint(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canValidateProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, sprintID, ok := parseAppAndSprint(r, w)
		if !ok {
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.CloseSprint(r.Context(), identity.TenantID, appID, sprintID)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func planSprint(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canWriteProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, sprintID, ok := parseAppAndSprint(r, w)
		if !ok {
			return
		}
		var req struct {
			DemandIDs []uuid.UUID `json:"demandIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err := svc.PlanSprint(r.Context(), ports.PlanSprintCommand{
			TenantID:      identity.TenantID,
			ApplicationID: appID,
			SprintID:      sprintID,
			DemandIDs:     req.DemandIDs,
		})
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func listBacklog(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		backlogOnly := r.URL.Query().Get("backlogOnly") == "true"
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.ListBacklog(r.Context(), identity.TenantID, appID, backlogOnly)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func reorderBacklog(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canWriteProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		var req struct {
			DemandIDs []uuid.UUID `json:"demandIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = svc.ReorderBacklog(r.Context(), ports.ReorderBacklogCommand{
			TenantID:      identity.TenantID,
			ApplicationID: appID,
			DemandIDs:     req.DemandIDs,
		})
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func getKanbanConfig(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.GetKanbanConfig(r.Context(), identity.TenantID, appID)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func saveKanbanConfig(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canWriteProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		var req struct {
			Columns []domain.KanbanColumn `json:"columns"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.SaveKanbanConfig(r.Context(), ports.UpdateKanbanConfigCommand{
			TenantID:      identity.TenantID,
			ApplicationID: appID,
			Columns:       req.Columns,
		})
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func getBurndown(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, sprintID, ok := parseAppAndSprint(r, w)
		if !ok {
			return
		}
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.GetSprintBurndown(r.Context(), identity.TenantID, appID, sprintID)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func getVelocity(svc ports.ProjectService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !canReadProject(r.Context(), authorizer) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		appID, err := parseAppID(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
			return
		}
		lastN := 3
		identity, _ := authx.FromContext(r.Context())
		item, err := svc.GetVelocity(r.Context(), identity.TenantID, appID, lastN)
		if err != nil {
			writeProjectError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, item)
	}
}

func parseAppAndSprint(r *http.Request, w http.ResponseWriter) (uuid.UUID, uuid.UUID, bool) {
	appID, err := parseAppID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid applicationId")
		return uuid.Nil, uuid.Nil, false
	}
	sprintID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid sprint id")
		return uuid.Nil, uuid.Nil, false
	}
	return appID, sprintID, true
}

func writeProjectError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEpicNotFound), errors.Is(err, domain.ErrSprintNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	case errors.Is(err, domain.ErrApplicationNotAgile),
		errors.Is(err, domain.ErrSprintNotPlanned),
		errors.Is(err, domain.ErrSprintNotActive),
		errors.Is(err, domain.ErrActiveSprintExists):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
