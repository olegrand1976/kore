package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, svc ports.WorkflowService, tokens *authx.TokenIssuer, authorizer authx.Authorizer) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(tokens))
		pr.Post("/workflows", defineWorkflow(svc, authorizer))
		pr.Get("/workflows/{code}", getDefinition(svc, authorizer))
		pr.Get("/workflow-instances/{id}", getInstance(svc))
		pr.Get("/workflow-instances/{id}/actions", availableActions(svc))
		pr.Post("/workflow-instances/{id}/fire", fireTransition(svc))
	})
}

func defineWorkflow(svc ports.WorkflowService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "workflow", authx.ActionWrite) &&
			!authorizer.Can(r.Context(), "org", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Code        string              `json:"code"`
			EntityType  string              `json:"entityType"`
			States      []domain.State      `json:"states"`
			Transitions []domain.Transition `json:"transitions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		def := domain.WorkflowDefinition{
			ID:          uuid.New(),
			TenantID:    identity.TenantID,
			Code:        req.Code,
			EntityType:  req.EntityType,
			States:      req.States,
			Transitions: req.Transitions,
		}
		if err := svc.DefineWorkflow(r.Context(), def); err != nil {
			writeWorkflowError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusCreated, map[string]string{"code": def.Code})
	}
}

func getDefinition(svc ports.WorkflowService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "workflow", authx.ActionRead) &&
			!authorizer.Can(r.Context(), "org", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		code := chi.URLParam(r, "code")
		identity, _ := authx.FromContext(r.Context())
		def, err := svc.GetDefinition(r.Context(), identity.TenantID, code)
		if err != nil {
			writeWorkflowError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, def)
	}
}

func getInstance(svc ports.WorkflowService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid instance id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		inst, err := svc.GetInstance(r.Context(), identity.TenantID, id)
		if err != nil {
			writeWorkflowError(w, err)
			return
		}
		history, err := svc.History(r.Context(), identity.TenantID, id)
		if err != nil {
			writeWorkflowError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"instance": inst,
			"history":  history,
		})
	}
}

func availableActions(svc ports.WorkflowService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid instance id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		actions, err := svc.AvailableActions(r.Context(), identity.TenantID, id, identity)
		if err != nil {
			writeWorkflowError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, actions)
	}
}

func fireTransition(svc ports.WorkflowService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid instance id")
			return
		}
		var req struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		inst, err := svc.Fire(r.Context(), ports.FireTransitionCommand{
			TenantID:   identity.TenantID,
			InstanceID: id,
			Action:     domain.ActionCode(req.Action),
			Actor:      identity,
		})
		if err != nil {
			writeWorkflowError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, inst)
	}
}

func writeWorkflowError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrTransitionNotAllowed):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	case errors.Is(err, domain.ErrGuardFailed):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrActionNotPermitted):
		httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
	case errors.Is(err, domain.ErrWorkflowNotFound), errors.Is(err, domain.ErrInstanceNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidDefinition):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
