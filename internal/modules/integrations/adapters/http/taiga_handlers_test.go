package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/app"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

type stubTaigaRepo struct{}

func (stubTaigaRepo) UpsertExternalLink(context.Context, domain.ExternalLink) error { return nil }
func (stubTaigaRepo) InsertApplicationProjectLink(context.Context, domain.ExternalLink) error {
	return nil
}
func (stubTaigaRepo) FindExternalLinkByKore(context.Context, kernel.TenantID, string, uuid.UUID) (domain.ExternalLink, error) {
	return domain.ExternalLink{}, domain.ErrExternalLinkNotFound
}
func (stubTaigaRepo) UpsertUserMapping(context.Context, domain.UserMapping) error { return nil }
func (stubTaigaRepo) ListLinkedTaigaProjectIDs(context.Context, kernel.TenantID) ([]string, error) {
	return nil, nil
}
func (stubTaigaRepo) ListUserMappings(context.Context, kernel.TenantID, string) ([]domain.UserMapping, error) {
	return nil, nil
}

type moduleAuthorizer map[authx.Module]map[authx.Action]bool

func (a moduleAuthorizer) Can(_ context.Context, mod authx.Module, act authx.Action) bool {
	perms, ok := a[mod]
	if !ok {
		return false
	}
	return perms[act]
}

func TestTaigaWebhook_NotConfigured(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := taigaWebhook(svc, nil, nil, "", uuid.New().String(), app.TaigaKoreMapping{})
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestTaigaWebhook_InvalidSecret(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := taigaWebhook(svc, nil, nil, "expected-secret", uuid.New().String(), app.TaigaKoreMapping{})
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Taiga-Webhook-Secret", "wrong")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTaigaWebhook_OK_NativeSignature(t *testing.T) {
	tenantID := uuid.New()
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := taigaWebhook(svc, nil, nil, "expected-secret", tenantID.String(), app.TaigaKoreMapping{})
	body := []byte(`{"action":"create","type":"userstory","data":{"id":1}}`)
	mac := hmac.New(sha1.New, []byte("expected-secret"))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader(body))
	req.Header.Set("X-TAIGA-WEBHOOK-SIGNATURE", sig)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestTaigaWebhook_OK(t *testing.T) {
	tenantID := uuid.New()
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := taigaWebhook(svc, nil, nil, "expected-secret", tenantID.String(), app.TaigaKoreMapping{})
	body := []byte(`{"action":"create","type":"userstory","data":{"id":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader(body))
	req.Header.Set("X-Taiga-Webhook-Secret", "expected-secret")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestTaigaWebhook_TenantFromMapping(t *testing.T) {
	defaultTenant := uuid.New().String()
	mappedTenant := uuid.New().String()
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := taigaWebhook(svc, nil, nil, "expected-secret", defaultTenant, app.TaigaKoreMapping{
		Projects: map[string]app.TaigaProjectMapping{
			"1": {KoreTenantID: mappedTenant},
		},
	})
	body := []byte(`{"action":"create","type":"userstory","data":{"id":1,"project":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader(body))
	req.Header.Set("X-Taiga-Webhook-Secret", "expected-secret")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestFindTaigaLinkByDemand_ForbiddenWithoutTmaRead(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := findTaigaLinkByDemand(svc, moduleAuthorizer{})
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		Profile:  "Collaborateur",
	})
	req := httptest.NewRequest(http.MethodGet, "/integrations/taiga/links/by-demand/"+uuid.New().String(), nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusForbidden)
	}
}

func TestFindTaigaLinkByDemand_NotFound(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	authz := moduleAuthorizer{
		"tma": {authx.ActionRead: true},
	}
	h := findTaigaLinkByDemand(svc, authz)
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		Profile:  "Collaborateur",
	})
	demandID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/integrations/taiga/links/by-demand/"+demandID.String(), nil).WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", demandID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusNotFound)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error == nil || env.Error.Code != httpx.ErrCodeNotFound {
		t.Fatalf("unexpected envelope: %+v", env.Error)
	}
}

func TestListUnlinkedTaigaProjects_ForbiddenWithoutOrgWrite(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := listUnlinkedTaigaProjects(svc, moduleAuthorizer{})
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		Profile:  "Collaborateur",
	})
	req := httptest.NewRequest(http.MethodGet, "/integrations/taiga/projects/unlinked", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusForbidden)
	}
}

func TestFindTaigaLinkByApplication_ForbiddenWithoutOrgRead(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := findTaigaLinkByApplication(svc, moduleAuthorizer{})
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		Profile:  "Collaborateur",
	})
	req := httptest.NewRequest(http.MethodGet, "/integrations/taiga/links/by-application/"+uuid.New().String(), nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusForbidden)
	}
}

func TestListTaigaUserMappings_OK(t *testing.T) {
	repo := &listMappingsRepo{}
	tenant := kernel.NewTenantID(uuid.New())
	koreUser := uuid.New()
	repo.mappings = []domain.UserMapping{{
		TenantID:         tenant,
		Provider:         "taiga",
		ExternalUserID:   "42",
		ExternalUsername: "alice",
		KoreUserID:       koreUser,
		MatchMethod:      "email",
	}}
	svc := app.NewTaigaService(repo, app.TaigaConfig{}, nil, nil, nil)
	h := listTaigaUserMappings(svc, moduleAuthorizer{
		"integrations": {authx.ActionRead: true},
	})
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: tenant,
		UserID:   uuid.New(),
		Profile:  "Administrateur",
	})
	req := httptest.NewRequest(http.MethodGet, "/integrations/taiga/user-mappings", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var env struct {
		Data []domain.UserMapping `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 1 || env.Data[0].ExternalUsername != "alice" {
		t.Fatalf("data: %+v", env.Data)
	}
}

func TestListTaigaUserMappings_Forbidden(t *testing.T) {
	svc := app.NewTaigaService(&stubTaigaRepo{}, app.TaigaConfig{}, nil, nil, nil)
	h := listTaigaUserMappings(svc, moduleAuthorizer{})
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		Profile:  "Collaborateur",
	})
	req := httptest.NewRequest(http.MethodGet, "/integrations/taiga/user-mappings", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusForbidden)
	}
}

func TestUpsertTaigaUserMapping_Conflict(t *testing.T) {
	repo := &conflictMappingRepo{}
	svc := app.NewTaigaService(repo, app.TaigaConfig{}, nil, nil, nil)
	h := upsertTaigaUserMapping(svc, moduleAuthorizer{
		"integrations": {authx.ActionWrite: true},
	})
	ctx := authx.WithIdentity(context.Background(), authx.Identity{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		Profile:  "Administrateur",
	})
	body := `{"taigaUserId":42,"taigaUsername":"alice","koreUserId":"` + uuid.New().String() + `","matchMethod":"email"}`
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/user-mappings", strings.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status: got %d want %d body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

type conflictMappingRepo struct {
	stubTaigaRepo
}

func (conflictMappingRepo) UpsertUserMapping(context.Context, domain.UserMapping) error {
	return domain.ErrTaigaKoreUserAlreadyMapped
}

type listMappingsRepo struct {
	stubTaigaRepo
	mappings []domain.UserMapping
}

func (r *listMappingsRepo) ListUserMappings(_ context.Context, tenant kernel.TenantID, provider string) ([]domain.UserMapping, error) {
	var out []domain.UserMapping
	for _, m := range r.mappings {
		if m.TenantID == tenant && m.Provider == provider {
			out = append(out, m)
		}
	}
	return out, nil
}
