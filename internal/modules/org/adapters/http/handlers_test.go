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

func (s *equipeOrgService) UpdateSite(_ context.Context, cmd ports.UpdateSiteCommand) (domain.SiteSummary, error) {
	if s.err != nil {
		return domain.SiteSummary{}, s.err
	}
	return domain.SiteSummary{ID: cmd.SiteID, Libelle: cmd.Libelle}, nil
}

func (s *equipeOrgService) UpdateService(_ context.Context, cmd ports.UpdateServiceCommand) (domain.ServiceSummary, error) {
	if s.err != nil {
		return domain.ServiceSummary{}, s.err
	}
	return domain.ServiceSummary{ID: cmd.ServiceID, Libelle: cmd.Libelle}, nil
}

func (s *equipeOrgService) UpdateEquipe(_ context.Context, cmd ports.UpdateEquipeCommand) (domain.Equipe, error) {
	if s.err != nil {
		return domain.Equipe{}, s.err
	}
	return domain.Equipe{
		ID:            cmd.EquipeID,
		TenantID:      cmd.TenantID,
		Libelle:       cmd.Libelle,
		ResponsableID: cmd.ResponsableID,
	}, nil
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
		Libelle:           cmd.Libelle,
		Proprietaire:      cmd.Proprietaire,
		ModeFacturation:   cmd.ModeFacturation,
		UOActivee:         cmd.UOActivee,
		ChefUtilisateurID: cmd.ChefUtilisateurID,
		BudgetDefautID:    cmd.BudgetDefautID,
		Active:            true,
		SiteIDs:           cmd.SiteIDs,
		ServiceIDs:        cmd.ServiceIDs,
		EquipeIDs:         cmd.EquipeIDs,
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
	siteID := uuid.New()
	chefID := uuid.New()

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/applications", map[string]any{
		"serviceIds":        []string{serviceID.String()},
		"siteIds":           []string{siteID.String()},
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
	if len(svc.created.ServiceIDs) != 1 || svc.created.ServiceIDs[0] != serviceID {
		t.Fatalf("serviceIds = %v", svc.created.ServiceIDs)
	}
	if len(svc.created.SiteIDs) != 1 || svc.created.SiteIDs[0] != siteID {
		t.Fatalf("siteIds = %v", svc.created.SiteIDs)
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

func TestUpdateSite_ok(t *testing.T) {
	siteID := uuid.New()
	svc := &equipeOrgService{}
	handler := updateSite(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/sites/"+siteID.String(), map[string]any{
		"libelle": "Lyon",
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", siteID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateService_ok(t *testing.T) {
	serviceID := uuid.New()
	svc := &equipeOrgService{}
	handler := updateService(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/services/"+serviceID.String(), map[string]any{
		"libelle": "Support",
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", serviceID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateEquipe_ok(t *testing.T) {
	equipeID := uuid.New()
	responsable := uuid.New()
	svc := &equipeOrgService{}
	handler := updateEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/equipes/"+equipeID.String(), map[string]any{
		"libelle":       "Équipe Data",
		"responsableId": responsable.String(),
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", equipeID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestUpdateEquipe_forbiddenWithoutOrgWrite(t *testing.T) {
	equipeID := uuid.New()
	svc := &equipeOrgService{}
	handler := updateEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionRead, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/equipes/"+equipeID.String(), map[string]any{
		"libelle": "X",
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", equipeID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestUpdateEquipe_unprocessableEmptyLibelle(t *testing.T) {
	equipeID := uuid.New()
	svc := &equipeOrgService{err: domain.ErrInvalidEquipeLibelle}
	handler := updateEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/equipes/"+equipeID.String(), map[string]any{
		"libelle":       "   ",
		"responsableId": nil,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", equipeID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestUpdateEquipe_notFound(t *testing.T) {
	equipeID := uuid.New()
	svc := &equipeOrgService{err: domain.ErrEquipeNotFound}
	handler := updateEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/equipes/"+equipeID.String(), map[string]any{
		"libelle":       "X",
		"responsableId": nil,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", equipeID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestUpdateEquipe_unknownResponsable(t *testing.T) {
	equipeID := uuid.New()
	svc := &equipeOrgService{err: domain.ErrUserNotFound}
	handler := updateEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/equipes/"+equipeID.String(), map[string]any{
		"libelle":       "Dev",
		"responsableId": uuid.New().String(),
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", equipeID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestUpdateEquipe_invalidID(t *testing.T) {
	svc := &equipeOrgService{}
	handler := updateEquipe(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/equipes/not-a-uuid", map[string]any{
		"libelle":       "X",
		"responsableId": nil,
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	handler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

// anyModuleAuthorizer grants ActionRead on any of the listed modules.
type anyModuleAuthorizer struct {
	modules []authx.Module
}

func (a anyModuleAuthorizer) Can(_ context.Context, module authx.Module, action authx.Action) bool {
	if action != authx.ActionRead {
		return false
	}
	for _, m := range a.modules {
		if m == module {
			return true
		}
	}
	return false
}

type stubClientService struct {
	ports.ClientService
	listed    bool
	created   *ports.CreateClientCommand
	updated   *ports.UpdateClientCommand
	items     []domain.Client
	createErr error
	updateErr error
}

func (s *stubClientService) ListClients(context.Context, kernel.TenantID) ([]domain.Client, error) {
	s.listed = true
	return s.items, nil
}

func (s *stubClientService) CreateClient(_ context.Context, cmd ports.CreateClientCommand) (domain.Client, error) {
	s.created = &cmd
	if s.createErr != nil {
		return domain.Client{}, s.createErr
	}
	return domain.Client{ID: uuid.New(), RaisonSociale: cmd.RaisonSociale, Pays: cmd.Pays}, nil
}

func (s *stubClientService) UpdateClient(_ context.Context, cmd ports.UpdateClientCommand) (domain.Client, error) {
	s.updated = &cmd
	if s.updateErr != nil {
		return domain.Client{}, s.updateErr
	}
	return domain.Client{ID: cmd.ClientID, RaisonSociale: cmd.RaisonSociale, Pays: cmd.Pays}, nil
}

func TestListClients_forbiddenWithoutOrgCraOrSsiiRead(t *testing.T) {
	svc := &stubClientService{items: []domain.Client{{RaisonSociale: "Acme"}}}
	handler := listClients(svc, stubAuthorizer{module: "budget", action: authx.ActionRead, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodGet, "/clients", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if svc.listed {
		t.Fatal("expected ListClients not called when read is forbidden")
	}
}

func TestListClients_allowedWithCraRead(t *testing.T) {
	svc := &stubClientService{items: []domain.Client{{ID: uuid.New(), RaisonSociale: "Acme"}}}
	handler := listClients(svc, anyModuleAuthorizer{modules: []authx.Module{"cra"}})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodGet, "/clients", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !svc.listed {
		t.Fatal("expected ListClients to be called")
	}
}

func TestCreateClient_rejectsInvalidPays(t *testing.T) {
	svc := &stubClientService{createErr: domain.ErrInvalidPays}
	handler := createClient(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/clients", map[string]any{
		"raisonSociale": "Acme",
		"pays":          "DE",
	}))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestCreateClient_rejectsEmptyName(t *testing.T) {
	svc := &stubClientService{createErr: domain.ErrInvalidClientName}
	handler := createClient(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/clients", map[string]any{
		"raisonSociale": "  ",
	}))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestUpdateClient_ok(t *testing.T) {
	clientID := uuid.New()
	svc := &stubClientService{}
	handler := updateClient(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/clients/"+clientID.String(), map[string]any{
		"raisonSociale": "Acme BE",
		"pays":          "BE",
		"ville":         "Bruxelles",
	})
	req = req.WithContext(req.Context())
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", clientID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.updated == nil || svc.updated.RaisonSociale != "Acme BE" || svc.updated.Pays != "BE" {
		t.Fatalf("updated cmd = %+v", svc.updated)
	}
}

func TestUpdateClient_notFound(t *testing.T) {
	clientID := uuid.New()
	svc := &stubClientService{updateErr: domain.ErrClientNotFound}
	handler := updateClient(svc, stubAuthorizer{module: "org", action: authx.ActionWrite, allow: true})

	rec := httptest.NewRecorder()
	req := requestWithIdentity(t, http.MethodPut, "/clients/"+clientID.String(), map[string]any{
		"raisonSociale": "Acme",
	})
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", clientID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
