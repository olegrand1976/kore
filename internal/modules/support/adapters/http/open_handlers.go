package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/internal/modules/support/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

const (
	defaultOpenTicketApplicationLibelle = "Lessons-studio"
	openTicketMaxBodyBytes              = 1 << 20 // 1 MiB
	openTicketMaxSubjectRunes           = 200
	openTicketMaxDescriptionRunes       = 100_000 // includes diagnostic markdown + screenshot data-url
)

// OpenApplicationResolver resolves the application targeted by an open ticket.
// Application ensure/create belongs to bootstrap scripts, not this hot path.
type OpenApplicationResolver interface {
	ListApplications(ctx context.Context, tenant kernel.TenantID, filter orgports.ApplicationListFilter) ([]orgdomain.Application, error)
	GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (orgdomain.Application, error)
}

// OpenTicketResponse is the stable JSON contract for POST /open/tickets.
type OpenTicketResponse struct {
	ID            string `json:"id"`
	ApplicationID string `json:"applicationId"`
	Subject       string `json:"subject"`
	State         string `json:"state"`
	Priority      string `json:"priority"`
}

// RegisterOpenRoutes mounts API-key public ticket routes under the given router
// (typically /api/v1/open).
func RegisterOpenRoutes(r chi.Router, svc ports.SupportService, org OpenApplicationResolver, channels kernel.RequestChannelReader) {
	r.Post("/tickets", openCreateTicket(svc, org, channels))
}

func openCreateTicket(svc ports.SupportService, org OpenApplicationResolver, channels kernel.RequestChannelReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireRequestChannel(w, r, channels, kernel.RequestChannelSupport) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, openTicketMaxBodyBytes)
		var req struct {
			ApplicationID      *uuid.UUID `json:"applicationId"`
			ApplicationLibelle string     `json:"applicationLibelle"`
			Subject            string     `json:"subject"`
			Description        string     `json:"description"`
			Priority           string     `json:"priority"`
		}
		dec := json.NewDecoder(io.LimitReader(r.Body, openTicketMaxBodyBytes+1))
		if err := dec.Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				httpx.WriteError(w, http.StatusRequestEntityTooLarge, httpx.ErrCodeValidation, "body too large")
				return
			}
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "invalid body")
			return
		}
		subject := strings.TrimSpace(req.Subject)
		description := strings.TrimSpace(req.Description)
		if subject == "" || description == "" {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "subject and description are required")
			return
		}
		if utf8.RuneCountInString(subject) > openTicketMaxSubjectRunes {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "subject too long")
			return
		}
		if utf8.RuneCountInString(description) > openTicketMaxDescriptionRunes {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrCodeValidation, "description too long")
			return
		}
		identity, ok := authx.FromContext(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrCodeUnauthorized, "unauthorized")
			return
		}
		appID, err := resolveOpenTicketApplication(r.Context(), org, identity.TenantID, req.ApplicationID, req.ApplicationLibelle)
		if err != nil {
			if errors.Is(err, orgdomain.ErrApplicationNotFound) {
				httpx.WriteError(w, http.StatusNotFound, httpx.ErrCodeNotFound, "application not found")
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "failed to resolve application")
			return
		}
		t, err := svc.Create(r.Context(), ports.CreateTicketCommand{
			TenantID:      identity.TenantID,
			ApplicationID: appID,
			Subject:       subject,
			Description:   description,
			Priority:      req.Priority,
			ReporterID:    nil,
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "failed to create ticket")
			return
		}
		httpx.WriteData(w, http.StatusCreated, toOpenTicketResponse(t))
	}
}

func toOpenTicketResponse(t domain.Ticket) OpenTicketResponse {
	return OpenTicketResponse{
		ID:            t.ID.String(),
		ApplicationID: t.ApplicationID.String(),
		Subject:       t.Subject,
		State:         string(t.State),
		Priority:      string(t.Priority),
	}
}

func resolveOpenTicketApplication(
	ctx context.Context,
	org OpenApplicationResolver,
	tenant kernel.TenantID,
	applicationID *uuid.UUID,
	libelle string,
) (uuid.UUID, error) {
	if applicationID != nil && *applicationID != uuid.Nil {
		app, err := org.GetApplication(ctx, tenant, *applicationID)
		if err != nil {
			return uuid.Nil, err
		}
		return app.ID, nil
	}
	label := strings.TrimSpace(libelle)
	if label == "" {
		label = defaultOpenTicketApplicationLibelle
	}
	apps, err := org.ListApplications(ctx, tenant, orgports.ApplicationListFilter{})
	if err != nil {
		return uuid.Nil, err
	}
	for _, app := range apps {
		if strings.EqualFold(strings.TrimSpace(app.Libelle), label) {
			return app.ID, nil
		}
	}
	return uuid.Nil, orgdomain.ErrApplicationNotFound
}
