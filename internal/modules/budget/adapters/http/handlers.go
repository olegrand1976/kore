package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/budget/domain"
	"github.com/kore/kore/internal/modules/budget/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(r chi.Router, budgets ports.BudgetService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/budgets", listBudgets(budgets, authorizer))
		pr.Post("/budgets", createBudget(budgets, authorizer))
		pr.Get("/budgets/{id}", getBudget(budgets))
		pr.Post("/budgets/{id}/estimates", addEstimate(budgets, authorizer))
		pr.Post("/budgets/{id}/quotes", addQuote(budgets, authorizer))
		pr.Post("/budgets/{id}/recompute", recomputeConsumption(budgets, authorizer))
		pr.Post("/budgets/{id}/approve", approveConsumption(budgets, authorizer))
	})
}

func createBudget(budgets ports.BudgetService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "budget", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ApplicationID uuid.UUID         `json:"applicationId"`
			Type          domain.BudgetType `json:"type"`
			PlannedDays   float64           `json:"plannedDays"`
			PlannedUO     float64           `json:"plannedUO"`
			PlannedAmount int64             `json:"plannedAmount"`
			Currency      string            `json:"currency"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		b, err := budgets.CreateBudget(r.Context(), ports.CreateBudgetCommand{
			TenantID:      identity.TenantID,
			ApplicationID: req.ApplicationID,
			Type:          req.Type,
			PlannedDays:   req.PlannedDays,
			PlannedUO:     req.PlannedUO,
			PlannedAmount: req.PlannedAmount,
			Currency:      req.Currency,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, b)
	}
}

func listBudgets(budgets ports.BudgetService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "budget", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := budgets.List(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func getBudget(budgets ports.BudgetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		b, err := budgets.Get(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, b)
	}
}

func addEstimate(budgets ports.BudgetService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "budget", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			DemandID   uuid.UUID `json:"demandId"`
			EffortUO   float64   `json:"effortUO"`
			EffortDays float64   `json:"effortDays"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		e, err := budgets.AddEstimate(r.Context(), ports.EstimateCommand{
			TenantID:   identity.TenantID,
			BudgetID:   budgetID,
			DemandID:   req.DemandID,
			EffortUO:   req.EffortUO,
			EffortDays: req.EffortDays,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, e)
	}
}

func addQuote(budgets ports.BudgetService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "budget", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			DemandID             uuid.UUID  `json:"demandId"`
			Amount               int64      `json:"amount"`
			EffortUO             float64    `json:"effortUO"`
			EffortDays           float64    `json:"effortDays"`
			SupersedesEstimateID *uuid.UUID `json:"supersedesEstimateId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		q, err := budgets.AddQuote(r.Context(), ports.QuoteCommand{
			TenantID:             identity.TenantID,
			BudgetID:             budgetID,
			DemandID:             req.DemandID,
			Amount:               req.Amount,
			EffortUO:             req.EffortUO,
			EffortDays:           req.EffortDays,
			SupersedesEstimateID: req.SupersedesEstimateID,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, q)
	}
}

func recomputeConsumption(budgets ports.BudgetService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "budget", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		period, err := kernel.NewPeriod(req.Start, req.End)
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		triple, err := budgets.RecomputeConsumption(r.Context(), identity.TenantID, budgetID, period)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, triple)
	}
}

func approveConsumption(budgets ports.BudgetService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "budget", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		budgetID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			Start time.Time `json:"start"`
			End   time.Time `json:"end"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		period, err := kernel.NewPeriod(req.Start, req.End)
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = budgets.Approve(r.Context(), ports.ApproveConsumptionCommand{
			TenantID:   identity.TenantID,
			BudgetID:   budgetID,
			Period:     period,
			ApprovedBy: identity.UserID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrBudgetAlreadyApproved) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "approved"})
	}
}
