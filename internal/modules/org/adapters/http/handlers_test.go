package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
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

type applicationOrgService struct {
	ports.OrganizationService
	created *ports.CreateApplicationCommand
	updated *ports.UpdateApplicationCommand
	active  *ports.SetApplicationActiveCommand
	listed  []domain.Application
	app     domain.Application
	err     error
}

func (s *applicationOrgService) CreateApplication(_ context.Context, cmd ports.CreateApplicationCommand) (domain.Application, error) {
	if s.err != nil {
		return domain.Application{}, s.err
	}
	s.created = &cmd
	return domain.Application{
		ID:                uuid.New(),
		TenantID:          cmd.TenantID,
		ServiceID:         cmd.ServiceID,
		Libelle:           cmd.Libelle,
		Proprietaire:      cmd.Proprietaire,
		ModeFacturation:   cmd.ModeFacturation,
		UOActivee:         cmd.UOActivee,
		ChefUtilisateurID: cmd.ChefUtilisateurID,
		BudgetDefautID:    cmd.BudgetDefautID,
		Active:            true,
	}, nil
}

func (s *applicationOrgService) UpdateApplication(_ context.Context, cmd ports.UpdateApplicationCommand) (domain.Application, error) {
	if s.err != nil {
		return domain.Application{}, s.err
	}
	s.updated = &cmd
	app := s.app
	if app.ID == uuid.Nil {
		app.ID = cmd.ApplicationID
	}
	if cmd.Libelle != nil {
		app.Libelle = *cmd.Libelle
	}
	if cmd.Proprietaire != nil {
		app.Proprietaire = *cmd.Proprietaire
	}
	if cmd.ModeFacturation != nil {
		app.ModeFacturation = *cmd.ModeFacturation
	}
	if cmd.UOActivee != nil {
		app.UOActivee = *cmd.UOActivee
	}
	if cmd.Active != nil {
		app.Active = *cmd.Active
	}
	return app, nil
}

func (s *applicationOrgService) SetApplicationActive(_ context.Context, cmd ports.SetApplicationActiveCommand) (domain.Application, error) {
	if s.err != nil {
		return domain.Application{}, s.err
	}
	s.active = &cmd
	return domain.Application{ID: cmd.ApplicationID, Active: cmd.Active}, nil
}

func (s *applicationOrgService) ListApplications(context.Context, kernel.TenantID, ports.ApplicationListFilter) ([]domain.Application, error) {
	return s.listed, s.err
}

func TestCreateApplication_created(t *testing.T) {
	svc := &applicationOrgService{}
	handler := createApplication(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	serviceID := uuid.New()
	chefID := uuid.New()

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/applications", map[string]any{
		"serviceId":         serviceID.String(),
		"libelle":           "Portail",
		"proprietaire":      "ACME",
		"modeFacturation":   "forfait",
		"uoActivee":         true,
		"chefUtilisateurId": chefID.String(),
	}))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", rec.Code, rec.Body.String())
	}
	if svc.created == nil || svc.created.Libelle != "Portail" || svc.created.Proprietaire != "ACME" {
		t.Fatalf("created = %+v", svc.created)
	}
	if svc.created.ChefUtilisateurID == nil || *svc.created.ChefUtilisateurID != chefID {
		t.Fatalf("chef = %v", svc.created.ChefUtilisateurID)
	}
}

func TestUpdateApplication_proprietaireOnly(t *testing.T) {
	appID := uuid.New()
	svc := &applicationOrgService{app: domain.Application{ID: appID, Libelle: "App", Active: true}}
	handler := updateApplication(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	req := requestWithIdentity(t, http.MethodPut, "/applications/"+appID.String(), map[string]any{
		"proprietaire": "Société Y",
	})
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{appID.String()}},
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if svc.updated == nil || svc.updated.Proprietaire == nil || *svc.updated.Proprietaire != "Société Y" {
		t.Fatalf("updated = %+v", svc.updated)
	}
}

func TestUpdateApplication_clearBudgetDefautHTTP(t *testing.T) {
	appID := uuid.New()
	svc := &applicationOrgService{app: domain.Application{ID: appID, Libelle: "App", Active: true}}
	handler := updateApplication(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	req := requestWithIdentity(t, http.MethodPut, "/applications/"+appID.String(), map[string]any{
		"budgetDefautId": nil,
	})
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{appID.String()}},
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if svc.updated == nil || svc.updated.BudgetDefautID == nil {
		t.Fatalf("expected BudgetDefautID set on command, got %+v", svc.updated)
	}
	if *svc.updated.BudgetDefautID != nil {
		t.Fatalf("expected clear (ptr to nil), got %v", *svc.updated.BudgetDefautID)
	}
}

func TestUpdateApplication_notFoundHTTP(t *testing.T) {
	appID := uuid.New()
	svc := &applicationOrgService{err: domain.ErrApplicationNotFound}
	handler := updateApplication(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	req := requestWithIdentity(t, http.MethodPut, "/applications/"+appID.String(), map[string]any{
		"libelle": "X",
	})
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{appID.String()}},
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestDeactivateApplication_ok(t *testing.T) {
	appID := uuid.New()
	svc := &applicationOrgService{}
	handler := deactivateApplication(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	req := requestWithIdentity(t, http.MethodPatch, "/applications/"+appID.String()+"/deactivate", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{appID.String()}},
	}))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.active == nil || svc.active.Active {
		t.Fatalf("active cmd = %+v", svc.active)
	}
}

func TestListApplications_activeAll(t *testing.T) {
	svc := &applicationOrgService{listed: []domain.Application{{Libelle: "A"}, {Libelle: "B"}}}
	handler := listApplications(svc, stubAuthorizer{module: "org", action: authx.ActionRead, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodGet, "/applications?active=all", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
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
