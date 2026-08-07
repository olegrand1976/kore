package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(r chi.Router, svc ports.ReportingService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/gantt", getGantt(svc, authorizer))
		pr.Get("/planning", getPlanning(svc, authorizer))
		pr.Get("/dashboards/home", getHomeDashboard(svc, authorizer, entitlements))
		pr.Get("/dashboards/{code}", getDashboard(svc, authorizer))
		pr.Post("/reports/run", runReport(svc, authorizer))
		pr.Get("/billing-stats", getBillingStats(svc, authorizer))
	})
}

func getGantt(svc ports.ReportingService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "reporting", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		period, err := parsePeriodQuery(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		view, err := svc.GetGantt(r.Context(), ports.GanttQuery{TenantID: identity.TenantID, Period: period})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, view)
	}
}

func getPlanning(svc ports.ReportingService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "reporting", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		period, err := parsePeriodQuery(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		view, err := svc.GetPlanning(r.Context(), ports.PlanningQuery{TenantID: identity.TenantID, Period: period})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, view)
	}
}

func getDashboard(svc ports.ReportingService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "reporting", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		code := chi.URLParam(r, "code")
		identity, _ := authx.FromContext(r.Context())
		dash, err := svc.GetDashboard(r.Context(), identity.TenantID, code)
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, dash)
	}
}

func getHomeDashboard(svc ports.ReportingService, authorizer authx.Authorizer, entitlements authx.EntitlementReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		moduleOn := func(code authx.Module) bool {
			if entitlements == nil {
				return true
			}
			enabled, err := entitlements.IsModuleEnabled(r.Context(), identity.TenantID, code)
			return err == nil && enabled
		}
		canRead := func(code authx.Module) bool {
			return moduleOn(code) && authorizer.Can(r.Context(), code, authx.ActionRead)
		}
		q := ports.HomeDashboardQuery{
			TenantID:         identity.TenantID,
			UserID:           identity.UserID,
			IncludeCRA:       moduleOn("cra"),
			IncludeLeave:     canRead("conges"),
			IncludeTMA:       canRead("tma"),
			IncludeBudget:    canRead("budget"),
			IncludeBilling:   moduleOn("cra") && (authorizer.Can(r.Context(), "reporting", authx.ActionRead) || authorizer.Can(r.Context(), "cra", authx.ActionValidate)),
			CanValidateLeave: authorizer.Can(r.Context(), "conges", authx.ActionValidate),
		}
		home, err := svc.GetHomeDashboard(r.Context(), q)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, home)
	}
}

func runReport(svc ports.ReportingService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "reporting", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ReportCode string         `json:"reportCode"`
			Params     map[string]any `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		result, err := svc.RunReport(r.Context(), ports.RunReportCommand{
			TenantID:   identity.TenantID,
			ReportCode: req.ReportCode,
			Params:     req.Params,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

func getBillingStats(svc ports.ReportingService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "reporting", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		period, err := parsePeriodQuery(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		stats, err := svc.GetBillingStats(r.Context(), ports.BillingStatsQuery{
			TenantID: identity.TenantID,
			Period:   period,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, stats)
	}
}

func parsePeriodQuery(r *http.Request) (kernel.Period, error) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	windowStr := r.URL.Query().Get("window")
	if startStr == "" && endStr == "" && windowStr == "60" {
		now := time.Now().UTC()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 59)
		return kernel.NewPeriod(start, end)
	}
	if startStr == "" || endStr == "" {
		now := time.Now().UTC()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, -1)
		return kernel.NewPeriod(start, end)
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return kernel.Period{}, err
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return kernel.Period{}, err
	}
	return kernel.NewPeriod(start, end)
}
