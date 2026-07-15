package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func RegisterRoutes(
	r chi.Router,
	svc ports.NotificationService,
	devices ports.DeviceService,
	tokens *authx.TokenIssuer,
	authorizer authx.Authorizer,
	entitlements authx.EntitlementReader,
) {
	r.Group(func(pr chi.Router) {
		pr.Use(httpx.AuthStack(tokens, entitlements))
		pr.Get("/notification-rules", listRules(svc, authorizer))
		pr.Post("/notification-rules", defineRule(svc, authorizer))
		pr.Get("/notifications", listSent(svc, authorizer))
		pr.Post("/devices/register", registerDevice(devices))
		pr.Delete("/devices/register", unregisterDevice(devices))
	})
}

type defineRuleRequest struct {
	Code             string                 `json:"code"`
	Trigger          string                 `json:"trigger"`
	Frequency        string                 `json:"frequency"`
	RecipientsPolicy domain.RecipientPolicy `json:"recipientPolicy"`
	Template         string                 `json:"template"`
	AttachPDF        bool                   `json:"attachPdf"`
}

func defineRule(svc ports.NotificationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "notifications", authx.ActionWrite) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		var req defineRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		frequency, err := domain.ParseFrequency(req.Frequency)
		if err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		identity, _ := authx.FromContext(r.Context())
		rule := domain.NotificationRule{
			ID:               uuid.New(),
			TenantID:         identity.TenantID,
			Code:             req.Code,
			Trigger:          req.Trigger,
			Frequency:        frequency,
			RecipientsPolicy: req.RecipientsPolicy,
			Template:         req.Template,
			AttachPDF:        req.AttachPDF,
		}
		if err := svc.DefineRule(r.Context(), rule); err != nil {
			if errors.Is(err, domain.ErrInvalidFrequency) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, rule)
	}
}

func listRules(svc ports.NotificationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "notifications", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		rules, err := svc.ListRules(r.Context(), identity.TenantID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, rules)
	}
}

func listSent(svc ports.NotificationService, authorizer authx.Authorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorizer.Can(r.Context(), "notifications", authx.ActionRead) {
			httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, "forbidden")
			return
		}
		identity, _ := authx.FromContext(r.Context())
		messages, err := svc.ListSent(r.Context(), ports.SentFilter{
			TenantID: identity.TenantID,
			Limit:    100,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, messages)
	}
}

var _ = kernel.TenantID{}

func registerDevice(devices ports.DeviceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Platform string `json:"platform"`
			Token    string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		if err := devices.RegisterDevice(r.Context(), ports.RegisterDeviceCommand{
			TenantID: identity.TenantID,
			UserID:   identity.UserID,
			Platform: req.Platform,
			Token:    req.Token,
		}); err != nil {
			if errors.Is(err, domain.ErrInvalidDevicePlatform) || errors.Is(err, domain.ErrEmptyDeviceToken) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "registered"})
	}
}

func unregisterDevice(devices ports.DeviceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		if err := devices.UnregisterDevice(r.Context(), ports.UnregisterDeviceCommand{
			TenantID: identity.TenantID,
			UserID:   identity.UserID,
			Token:    req.Token,
		}); err != nil {
			httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "unregistered"})
	}
}
