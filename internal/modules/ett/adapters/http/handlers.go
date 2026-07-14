package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ett/domain"
	"github.com/kore/kore/internal/modules/ett/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, svc ports.ETTService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Post("/ett/clock-in", clockIn(svc, authorizer))
		pr.Post("/ett/clock-out", clockOut(svc, authorizer))
		pr.Get("/ett/records", listRecords(svc, authorizer))
		pr.Post("/ett/records/{id}/correct", correctRecord(svc, authorizer))
		pr.Get("/ett/records/{id}/audit", getAuditTrail(svc, authorizer))
		pr.Get("/ett/reconciliation", compareCRA(svc, authorizer))
	})
}

func clockIn(svc ports.ETTService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ett", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			At time.Time `json:"at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if req.At.IsZero() {
			req.At = time.Now().UTC()
		}
		identity, _ := authx.FromContext(r.Context())
		rec, err := svc.ClockIn(r.Context(), ports.ClockInCommand{
			TenantID: identity.TenantID,
			UserID:   identity.UserID,
			At:       req.At,
		})
		if err != nil {
			if errors.Is(err, domain.ErrNotSalarieETT) {
				httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, rec)
	}
}

func clockOut(svc ports.ETTService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ett", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			At time.Time `json:"at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if req.At.IsZero() {
			req.At = time.Now().UTC()
		}
		identity, _ := authx.FromContext(r.Context())
		rec, err := svc.ClockOut(r.Context(), ports.ClockOutCommand{
			TenantID: identity.TenantID,
			UserID:   identity.UserID,
			At:       req.At,
		})
		if err != nil {
			if errors.Is(err, domain.ErrNotSalarieETT) {
				httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, rec)
	}
}

func listRecords(svc ports.ETTService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ett", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		q := ports.RecordsQuery{TenantID: identity.TenantID}
		if uid := r.URL.Query().Get("userId"); uid != "" {
			id, err := uuid.Parse(uid)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid userId")
				return
			}
			q.UserID = &id
		}
		items, err := svc.ListRecords(r.Context(), q)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func correctRecord(svc ports.ETTService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ett", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		var req struct {
			ClockIn  *time.Time `json:"clockIn"`
			ClockOut *time.Time `json:"clockOut"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		rec, err := svc.CorrectRecord(r.Context(), ports.CorrectRecordCommand{
			TenantID: identity.TenantID,
			RecordID: id,
			ActorID:  identity.UserID,
			ClockIn:  req.ClockIn,
			ClockOut: req.ClockOut,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, rec)
	}
}

func getAuditTrail(svc ports.ETTService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ett", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.GetAuditTrail(r.Context(), identity.TenantID, id)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func compareCRA(svc ports.ETTService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "ett", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		month := r.URL.Query().Get("month")
		if month == "" {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "month query required (YYYY-MM)")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		userID := identity.UserID
		if raw := r.URL.Query().Get("userId"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid userId")
				return
			}
			if !authorizer.Can(r.Context(), "ett", authx.ActionValidate) && id != identity.UserID {
				httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
				return
			}
			userID = id
		}
		report, err := svc.CompareCRA(r.Context(), identity.TenantID, userID, month)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, report)
	}
}
