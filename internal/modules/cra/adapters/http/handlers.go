package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(r chi.Router, svc ports.CRAService, tokens *authx.TokenIssuer, authorizer authx.Authorizer) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(tokens))
		pr.Get("/timesheets", getTimesheet(svc, authorizer))
		pr.Put("/timesheets/{id}/weeks/{week}", saveWeek(svc, authorizer))
		pr.Post("/timesheets/{id}/weeks/{week}/submit", submitWeek(svc, authorizer))
		pr.Put("/timesheets/{id}/commercial-info", completeCommercialInfo(svc, authorizer))
		pr.Post("/timesheets/{id}/pdf", generatePDF(svc, authorizer))
		pr.Post("/timesheets/{id}/validate", validateFinal(svc, authorizer))
	})
}

func getTimesheet(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		monthRaw := r.URL.Query().Get("month")
		if monthRaw == "" {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "month query required (YYYY-MM)")
			return
		}
		month, err := domain.ParseMonth(monthRaw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		ts, err := svc.GetOrCreate(r.Context(), identity.TenantID, identity.UserID, month)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, ts)
	}
}

func saveWeek(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, week, err := parseTimesheetWeek(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		var req struct {
			Lines []struct {
				SourceType string `json:"sourceType"`
				SourceID   string `json:"sourceId"`
				Day        string `json:"day"`
				Duration   int    `json:"duration"`
				Comment    string `json:"comment"`
			} `json:"lines"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		lines := make([]domain.TimeLine, 0, len(req.Lines))
		for _, l := range req.Lines {
			day, err := time.Parse("2006-01-02", l.Day)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid day format")
				return
			}
			lines = append(lines, domain.TimeLine{
				Source:   domain.SourceRef{Type: l.SourceType, ID: l.SourceID},
				Day:      day,
				Duration: kernel.Duration{Minutes: l.Duration},
				Comment:  l.Comment,
			})
		}
		ts, err := svc.SaveWeek(r.Context(), ports.SaveWeekCommand{
			TenantID:    identity.TenantID,
			TimesheetID: id,
			WeekNumber:  week,
			Lines:       lines,
		})
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, ts)
	}
}

func submitWeek(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, week, err := parseTimesheetWeek(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		if err := svc.SubmitWeek(r.Context(), ports.SubmitWeekCommand{
			TenantID:    identity.TenantID,
			TimesheetID: id,
			WeekNumber:  week,
			UserID:      identity.UserID,
		}); err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "submitted"})
	}
}

func completeCommercialInfo(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheet id")
			return
		}
		var info domain.CommercialInfo
		if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		if err := svc.CompleteCommercialInfo(r.Context(), ports.CommercialCommand{
			TenantID:    identity.TenantID,
			TimesheetID: id,
			Info:        info,
		}); err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func generatePDF(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheet id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		doc, err := svc.GeneratePDF(r.Context(), identity.TenantID, id)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, doc)
	}
}

func validateFinal(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheet id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		if err := svc.ValidateFinal(r.Context(), ports.ManagerValidateCommand{
			TenantID:    identity.TenantID,
			TimesheetID: id,
			ManagerID:   identity.UserID,
		}); err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "validated"})
	}
}

func parseTimesheetWeek(r *http.Request) (uuid.UUID, domain.WeekNumber, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, 0, err
	}
	weekNum, err := strconv.Atoi(chi.URLParam(r, "week"))
	if err != nil {
		return uuid.Nil, 0, err
	}
	return id, domain.WeekNumber(weekNum), nil
}

func writeCRAError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrCRAAlreadyValidated):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	case errors.Is(err, domain.ErrCommercialInfoRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrDayCapacityExceeded):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrCRAConflictAbsence):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	case errors.Is(err, domain.ErrTimesheetNotFound), errors.Is(err, domain.ErrWeekNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
