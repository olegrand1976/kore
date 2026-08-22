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
	"github.com/kore/kore/internal/modules/integrations/app"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

type stubTaigaRepo struct{}

func (stubTaigaRepo) UpsertExternalLink(context.Context, domain.ExternalLink) error { return nil }
func (stubTaigaRepo) FindExternalLinkByKore(context.Context, kernel.TenantID, string, uuid.UUID) (domain.ExternalLink, error) {
	return domain.ExternalLink{}, domain.ErrExternalLinkNotFound
}
func (stubTaigaRepo) UpsertUserMapping(context.Context, domain.UserMapping) error { return nil }

type moduleAuthorizer map[authx.Module]map[authx.Action]bool

func (a moduleAuthorizer) Can(_ context.Context, mod authx.Module, act authx.Action) bool {
	perms, ok := a[mod]
	if !ok {
		return false
	}
	return perms[act]
}

func TestTaigaWebhook_NotConfigured(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{})
	h := taigaWebhook(svc, "", "")
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestTaigaWebhook_InvalidSecret(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{})
	h := taigaWebhook(svc, "expected-secret", uuid.New().String())
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Taiga-Webhook-Secret", "wrong")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTaigaWebhook_OK(t *testing.T) {
	tenantID := uuid.New()
	svc := app.NewTaigaService(stubTaigaRepo{})
	h := taigaWebhook(svc, "expected-secret", tenantID.String())
	body := []byte(`{"action":"create","type":"userstory","data":{"id":1}}`)
	req := httptest.NewRequest(http.MethodPost, "/integrations/taiga/webhook", bytes.NewReader(body))
	req.Header.Set("X-Taiga-Webhook-Secret", "expected-secret")
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestFindTaigaLinkByDemand_ForbiddenWithoutTmaRead(t *testing.T) {
	svc := app.NewTaigaService(stubTaigaRepo{})
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
	svc := app.NewTaigaService(stubTaigaRepo{})
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
