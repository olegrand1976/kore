package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

// stubAuthorizer autorise un couple (module, action) donné et refuse tout le reste.
type stubAuthorizer struct {
	module authx.Module
	action authx.Action
	allow  bool
}

func (s stubAuthorizer) Can(_ context.Context, module authx.Module, action authx.Action) bool {
	return s.allow && module == s.module && action == s.action
}

// equipeOrgService n'implémente que CreateEquipe/ListSites : les autres méthodes de
// ports.OrganizationService ne sont pas sollicitées par les handlers testés.
type equipeOrgService struct {
	ports.OrganizationService
	created *ports.CreateEquipeCommand
	err     error
	sites   []domain.SiteSummary
}

func (s *equipeOrgService) CreateEquipe(_ context.Context, cmd ports.CreateEquipeCommand) (domain.Equipe, error) {
	if s.err != nil {
		return domain.Equipe{}, s.err
	}
	s.created = &cmd
	return domain.Equipe{
		ID:            uuid.New(),
		TenantID:      cmd.TenantID,
		ApplicationID: cmd.ApplicationID,
		Libelle:       cmd.Libelle,
		ResponsableID: cmd.ResponsableID,
	}, nil
}

func (s *equipeOrgService) ListSites(context.Context, kernel.TenantID) ([]domain.SiteSummary, error) {
	return s.sites, nil
}

func requestWithIdentity(t *testing.T, method, target string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	ctx := authx.WithIdentity(req.Context(), authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Profile:  authx.ProfileAdmin,
	})
	return req.WithContext(ctx)
}

func decodeErrorCode(t *testing.T, rec *httptest.ResponseRecorder) httpx.ErrorCode {
	t.Helper()
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected an error envelope")
	}
	return env.Error.Code
}

func TestCreateEquipe_forbiddenWithoutOrgWrite(t *testing.T) {
	svc := &equipeOrgService{}
	handler := createEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionRead, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/equipes", map[string]any{
		"applicationId": uuid.New().String(),
		"libelle":       "Équipe Dev",
	}))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if svc.created != nil {
		t.Fatal("expected no creation when org:E is missing")
	}
}

func TestCreateEquipe_unprocessableWithoutApplication(t *testing.T) {
	svc := &equipeOrgService{err: domain.ErrEquipeWithoutApplication}
	handler := createEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/equipes", map[string]any{
		"libelle": "Équipe orpheline",
	}))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	if code := decodeErrorCode(t, rec); code != httpx.ErrCodeValidation {
		t.Fatalf("code = %s, want %s", code, httpx.ErrCodeValidation)
	}
}

func TestCreateEquipe_created(t *testing.T) {
	svc := &equipeOrgService{}
	handler := createEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	appID := uuid.New()
	responsable := uuid.New()
	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/equipes", map[string]any{
		"applicationId": appID.String(),
		"libelle":       "Équipe TMA",
		"responsableId": responsable.String(),
	}))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", rec.Code, rec.Body.String())
	}
	if svc.created == nil {
		t.Fatal("expected CreateEquipe to be called")
	}
	if svc.created.ApplicationID != appID {
		t.Fatalf("applicationID = %v, want %v", svc.created.ApplicationID, appID)
	}
	if svc.created.Libelle != "Équipe TMA" {
		t.Fatalf("libelle = %q", svc.created.Libelle)
	}
	if svc.created.ResponsableID == nil || *svc.created.ResponsableID != responsable {
		t.Fatalf("responsableID = %v, want %v", svc.created.ResponsableID, responsable)
	}
}

func TestCreateEquipe_badRequestOnInvalidBody(t *testing.T) {
	svc := &equipeOrgService{}
	handler := createEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	req := httptest.NewRequest(http.MethodPost, "/equipes", bytes.NewBufferString("{not-json"))
	req = req.WithContext(authx.WithIdentity(req.Context(), authx.Identity{
		TenantID: kernel.NewTenantID(uuid.New()),
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestListSites_forbiddenWithoutOrgRead(t *testing.T) {
	handler := listSites(&equipeOrgService{}, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodGet, "/sites", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestListSites_returnsSites(t *testing.T) {
	svc := &equipeOrgService{sites: []domain.SiteSummary{{ID: uuid.New(), Libelle: "Paris HQ"}}}
	handler := listSites(svc, stubAuthorizer{module: "org", action: authx.ActionRead, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodGet, "/sites", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var payload struct {
		Data []domain.SiteSummary `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Data) != 1 || payload.Data[0].Libelle != "Paris HQ" {
		t.Fatalf("data = %+v", payload.Data)
	}
}
