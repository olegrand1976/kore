package http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/publicsite/app"
	"github.com/kore/kore/internal/modules/publicsite/domain"
	"github.com/kore/kore/internal/modules/publicsite/ports"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/httpx"
)

const (
	rateLimitWindow = time.Minute
	rateLimitMax    = 30
)

func RegisterRoutes(r chi.Router, svc *app.Service, appCache cache.Cache, keys cache.KeyBuilder) {
	r.Route("/public", func(pr chi.Router) {
		pr.Get("/pricing", getPricing(svc))
		pr.Get("/modules", listModules(svc))
		pr.With(rateLimit(appCache, keys, "leads")).Post("/leads", captureLead(svc))
		pr.Get("/booking/slots", listSlots(svc))
		pr.With(rateLimit(appCache, keys, "booking")).Post("/booking/appointments", bookAppointment(svc))
		pr.Post("/booking/appointments/{token}/cancel", cancelAppointment(svc))
		pr.Post("/booking/appointments/{token}/reschedule", rescheduleAppointment(svc))
	})
}

func getPricing(svc ports.PublicSiteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalog, err := svc.GetPricing(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]any{
			"catalog":        catalog,
			"publishableKey": svc.PublishableKey(),
		})
	}
}

func listModules(svc ports.PublicSiteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modules, err := svc.ListModules(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, modules)
	}
}

func captureLead(svc ports.PublicSiteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email     string `json:"email"`
			Company   string `json:"company"`
			Size      string `json:"size"`
			Need      string `json:"need"`
			UTMSource string `json:"utmSource"`
			Consent   bool   `json:"consent"`
			Honeypot  string `json:"website"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if req.Honeypot != "" {
			httpx.WriteData(w, http.StatusCreated, map[string]string{"status": "ok"})
			return
		}
		lead, err := svc.CaptureLead(r.Context(), ports.CaptureLeadCommand{
			Email:     req.Email,
			Company:   req.Company,
			Size:      req.Size,
			Need:      req.Need,
			UTMSource: req.UTMSource,
			Consent:   req.Consent,
		})
		if err != nil {
			if errors.Is(err, domain.ErrConsentRequired) {
				httpx.WriteError(w, http.StatusUnprocessableEntity, httpx.ErrCodeValidation, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusCreated, lead)
	}
}

func listSlots(svc ports.PublicSiteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := ports.SlotFilter{}
		if raw := r.URL.Query().Get("commercialId"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid commercialId")
				return
			}
			filter.CommercialID = &id
		}
		if raw := r.URL.Query().Get("from"); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid from")
				return
			}
			filter.From = t
		}
		if raw := r.URL.Query().Get("to"); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid to")
				return
			}
			filter.To = t
		}
		slots, err := svc.AvailableSlots(r.Context(), filter)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, slots)
	}
}

func bookAppointment(svc ports.PublicSiteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			LeadID       uuid.UUID             `json:"leadId"`
			CommercialID uuid.UUID             `json:"commercialId"`
			SlotID       uuid.UUID             `json:"slotId"`
			Channel      domain.MeetingChannel `json:"channel"`
			Email        string                `json:"email"`
			Name         string                `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		if req.Channel == "" {
			req.Channel = domain.ChannelVideo
		}
		appt, err := svc.BookAppointment(r.Context(), ports.BookCommand{
			LeadID:       req.LeadID,
			CommercialID: req.CommercialID,
			SlotID:       req.SlotID,
			Channel:      req.Channel,
			Email:        req.Email,
			Name:         req.Name,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrSlotAlreadyBooked):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			case errors.Is(err, domain.ErrSlotExpired):
				httpx.WriteError(w, http.StatusGone, httpx.ErrCodeValidation, err.Error())
			case errors.Is(err, domain.ErrSlotNotFound):
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusCreated, appt)
	}
}

func cancelAppointment(svc ports.PublicSiteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if err := svc.CancelAppointment(r.Context(), token); err != nil {
			if errors.Is(err, domain.ErrAppointmentNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			return
		}
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "canceled"})
	}
}

func rescheduleAppointment(svc ports.PublicSiteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		var req struct {
			NewSlotID uuid.UUID `json:"newSlotId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		appt, err := svc.Reschedule(r.Context(), ports.RescheduleCommand{
			Token:     token,
			NewSlotID: req.NewSlotID,
		})
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrAppointmentNotFound):
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, err.Error())
			case errors.Is(err, domain.ErrSlotAlreadyBooked), errors.Is(err, domain.ErrSlotExpired):
				httpx.WriteError(w, http.StatusConflict, httpx.ErrCodeConflict, err.Error())
			default:
				httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
			}
			return
		}
		httpx.WriteData(w, http.StatusOK, appt)
	}
}

func rateLimit(appCache cache.Cache, keys cache.KeyBuilder, scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appCache == nil || keys == nil {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			key := keys.PublicKey("publicsite", "ratelimit", scope, ip)
			var count int
			found, err := appCache.Get(r.Context(), key, &count)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if found && count >= rateLimitMax {
				httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrCodeTooManyRequests, "too many requests")
				return
			}
			count++
			_ = appCache.Set(r.Context(), key, count, rateLimitWindow)
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
