package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/modules/conges/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, leaves ports.LeaveService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/leave-requests", listLeaveRequests(leaves))
		pr.Post("/leave-requests", createLeaveRequest(leaves, authorizer))
		pr.Post("/leave-requests/{id}/approve", approveLeaveRequest(leaves, authorizer))
		pr.Post("/leave-requests/{id}/reject", rejectLeaveRequest(leaves, authorizer))
		pr.Get("/leave-balances", listLeaveBalances(leaves))
	})
}

func listLeaveRequests(leaves ports.LeaveService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		var userID *uuid.UUID
		if uid := r.URL.Query().Get("userId"); uid != "" {
			parsed, err := uuid.Parse(uid)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid userId")
				return
			}
			userID = &parsed
		}
		var status *domain.LeaveStatus
		if st := r.URL.Query().Get("status"); st != "" {
			s := domain.LeaveStatus(st)
			status = &s
		}
		items, err := leaves.List(r.Context(), identity.TenantID, userID, status)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createLeaveRequest(leaves ports.LeaveService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "conges", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Type  domain.LeaveType `json:"type"`
			From  time.Time        `json:"from"`
			To    time.Time        `json:"to"`
			Motif string           `json:"motif"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		created, err := leaves.Request(r.Context(), ports.RequestLeaveCommand{
			TenantID: identity.TenantID,
			UserID:   identity.UserID,
			Type:     req.Type,
			From:     req.From,
			To:       req.To,
			Motif:    req.Motif,
		})
		if err != nil {
			if errors.Is(err, domain.ErrInvalidDateRange) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, created)
	}
}

func approveLeaveRequest(leaves ports.LeaveService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "conges", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = leaves.Approve(r.Context(), ports.DecideLeaveCommand{
			TenantID:  identity.TenantID,
			ID:        id,
			DecidedBy: identity.UserID,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrLeaveAlreadyDecided):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			case errors.Is(err, domain.ErrLeavePastDate):
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "approved"})
	}
}

func rejectLeaveRequest(leaves ports.LeaveService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "conges", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		err = leaves.Reject(r.Context(), ports.DecideLeaveCommand{
			TenantID:  identity.TenantID,
			ID:        id,
			DecidedBy: identity.UserID,
		})
		if err != nil {
			if errors.Is(err, domain.ErrLeaveAlreadyDecided) {
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "rejected"})
	}
}

func listLeaveBalances(leaves ports.LeaveService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identity, _ := authx.FromContext(r.Context())
		items, err := leaves.Balance(r.Context(), identity.TenantID, identity.UserID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}
