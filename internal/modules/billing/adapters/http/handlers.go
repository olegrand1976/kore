package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kore/kore/internal/modules/billing/app"
	"github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/modules/billing/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
)

func RegisterRoutes(r chi.Router, svc *app.Service, tokens *authx.TokenIssuer, authorizer authx.Authorizer, webhookSecret string) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthMiddleware(tokens))
		pr.Post("/billing/checkout-session", checkoutSession(svc, authorizer))
		pr.Post("/billing/portal-session", portalSession(svc, authorizer))
		pr.Get("/billing/subscription", getSubscription(svc, authorizer))
		pr.Post("/billing/cancel", cancelSubscription(svc, authorizer))
	})
	r.Post("/webhooks/stripe", webhookHandler(svc, webhookSecret))
}

func checkoutSession(svc ports.SubscriptionService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "billing", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			Modules    []domain.ModuleCode `json:"modules"`
			Seats      int                 `json:"seats"`
			SuccessURL string              `json:"successUrl"`
			CancelURL  string              `json:"cancelUrl"`
			Email      string              `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		session, err := svc.StartCheckout(r.Context(), ports.CheckoutCommand{
			TenantID:      identity.TenantID,
			Modules:       req.Modules,
			Seats:         req.Seats,
			SuccessURL:    req.SuccessURL,
			CancelURL:     req.CancelURL,
			CustomerEmail: req.Email,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, session)
	}
}

func portalSession(svc ports.SubscriptionService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "billing", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req struct {
			ReturnURL string `json:"returnUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		session, err := svc.OpenCustomerPortal(r.Context(), identity.TenantID, req.ReturnURL)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, session)
	}
}

func getSubscription(svc ports.SubscriptionService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "billing", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		sub, err := svc.GetSubscription(r.Context(), identity.TenantID)
		if err != nil {
			if errors.Is(err, domain.ErrSubscriptionNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, sub)
	}
}

func cancelSubscription(svc ports.SubscriptionService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "billing", authx.ActionValidate) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		if err := svc.Cancel(r.Context(), identity.TenantID); err != nil {
			if errors.Is(err, domain.ErrSubscriptionNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "canceled"})
	}
}

func webhookHandler(svc ports.SubscriptionService, _ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		signature := r.Header.Get("Stripe-Signature")
		if err := svc.HandleWebhook(r.Context(), payload, signature); err != nil {
			if errors.Is(err, domain.ErrInvalidWebhookSignature) {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "processed"})
	}
}
