package http

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/adapters/pdp"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

const (
	proformaRateLimitWindow = time.Minute
	proformaRateLimitMax    = 30
)

func RegisterRoutes(
	r chi.Router,
	svc ports.InvoicingService,
	tokens *authx.TokenIssuer,
	authorizer authx.Authorizer,
	entitlements authx.EntitlementReader,
	invoicingEnabled kernel.InvoicingEnabledReader,
	pdpWebhookSecret string,
	publicBaseURL string,
	appCache cache.Cache,
	keys cache.KeyBuilder,
) {
	r.Post("/webhooks/pdp", pdpWebhook(svc, pdpWebhookSecret))
	r.With(proformaRateLimit(appCache, keys, "get")).Get("/public/proforma/{token}", getProformaPublic(svc))
	r.With(proformaRateLimit(appCache, keys, "validate")).Post("/public/proforma/{token}/validate", validateProformaPublic(svc))
	r.With(proformaRateLimit(appCache, keys, "reject")).Post("/public/proforma/{token}/reject", rejectProformaPublic(svc))
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/invoices", listInvoices(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices", createInvoice(svc, authorizer, invoicingEnabled))
		pr.Get("/invoices/{id}", getInvoice(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices/compute-virtual", computeVirtual(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices/{id}/transmit", transmitInvoice(svc, authorizer, invoicingEnabled))
		pr.Post("/invoices/{id}/emit-proforma", emitProforma(svc, authorizer, invoicingEnabled, publicBaseURL))
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
	case errors.Is(err, domain.ErrInvalidInvoiceState),
		errors.Is(err, domain.ErrProformaAlreadyValidated),
		errors.Is(err, domain.ErrProformaAlreadyRejected),
		errors.Is(err, domain.ErrProformaConflict):
		httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
		return true
	case errors.Is(err, domain.ErrProformaTokenInvalid):
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
		return true
	case errors.Is(err, domain.ErrProformaTokenExpired):
		httpx.WriteError(w, http.StatusGone, httpx.ErrCodeConflict, err.Error())
		return true
	case errors.Is(err, domain.ErrZeroUnitPrice),
		errors.Is(err, domain.ErrNoBillableContent),
		errors.Is(err, domain.ErrInvalidInvoiceLine),
		errors.Is(err, domain.ErrNoClientEmail),
		errors.Is(err, domain.ErrProformaCommentRequired),
		errors.Is(err, domain.ErrProformaCommentTooLong):
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

func emitProforma(svc ports.InvoicingService, authorizer authx.Authorizer, invoicingEnabled kernel.InvoicingEnabledReader, publicBaseURL string) http.HandlerFunc {
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
		var req struct {
			RecipientEmail string `json:"recipientEmail"`
		}
		if r.Body != nil && r.ContentLength != 0 {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		identity, _ := authx.FromContext(r.Context())
		inv, err := svc.EmitProforma(r.Context(), ports.EmitProformaCommand{
			TenantID:       identity.TenantID,
			InvoiceID:      id,
			ActorID:        identity.UserID,
			RecipientEmail: req.RecipientEmail,
			PublicBaseURL:  publicBaseURL,
		})
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, inv)
	}
}

func getProformaPublic(svc ports.InvoicingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		preview, err := svc.GetProformaByToken(r.Context(), token)
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, preview)
	}
}

func validateProformaPublic(svc ports.InvoicingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		var req struct {
			Comment string `json:"comment"`
		}
		if r.Body != nil && r.ContentLength != 0 {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		preview, err := svc.ValidateProformaByToken(r.Context(), ports.ProformaDecisionCommand{
			Token:   token,
			Comment: req.Comment,
		})
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, preview)
	}
}

func rejectProformaPublic(svc ports.InvoicingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		var req struct {
			Comment string `json:"comment"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		preview, err := svc.RejectProformaByToken(r.Context(), ports.ProformaDecisionCommand{
			Token:   token,
			Comment: req.Comment,
		})
		if err != nil {
			if writeInvoicingErr(w, err) {
				return
			}
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, preview)
	}
}

func proformaRateLimit(appCache cache.Cache, keys cache.KeyBuilder, scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appCache == nil || keys == nil {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			key := keys.PublicKey("invoicing", "ratelimit", "proforma", scope, ip)
			var count int
			found, err := appCache.Get(r.Context(), key, &count)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if found && count >= proformaRateLimitMax {
				httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrCodeTooManyRequests, "too many requests")
				return
			}
			count++
			_ = appCache.Set(r.Context(), key, count, proformaRateLimitWindow)
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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
