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

func RegisterRoutes(r chi.Router, svc ports.CRAService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader, invoicingEnabled kernel.InvoicingEnabledReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/timesheets/recent", listTimesheets(svc, authorizer))
		pr.Get("/timesheets", getTimesheet(svc, authorizer))
		pr.Get("/timesheets/{id}", getTimesheetByID(svc, authorizer))
		pr.Delete("/timesheets/{id}", deleteTimesheet(svc, authorizer))
		pr.Put("/timesheets/{id}/weeks/{week}", saveWeek(svc, authorizer))
		pr.Post("/timesheets/{id}/weeks/{week}/submit", submitWeek(svc, authorizer))
		pr.Put("/timesheets/{id}/commercial-info", completeCommercialInfo(svc, authorizer))
		pr.Post("/timesheets/{id}/pdf", generatePDF(svc, authorizer))
		pr.Post("/timesheets/{id}/validate", validateFinal(svc, authorizer))
		pr.Post("/timesheets/{id}/reject", rejectTimesheet(svc, authorizer))
		pr.Post("/timesheets/{id}/unvalidate", unvalidateTimesheet(svc, authorizer))
		pr.Get("/prestations", listPrestations(svc, authorizer))
		pr.Get("/prestations/export.xml", exportPrestationsXML(svc, authorizer))
		pr.Get("/prestations/billable-summary", billableSummary(svc, authorizer))
		pr.Post("/prestations/validate-all", validateAllPrestations(svc, authorizer))
		pr.Post("/prestations/create-invoices", createInvoicesFromPrestations(svc, authorizer, invoicingEnabled))
		pr.Post("/prestations/preview-invoices", previewInvoicesFromPrestations(svc, authorizer, invoicingEnabled))
		pr.Post("/timesheets/{id}/prefill-holidays", prefillHolidays(svc, authorizer))
		pr.Post("/timesheets/{id}/prefill-ett", prefillETT(svc, authorizer))
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
				ID          string `json:"id"`
				SourceType  string `json:"sourceType"`
				SourceID    string `json:"sourceId"`
				Day         string `json:"day"`
				Duration    int    `json:"duration"`
				Comment     string `json:"comment"`
				Origin      string `json:"origin"`
				Billable    *bool  `json:"billable"`
				WorkRefType string `json:"workRefType"`
				WorkRefID   string `json:"workRefId"`
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
			// Carry the client-side identity through: the service keeps a line ID only
			// when it already belongs to the week being saved, so stale or forged IDs
			// simply yield a fresh one. Without it every save regenerates all IDs,
			// which remounts the grid rows and drops the caret mid-edit.
			lineID, err := uuid.Parse(l.ID)
			if err != nil {
				lineID = uuid.Nil
			}
			origin := domain.LineOrigin(l.Origin)
			if origin != domain.OriginPrefill && origin != domain.OriginManual {
				origin = ""
			}
			lines = append(lines, domain.TimeLine{
				ID:          lineID,
				Source:      domain.SourceRef{Type: l.SourceType, ID: l.SourceID},
				Day:         day,
				Duration:    kernel.Duration{Minutes: l.Duration},
				Comment:     l.Comment,
				Origin:      origin,
				Billable:    billable,
				WorkRefType: l.WorkRefType,
				WorkRefID:   l.WorkRefID,
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
		result, err := svc.ValidateFinal(r.Context(), ports.ManagerValidateCommand{
			TenantID:    identity.TenantID,
			TimesheetID: id,
			ManagerID:   identity.UserID,
		})
		if err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"status":       "validated",
			"invoiceDraft": result.InvoiceDraft,
		})
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

func deleteTimesheet(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok || identity.Profile != authx.ProfileAdmin {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		if !authorizer.Can(r.Context(), "cra", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheet id")
			return
		}
		if err := svc.DeleteTimesheet(r.Context(), identity.TenantID, id); err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func unvalidateTimesheet(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authx.FromContext(r.Context())
		if !ok || identity.Profile != authx.ProfileAdmin {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		if !authorizer.Can(r.Context(), "cra", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheet id")
			return
		}
		if err := svc.UnvalidateTimesheet(r.Context(), identity.TenantID, id); err != nil {
			writeCRAError(w, err)
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "unvalidated"})
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

const maxInvoiceBatch = 100

func previewInvoicesFromPrestations(svc ports.CRAService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		if !authorizer.Can(r.Context(), "cra", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			TimesheetIDs []string `json:"timesheetIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if len(req.TimesheetIDs) == 0 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "timesheetIds required")
			return
		}
		if len(req.TimesheetIDs) > maxInvoiceBatch {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "timesheetIds max 100")
			return
		}
		ids, err := parseTimesheetIDList(req.TimesheetIDs)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		previews, err := svc.PreviewInvoicesFromTimesheets(r.Context(), identity.TenantID, ids)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{"previews": previews})
	}
}

func createInvoicesFromPrestations(svc ports.CRAService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		if !authorizer.Can(r.Context(), "cra", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			TimesheetIDs []string `json:"timesheetIds"`
			Items        []struct {
				TimesheetID    string   `json:"timesheetId"`
				ClientID       *string  `json:"clientId"`
				BillableHours  *float64 `json:"billableHours"`
				UnitPriceCents *int64   `json:"unitPriceCents"`
				TaxRate        *float64 `json:"taxRate"`
				Currency       *string  `json:"currency"`
				Description    *string  `json:"description"`
				MissionLabel   *string  `json:"missionLabel"`
			} `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())

		var outcomes []ports.InvoiceDraftOutcome
		var err error
		switch {
		case len(req.Items) > 0:
			if len(req.Items) > maxInvoiceBatch {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "items max 100")
				return
			}
			items := make([]ports.CreateInvoiceFromTimesheetItem, 0, len(req.Items))
			for _, raw := range req.Items {
				id, parseErr := uuid.Parse(raw.TimesheetID)
				if parseErr != nil {
					httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid timesheetId")
					return
				}
				item := ports.CreateInvoiceFromTimesheetItem{TimesheetID: id}
				if raw.ClientID != nil {
					cid, parseErr := uuid.Parse(*raw.ClientID)
					if parseErr != nil {
						httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid clientId")
						return
					}
					item.ClientID = &cid
				}
				item.BillableHours = raw.BillableHours
				item.UnitPriceCents = raw.UnitPriceCents
				item.TaxRate = raw.TaxRate
				item.Currency = raw.Currency
				item.Description = raw.Description
				item.MissionLabel = raw.MissionLabel
				items = append(items, item)
			}
			outcomes, err = svc.CreateInvoicesFromTimesheetItems(r.Context(), identity.TenantID, items)
		case len(req.TimesheetIDs) > 0:
			if len(req.TimesheetIDs) > maxInvoiceBatch {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "timesheetIds max 100")
				return
			}
			ids, parseErr := parseTimesheetIDList(req.TimesheetIDs)
			if parseErr != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, parseErr.Error())
				return
			}
			outcomes, err = svc.CreateInvoicesFromTimesheets(r.Context(), identity.TenantID, ids)
		default:
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "timesheetIds or items required")
			return
		}
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		created := 0
		for _, o := range outcomes {
			if o.Status == ports.InvoiceDraftCreated {
				created++
			}
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"created":  created,
			"outcomes": outcomes,
		})
	}
}

func parseTimesheetIDList(rawIDs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid timesheetId")
		}
		ids = append(ids, id)
	}
	return ids, nil
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

func prefillETT(svc ports.CRAService, authorizer authx.Authorizer) http.HandlerFunc {
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
		added, err := svc.PrefillFromETT(r.Context(), identity.TenantID, ts.UserID, ts.Month)
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
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeCRAAlreadyValidated, err.Error())
	case errors.Is(err, domain.ErrCRAAlreadyInvoiced):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeCRAAlreadyInvoiced, err.Error())
	case errors.Is(err, domain.ErrCRANotFinal):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
	case errors.Is(err, domain.ErrCommercialInfoRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeCommercialInfoRequired, err.Error())
	case errors.Is(err, domain.ErrDayCapacityExceeded):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeDayCapacityExceeded, err.Error())
	case errors.Is(err, domain.ErrCRAConflictAbsence):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeCRAConflictAbsence, err.Error())
	case errors.Is(err, domain.ErrWeekIncomplete):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeWeekIncomplete, err.Error())
	case errors.Is(err, domain.ErrCRANotSubmitted):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeCRANotSubmitted, err.Error())
	case errors.Is(err, domain.ErrTimesheetNotFound), errors.Is(err, domain.ErrWeekNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
	}
}
