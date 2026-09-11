package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

func TestWriteCRAError_BusinessCodes(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   httpx.ErrorCode
	}{
		{
			name:       "already validated",
			err:        domain.ErrCRAAlreadyValidated,
			wantStatus: http.StatusConflict,
			wantCode:   httpx.ErrCodeCRAAlreadyValidated,
		},
		{
			name:       "already invoiced",
			err:        domain.ErrCRAAlreadyInvoiced,
			wantStatus: http.StatusConflict,
			wantCode:   httpx.ErrCodeCRAAlreadyInvoiced,
		},
		{
			name:       "commercial info",
			err:        domain.ErrCommercialInfoRequired,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   httpx.ErrCodeCommercialInfoRequired,
		},
		{
			name:       "day capacity",
			err:        domain.ErrDayCapacityExceeded,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   httpx.ErrCodeDayCapacityExceeded,
		},
		{
			name:       "conflict absence",
			err:        domain.ErrCRAConflictAbsence,
			wantStatus: http.StatusConflict,
			wantCode:   httpx.ErrCodeCRAConflictAbsence,
		},
		{
			name:       "week incomplete",
			err:        domain.ErrWeekIncomplete,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   httpx.ErrCodeWeekIncomplete,
		},
		{
			name:       "no logged time",
			err:        domain.ErrCRANoLoggedTime,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   httpx.ErrCodeCRANoLoggedTime,
		},
		{
			name:       "not final",
			err:        domain.ErrCRANotFinal,
			wantStatus: http.StatusConflict,
			wantCode:   httpx.ErrCodeConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeCRAError(rec, tc.err)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d", rec.Code, tc.wantStatus)
			}
			var env httpx.Envelope
			if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if env.Error == nil {
				t.Fatal("expected error envelope")
			}
			if env.Error.Code != tc.wantCode {
				t.Fatalf("code: got %s want %s", env.Error.Code, tc.wantCode)
			}
		})
	}
}

func TestWriteCRAError_InternalFallback(t *testing.T) {
	rec := httptest.NewRecorder()
	writeCRAError(rec, errors.New("boom"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

type stubAuthorizer struct {
	module authx.Module
	action authx.Action
	allow  bool
}

func (s stubAuthorizer) Can(_ context.Context, module authx.Module, action authx.Action) bool {
	return s.allow && module == s.module && action == s.action
}

type stubDeleteCRAService struct {
	ports.CRAService
	deleted     ports.TimesheetID
	unvalidated ports.TimesheetID
	err         error
}

func (s *stubDeleteCRAService) DeleteTimesheet(_ context.Context, _ kernel.TenantID, id ports.TimesheetID) error {
	s.deleted = id
	return s.err
}

func (s *stubDeleteCRAService) UnvalidateTimesheet(_ context.Context, _ kernel.TenantID, id ports.TimesheetID) error {
	s.unvalidated = id
	return s.err
}

func serveDeleteTimesheet(t *testing.T, svc ports.CRAService, authorizer authx.Authorizer, identity authx.Identity, id string) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	r.Delete("/timesheets/{id}", deleteTimesheet(svc, authorizer))
	req := httptest.NewRequest(http.MethodDelete, "/timesheets/"+id, nil)
	req = req.WithContext(authx.WithIdentity(req.Context(), identity))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func adminIdentity() authx.Identity {
	return authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Profile:  authx.ProfileAdmin,
	}
}

func TestDeleteTimesheet_ForbiddenForNonAdmin(t *testing.T) {
	svc := &stubDeleteCRAService{}
	id := uuid.New()
	rec := serveDeleteTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Profile:  authx.ProfileCollaborateur,
	}, id.String())
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want 403", rec.Code)
	}
	if svc.deleted != uuid.Nil {
		t.Fatal("expected no delete for non-admin")
	}
}

func TestDeleteTimesheet_AdminOK(t *testing.T) {
	svc := &stubDeleteCRAService{}
	id := uuid.New()
	rec := serveDeleteTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), id.String())
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	if svc.deleted != id {
		t.Fatalf("deleted id: got %s want %s", svc.deleted, id)
	}
}

func TestDeleteTimesheet_InvalidID(t *testing.T) {
	svc := &stubDeleteCRAService{}
	rec := serveDeleteTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), "not-a-uuid")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
	if svc.deleted != uuid.Nil {
		t.Fatal("expected no delete for invalid id")
	}
}

func TestDeleteTimesheet_NotFound(t *testing.T) {
	svc := &stubDeleteCRAService{err: domain.ErrTimesheetNotFound}
	rec := serveDeleteTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), uuid.New().String())
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rec.Code)
	}
}

func TestDeleteTimesheet_FinalConflict(t *testing.T) {
	svc := &stubDeleteCRAService{err: domain.ErrCRAAlreadyValidated}
	rec := serveDeleteTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), uuid.New().String())
	if rec.Code != http.StatusConflict {
		t.Fatalf("status: got %d want 409", rec.Code)
	}
}

func TestDeleteTimesheet_InvoicedConflict(t *testing.T) {
	svc := &stubDeleteCRAService{err: domain.ErrCRAAlreadyInvoiced}
	rec := serveDeleteTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), uuid.New().String())
	if rec.Code != http.StatusConflict {
		t.Fatalf("status: got %d want 409", rec.Code)
	}
}

func serveUnvalidateTimesheet(t *testing.T, svc ports.CRAService, authorizer authx.Authorizer, identity authx.Identity, id string) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	r.Post("/timesheets/{id}/unvalidate", unvalidateTimesheet(svc, authorizer))
	req := httptest.NewRequest(http.MethodPost, "/timesheets/"+id+"/unvalidate", nil)
	req = req.WithContext(authx.WithIdentity(req.Context(), identity))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestUnvalidateTimesheet_ForbiddenForNonAdmin(t *testing.T) {
	svc := &stubDeleteCRAService{}
	id := uuid.New()
	rec := serveUnvalidateTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Profile:  authx.ProfileCollaborateur,
	}, id.String())
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want 403", rec.Code)
	}
	if svc.unvalidated != uuid.Nil {
		t.Fatal("expected no unvalidate for non-admin")
	}
}

func TestUnvalidateTimesheet_AdminOK(t *testing.T) {
	svc := &stubDeleteCRAService{}
	id := uuid.New()
	rec := serveUnvalidateTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), id.String())
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	if svc.unvalidated != id {
		t.Fatalf("unvalidated id: got %s want %s", svc.unvalidated, id)
	}
}

func TestUnvalidateTimesheet_NotFinal(t *testing.T) {
	svc := &stubDeleteCRAService{err: domain.ErrCRANotFinal}
	rec := serveUnvalidateTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), uuid.New().String())
	if rec.Code != http.StatusConflict {
		t.Fatalf("status: got %d want 409", rec.Code)
	}
}

func TestUnvalidateTimesheet_InvoicedConflict(t *testing.T) {
	svc := &stubDeleteCRAService{err: domain.ErrCRAAlreadyInvoiced}
	rec := serveUnvalidateTimesheet(t, svc, stubAuthorizer{module: "cra", action: authx.ActionWrite, allow: true}, adminIdentity(), uuid.New().String())
	if rec.Code != http.StatusConflict {
		t.Fatalf("status: got %d want 409", rec.Code)
	}
}

type stubListRecentService struct {
	ports.CRAService
	lastFilter ports.TimesheetSummaryFilter
	manager    bool
}

func (s *stubListRecentService) ListTimesheetSummaries(_ context.Context, _ kernel.TenantID, _ ports.UserID, managerView bool, filter ports.TimesheetSummaryFilter) ([]domain.TimesheetSummary, error) {
	s.manager = managerView
	s.lastFilter = filter
	return []domain.TimesheetSummary{}, nil
}

func serveListRecent(t *testing.T, svc ports.CRAService, authorizer authx.Authorizer, identity authx.Identity, rawQuery string) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	r.Get("/timesheets/recent", listTimesheets(svc, authorizer))
	req := httptest.NewRequest(http.MethodGet, "/timesheets/recent?"+rawQuery, nil)
	req = req.WithContext(authx.WithIdentity(req.Context(), identity))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestListTimesheets_ParsesPeriodAndUserFilters(t *testing.T) {
	svc := &stubListRecentService{}
	target := uuid.New()

	rec := serveListRecent(t, svc, stubAuthorizerList{read: true, validate: true}, adminIdentity(), "year=2026&month=09&userId="+target.String()+"&limit=50")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200 body=%s", rec.Code, rec.Body.String())
	}
	if svc.lastFilter.Year != "2026" || svc.lastFilter.MonthMM != "09" {
		t.Fatalf("filter period: %+v", svc.lastFilter)
	}
	if svc.lastFilter.UserID == nil || *svc.lastFilter.UserID != target {
		t.Fatalf("filter userId: %+v", svc.lastFilter.UserID)
	}
	if svc.lastFilter.Limit != 50 {
		t.Fatalf("limit: got %d want 50", svc.lastFilter.Limit)
	}
	if !svc.manager {
		t.Fatal("expected managerView")
	}
}

func TestListTimesheets_UserFilterRequiresValidate(t *testing.T) {
	svc := &stubListRecentService{}
	rec := serveListRecent(t, svc, stubAuthorizerList{read: true, validate: false}, authx.Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Profile:  authx.ProfileCollaborateur,
	}, "userId="+uuid.New().String())
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d want 403", rec.Code)
	}
}

func TestListTimesheets_RejectsInvalidYearMonth(t *testing.T) {
	svc := &stubListRecentService{}
	authz := stubAuthorizerList{read: true, validate: true}
	id := adminIdentity()

	cases := []struct {
		query string
	}{
		{"year=abcd"},
		{"year=26"},
		{"month=00"},
		{"month=13"},
		{"month=2026-13"},
		{"month=September"},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			rec := serveListRecent(t, svc, authz, id, tc.query)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status: got %d want 400 for %q", rec.Code, tc.query)
			}
		})
	}
}

type stubAuthorizerList struct {
	read, validate bool
}

func (s stubAuthorizerList) Can(_ context.Context, module authx.Module, action authx.Action) bool {
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
