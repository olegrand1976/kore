package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type stubGetOrCreateService struct {
	ports.CRAService
	lastUserID uuid.UUID
	created    bool
}

func (s *stubGetOrCreateService) GetOrCreate(_ context.Context, _ kernel.TenantID, userID ports.UserID, month domain.Month) (domain.Timesheet, bool, error) {
	s.lastUserID = userID
	return domain.Timesheet{
		ID:     uuid.New(),
		UserID: userID,
		Month:  month,
		Status: domain.StatusBrouillon,
	}, s.created, nil
}

type stubAuthorizerFlags struct {
	read     bool
	validate bool
}

func (s stubAuthorizerFlags) Can(_ context.Context, module authx.Module, action authx.Action) bool {
	if module != "cra" {
		return false
	}
	switch action {
	case authx.ActionRead:
		return s.read
	case authx.ActionValidate:
		return s.validate
	default:
		return false
	}
}

func serveGetTimesheet(t *testing.T, svc ports.CRAService, authorizer authx.Authorizer, identity authx.Identity, rawQuery string) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	r.Get("/timesheets", getTimesheet(svc, authorizer))
	req := httptest.NewRequest(http.MethodGet, "/timesheets?"+rawQuery, nil)
	req = req.WithContext(authx.WithIdentity(req.Context(), identity))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestGetTimesheet_UsesSessionUserByDefault(t *testing.T) {
	svc := &stubGetOrCreateService{created: true}
	identity := adminIdentity()
	rec := serveGetTimesheet(t, svc, stubAuthorizerFlags{read: true, validate: true}, identity, "month=2026-08")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200 body=%s", rec.Code, rec.Body.String())
	}
	if svc.lastUserID != identity.UserID {
		t.Fatalf("userId: got %s want session %s", svc.lastUserID, identity.UserID)
	}
}

func TestGetTimesheet_UserIdRequiresValidate(t *testing.T) {
	svc := &stubGetOrCreateService{}
	target := uuid.New()
	rec := serveGetTimesheet(t, svc, stubAuthorizerFlags{read: true, validate: false}, adminIdentity(), "month=2026-08&userId="+target.String())
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want 403", rec.Code)
	}
}

func TestGetTimesheet_ValidatorCanTargetCollaborator(t *testing.T) {
	svc := &stubGetOrCreateService{created: false}
	target := uuid.New()
	rec := serveGetTimesheet(t, svc, stubAuthorizerFlags{read: true, validate: true}, adminIdentity(), "month=2026-08&userId="+target.String())
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200 body=%s", rec.Code, rec.Body.String())
	}
	if svc.lastUserID != target {
		t.Fatalf("userId: got %s want %s", svc.lastUserID, target)
	}
	var envelope struct {
		Data struct {
			Created bool `json:"created"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("json: %v", err)
	}
	if envelope.Data.Created {
		t.Fatal("expected created=false")
	}
}
