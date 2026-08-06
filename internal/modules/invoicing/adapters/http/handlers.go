package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/adapters/pdp"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(r chi.Router, svc ports.InvoicingService, tokens *authx.TokenIssuer, authorizer authx.Authorizer, entitlements authx.EntitlementReader, invoicingEnabled kernel.InvoicingEnabledReader, pdpWebhookSecret string) {
	r.Post("/webhooks/pdp", pdpWebhook(svc, pdpWebhookSecret))
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/invoices", listInvoices(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices", createInvoice(svc, authorizer, invoicingEnabled))
		pr.Get("/invoices/{id}", getInvoice(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices/compute-virtual", computeVirtual(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices/{id}/transmit", transmitInvoice(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices/{id}/credit-note", createCreditNote(svc, authorizer, invoicingEnabled))
	})
}

func writeInvoicingErr(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, domain.ErrInvoicingDisabled):
		httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
		return true
	case errors.Is(err, domain.ErrAlreadyInvoiced):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
		return true
	case errors.Is(err, domain.ErrInvoiceNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
		return true
	case errors.Is(err, domain.ErrInvalidInvoiceState):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
		return true
	case errors.Is(err, domain.ErrZeroUnitPrice),
		errors.Is(err, domain.ErrNoBillableContent),
		errors.Is(err, domain.ErrInvalidInvoiceLine):
		httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
		return true
	default:
		return false
	}
}

func listInvoices(svc ports.InvoicingService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		items, err := svc.List(r.Context(), identity.TenantID)
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, items)
	}
}

func createInvoice(svc ports.InvoicingService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ClientID uuid.UUID                `json:"clientId"`
			Type     domain.InvoiceType       `json:"type"`
			Currency string                   `json:"currency"`
			Lines    []ports.InvoiceLineInput `json:"lines"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if req.ClientID == uuid.Nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "clientId required")
			return
		}
		if len(req.Lines) == 0 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "at least one line required")
			return
		}
		if req.Type == "" {
			req.Type = domain.InvoiceTypeStandard
		}
		identity, _ := authx.FromContext(r.Context())
		inv, err := svc.Create(r.Context(), ports.CreateInvoiceCommand{
			TenantID: identity.TenantID,
			ClientID: req.ClientID,
			Type:     req.Type,
			Currency: req.Currency,
			Lines:    req.Lines,
		})
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, inv)
	}
}

func getInvoice(svc ports.InvoicingService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		inv, err := svc.Get(r.Context(), identity.TenantID, id)
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, inv)
	}
}

func computeVirtual(svc ports.InvoicingService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ClientID  uuid.UUID                `json:"clientId"`
			MissionID *uuid.UUID               `json:"missionId"`
			Start     time.Time                `json:"start"`
			End       time.Time                `json:"end"`
			Lines     []ports.InvoiceLineInput `json:"lines"`
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
		inv, err := svc.ComputeVirtual(r.Context(), ports.ComputeVirtualCommand{
			TenantID:  identity.TenantID,
			ClientID:  req.ClientID,
			MissionID: req.MissionID,
			Period:    period,
			Lines:     req.Lines,
		})
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, inv)
	}
}

func transmitInvoice(svc ports.InvoicingService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		inv, err := svc.Transmit(r.Context(), identity.TenantID, id)
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, inv)
	}
}

func createCreditNote(svc ports.InvoicingService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireInvoicingEnabled(w, r, invoicingEnabled) {
			return
		}
		if !authorizer.Can(r.Context(), "invoicing", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid id")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		cn, err := svc.CreateCreditNote(r.Context(), identity.TenantID, id)
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, cn)
	}
}

func pdpWebhook(svc ports.InvoicingService, webhookSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if !pdp.VerifyWebhook(body, r.Header.Get("X-PDP-Signature"), webhookSecret) {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeForbidden, "invalid signature")
			return
		}
		var req struct {
			TenantID  string `json:"tenantId"`
			InvoiceID string `json:"invoiceId"`
			ReceiptID string `json:"receiptId"`
			Status    string `json:"status"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		tenantUUID, err := uuid.Parse(req.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid tenantId")
			return
		}
		invoiceID, err := uuid.Parse(req.InvoiceID)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid invoiceId")
			return
		}
		if err := svc.SyncPDPStatus(r.Context(), ports.PDPStatusEvent{
			TenantID:  kernel.NewTenantID(tenantUUID),
			InvoiceID: invoiceID,
			ReceiptID: req.ReceiptID,
			Status:    domain.InvoiceStatus(req.Status),
		}); err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "synced"})
	}
}
