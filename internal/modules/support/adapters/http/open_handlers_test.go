package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	supporthttp "github.com/kore/kore/internal/modules/support/adapters/http"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/internal/modules/support/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type openTicketOrgStub struct {
	apps []orgdomain.Application
}

func (s *openTicketOrgStub) ListApplications(_ context.Context, _ kernel.TenantID, _ orgports.ApplicationListFilter) ([]orgdomain.Application, error) {
	return s.apps, nil
}

func (s *openTicketOrgStub) GetApplication(_ context.Context, _ kernel.TenantID, id uuid.UUID) (orgdomain.Application, error) {
	for _, app := range s.apps {
		if app.ID == id {
			return app, nil
		}
	}
	return orgdomain.Application{}, orgdomain.ErrApplicationNotFound
}

type openTicketSupportStub struct {
	created []ports.CreateTicketCommand
	ticket  domain.Ticket
}

func (s *openTicketSupportStub) List(context.Context, kernel.TenantID) ([]domain.Ticket, error) {
	panic("unused")
}
func (s *openTicketSupportStub) Get(context.Context, kernel.TenantID, uuid.UUID) (domain.Ticket, error) {
	panic("unused")
}
func (s *openTicketSupportStub) Create(_ context.Context, cmd ports.CreateTicketCommand) (domain.Ticket, error) {
	s.created = append(s.created, cmd)
	s.ticket = domain.NewTicket(cmd.TenantID, cmd.ApplicationID, cmd.Subject, cmd.Description, kernel.NormalizeRequestPriority(cmd.Priority), cmd.DueAt, cmd.ReporterID)
	return s.ticket, nil
}
func (s *openTicketSupportStub) Assign(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) (domain.Ticket, error) {
	panic("unused")
}
func (s *openTicketSupportStub) TakeOver(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) (domain.Ticket, error) {
	panic("unused")
}
func (s *openTicketSupportStub) AddReply(context.Context, ports.AddReplyCommand) (domain.TicketReply, error) {
	panic("unused")
}
func (s *openTicketSupportStub) Resolve(context.Context, kernel.TenantID, uuid.UUID) (domain.Ticket, error) {
	panic("unused")
}
func (s *openTicketSupportStub) IngestInboundEmails(context.Context, kernel.TenantID, uuid.UUID) (int, error) {
	panic("unused")
}

type alwaysOnChannels struct{}

func (alwaysOnChannels) IsChannelEnabled(context.Context, kernel.TenantID, kernel.RequestChannel) (bool, error) {
	return true, nil
}

func TestOpenCreateTicket_resolvesExistingLessonsStudioApp(t *testing.T) {
	existingID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	org := &openTicketOrgStub{apps: []orgdomain.Application{{
		ID:       existingID,
		TenantID: tenant,
		Libelle:  "Lessons-studio",
		Active:   true,
	}}}
	support := &openTicketSupportStub{}
	r := chi.NewRouter()
	supporthttp.RegisterOpenRoutes(r, support, org, alwaysOnChannels{})

	body := `{"subject":"Bug UI","description":"page crash","priority":"high"}`
	req := httptest.NewRequest(http.MethodPost, "/tickets", strings.NewReader(body))
	req = req.WithContext(authx.WithIdentity(req.Context(), authx.Identity{
		UserID:   uuid.New(),
		TenantID: tenant,
		Profile:  authx.ProfileAdmin,
		Roles:    []string{"api_key"},
	}))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, support.created, 1)
	require.Equal(t, existingID, support.created[0].ApplicationID)
	require.Nil(t, support.created[0].ReporterID)
	require.Equal(t, "Bug UI", support.created[0].Subject)

	var payload struct {
		Data supporthttp.OpenTicketResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	require.NotEmpty(t, payload.Data.ID)
	require.Equal(t, "Bug UI", payload.Data.Subject)
	require.Equal(t, "open", payload.Data.State)
}

func TestOpenCreateTicket_missingAppReturnsNotFound(t *testing.T) {
	r := chi.NewRouter()
	supporthttp.RegisterOpenRoutes(r, &openTicketSupportStub{}, &openTicketOrgStub{}, alwaysOnChannels{})
	req := httptest.NewRequest(http.MethodPost, "/tickets", strings.NewReader(`{"subject":"x","description":"y"}`))
	req = req.WithContext(authx.WithIdentity(req.Context(), authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
	}))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOpenCreateTicket_reusesExistingAppByLibelle(t *testing.T) {
	existingID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	org := &openTicketOrgStub{apps: []orgdomain.Application{{
		ID:       existingID,
		TenantID: tenant,
		Libelle:  "lessons-studio",
		Active:   true,
	}}}
	support := &openTicketSupportStub{}
	r := chi.NewRouter()
	supporthttp.RegisterOpenRoutes(r, support, org, alwaysOnChannels{})

	body, err := json.Marshal(map[string]any{
		"subject":            "Reuse",
		"description":        "already exists",
		"applicationLibelle": "Lessons-studio",
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req = req.WithContext(authx.WithIdentity(req.Context(), authx.Identity{
		UserID:   uuid.New(),
		TenantID: tenant,
	}))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, existingID, support.created[0].ApplicationID)
}

func TestOpenCreateTicket_validation(t *testing.T) {
	r := chi.NewRouter()
	supporthttp.RegisterOpenRoutes(r, &openTicketSupportStub{}, &openTicketOrgStub{}, alwaysOnChannels{})
	req := httptest.NewRequest(http.MethodPost, "/tickets", strings.NewReader(`{"subject":"","description":""}`))
	req = req.WithContext(authx.WithIdentity(req.Context(), authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
	}))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestOpenCreateTicket_applicationNotFound(t *testing.T) {
	r := chi.NewRouter()
	supporthttp.RegisterOpenRoutes(r, &openTicketSupportStub{}, &openTicketOrgStub{}, alwaysOnChannels{})
	missing := uuid.New()
	body, err := json.Marshal(map[string]any{
		"subject":       "x",
		"description":   "y",
		"applicationId": missing.String(),
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req = req.WithContext(authx.WithIdentity(req.Context(), authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
	}))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOpenCreateTicket_descriptionTooLong(t *testing.T) {
	existingID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	org := &openTicketOrgStub{apps: []orgdomain.Application{{
		ID: existingID, TenantID: tenant, Libelle: "Lessons-studio", Active: true,
	}}}
	r := chi.NewRouter()
	supporthttp.RegisterOpenRoutes(r, &openTicketSupportStub{}, org, alwaysOnChannels{})
	body, err := json.Marshal(map[string]any{
		"subject":     "x",
		"description": strings.Repeat("d", 100_001),
	})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req = req.WithContext(authx.WithIdentity(req.Context(), authx.Identity{
		UserID: uuid.New(), TenantID: tenant,
	}))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
