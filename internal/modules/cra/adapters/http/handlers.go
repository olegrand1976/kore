package http

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(r chi.Router, svc ports.CRAService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/timesheets/recent", listTimesheets(svc, authorizer))
		pr.Get("/timesheets", getTimesheet(svc, authorizer))
		pr.Get("/timesheets/{id}", getTimesheetByID(svc, authorizer))
		pr.Put("/timesheets/{id}/weeks/{week}", saveWeek(svc, authorizer))
		pr.Post("/timesheets/{id}/weeks/{week}/submit", submitWeek(svc, authorizer))
		pr.Put("/timesheets/{id}/commercial-info", completeCommercialInfo(svc, authorizer))
		pr.Post("/timesheets/{id}/pdf", generatePDF(svc, authorizer))
		pr.Post("/timesheets/{id}/validate", validateFinal(svc, authorizer))
		pr.Post("/timesheets/{id}/reject", rejectTimesheet(svc, authorizer))
		pr.Get("/prestations", listPrestations(svc, authorizer))
		pr.Get("/prestations/export.xml", exportPrestationsXML(svc, authorizer))
		pr.Get("/prestations/billable-summary", billableSummary(svc, authorizer))
		pr.Post("/prestations/validate-all", validateAllPrestations(svc, authorizer))
		pr.Post("/timesheets/{id}/prefill-holidays", prefillHolidays(svc, authorizer))
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
				Billable   *bool  `json:"billable"`
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
			billable := true
			if l.Billable != nil {
				billable = *l.Billable
			}
			lines = append(lines, domain.TimeLine{
				Source:   domain.SourceRef{Type: l.SourceType, ID: l.SourceID},
				Day:      day,
				Duration: kernel.Duration{Minutes: l.Duration},
				Comment:  l.Comment,
				Billable: billable,
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

func getTimesheetByID(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheet id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		ts, err := svc.GetByID(r.Context(), identity.TenantID, id)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		if !canAccessTimesheet(r.Context(), authorizer, identity, ts) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		httpx.WriteData(w, http.StatusOK, ts)
	}
}

func listTimesheets(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		limit := 24
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 48 {
				limit = n
			}
		}
		identity, _ := authx.FromContext(r.Context())
		managerView := authorizer.Can(r.Context(), "cra", authx.ActionValidate)
		items, err := svc.ListTimesheetSummaries(r.Context(), identity.TenantID, identity.UserID, managerView, limit)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func canAccessTimesheet(ctx context.Context, authorizer authx.Authorizer, identity authx.Identity, ts domain.Timesheet) bool {
	if authorizer.Can(ctx, "cra", authx.ActionValidate) {
		return true
	}
	return ts.UserID == identity.UserID
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
		ts, err := svc.GetByID(r.Context(), identity.TenantID, id)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		if !canAccessTimesheet(r.Context(), authorizer, identity, ts) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		doc, err := svc.GeneratePDF(r.Context(), identity.TenantID, id)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		w.Header().Set("Content-Type", doc.MimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.Filename))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(doc.Content)
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

func rejectTimesheet(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
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
		var req struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		identity, _ := authx.FromContext(r.Context())
		if err := svc.RejectTimesheet(r.Context(), ports.RejectTimesheetCommand{
			TenantID: identity.TenantID, TimesheetID: id, ManagerID: identity.UserID, Reason: req.Reason,
		}); err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "rejected"})
	}
}

func listPrestations(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionValidate) {
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
		items, err := svc.ListPrestations(r.Context(), identity.TenantID, month)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func validateAllPrestations(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Month string `json:"month"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		month, err := domain.ParseMonth(req.Month)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		result, err := svc.ValidateAll(r.Context(), ports.ValidateAllCommand{
			TenantID: identity.TenantID, ManagerID: identity.UserID, Month: month,
		})
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, result)
	}
}

type prestationsXMLExport struct {
	Rows []ports.PrestationExportRow `xml:"row"`
}

func exportPrestationsXML(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionValidate) {
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
		rows, err := svc.ExportPrestationsXML(r.Context(), identity.TenantID, month)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		payload, err := xml.Marshal(prestationsXMLExport{Rows: rows})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(append([]byte(xml.Header), payload...))
	}
}

func billableSummary(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "cra", authx.ActionValidate) {
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
		items, err := svc.BillableSummary(r.Context(), identity.TenantID, month)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func prefillHolidays(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
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
		ts, err := svc.GetByID(r.Context(), identity.TenantID, id)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		if !canAccessTimesheet(r.Context(), authorizer, identity, ts) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		country := strings.TrimSpace(r.URL.Query().Get("country"))
		if country == "" {
			country = "FR"
		}
		added, err := svc.PrefillPublicHolidays(r.Context(), identity.TenantID, ts.UserID, ts.Month, country)
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]int{"added": added})
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
	case errors.Is(err, domain.ErrWeekIncomplete):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
	case errors.Is(err, domain.ErrTimesheetNotFound), errors.Is(err, domain.ErrWeekNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
