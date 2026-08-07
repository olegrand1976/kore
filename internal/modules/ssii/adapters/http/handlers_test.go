package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/pkg/kernel"
)

type stubAuthorizer struct {
	module authx.Module
	action authx.Action
	allow  bool
}

func (s stubAuthorizer) Can(_ context.Context, module authx.Module, action authx.Action) bool {
	return s.allow && module == s.module && action == s.action
}

type missionServiceStub struct {
	ports.SSIIService
	created *ports.CreateMissionCommand
	updated *ports.UpdateApplicationsCommand
	detail  ports.MissionDetail
	createErr error
	updateErr error
}

func (s *missionServiceStub) Create(_ context.Context, cmd ports.CreateMissionCommand) (domain.Mission, error) {
	if s.createErr != nil {
		return domain.Mission{}, s.createErr
	}
	s.created = &cmd
	return domain.Mission{
		ID:        uuid.New(),
		TenantID:  cmd.TenantID,
		ClientID:  cmd.ClientID,
		StartDate: cmd.StartDate,
		Title:     cmd.Title,
		RateUnit:  domain.RateUnitTJM,
		TJMAmount: cmd.TJMAmount,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *missionServiceStub) UpdateApplications(_ context.Context, cmd ports.UpdateApplicationsCommand) (ports.MissionDetail, error) {
	if s.updateErr != nil {
		return ports.MissionDetail{}, s.updateErr
	}
	s.updated = &cmd
	detail := s.detail
	detail.ID = cmd.MissionID
	detail.Applications = make([]ports.MissionApplication, 0, len(cmd.ApplicationIDs))
	for _, id := range cmd.ApplicationIDs {
		detail.Applications = append(detail.Applications, ports.MissionApplication{
			ApplicationID: id,
			Libelle:       id.String(),
			Active:        true,
		})
	}
	return detail, nil
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

func withMissionID(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateMission_forwardsApplicationIDs(t *testing.T) {
	svc := &missionServiceStub{}
	handler := createMission(svc, stubAuthorizer{module: "ssii", action: authx.ActionWrite, allow: true})

	clientID := uuid.New()
	appID := uuid.New()
	collabID := uuid.New()
	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/missions", map[string]any{
		"clientId":         clientID.String(),
		"startDate":        time.Now().UTC().Format(time.RFC3339),
		"rateUnit":         "tjm",
		"tjmAmount":        50000,
		"collaboratorIds":  []string{collabID.String()},
		"applicationIds":   []string{appID.String()},
	}))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", rec.Code, rec.Body.String())
	}
	if svc.created == nil {
		t.Fatal("expected Create to be called")
	}
	if len(svc.created.ApplicationIDs) != 1 || svc.created.ApplicationIDs[0] != appID {
		t.Fatalf("applicationIds = %v, want [%s]", svc.created.ApplicationIDs, appID)
	}
}

func TestCreateMission_rejectsInvalidApplication(t *testing.T) {
	svc := &missionServiceStub{createErr: domain.ErrInvalidApplication}
	handler := createMission(svc, stubAuthorizer{module: "ssii", action: authx.ActionWrite, allow: true})

	rec := httptest.NewRecorder()
	handler(rec, requestWithIdentity(t, http.MethodPost, "/missions", map[string]any{
		"clientId":        uuid.New().String(),
		"startDate":       time.Now().UTC().Format(time.RFC3339),
		"collaboratorIds": []string{uuid.New().String()},
		"applicationIds":  []string{uuid.New().String()},
	}))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
}

func TestUpdateApplications_forbiddenWithoutWrite(t *testing.T) {
	svc := &missionServiceStub{}
	handler := updateApplications(svc, stubAuthorizer{module: "ssii", action: authx.ActionRead, allow: true})

	missionID := uuid.New()
	rec := httptest.NewRecorder()
	req := withMissionID(requestWithIdentity(t, http.MethodPut, "/missions/"+missionID.String()+"/applications", map[string]any{
		"applicationIds": []string{uuid.New().String()},
	}), missionID.String())
	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if svc.updated != nil {
		t.Fatal("expected no update when ssii:E missing")
	}
}

func TestUpdateApplications_replacesLinks(t *testing.T) {
	svc := &missionServiceStub{}
	handler := updateApplications(svc, stubAuthorizer{module: "ssii", action: authx.ActionWrite, allow: true})

	missionID := uuid.New()
	appID := uuid.New()
	rec := httptest.NewRecorder()
	req := withMissionID(requestWithIdentity(t, http.MethodPut, "/missions/"+missionID.String()+"/applications", map[string]any{
		"applicationIds": []string{appID.String()},
	}), missionID.String())
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if svc.updated == nil {
		t.Fatal("expected UpdateApplications")
	}
	if svc.updated.MissionID != missionID {
		t.Fatalf("missionId = %s", svc.updated.MissionID)
	}
	if len(svc.updated.ApplicationIDs) != 1 || svc.updated.ApplicationIDs[0] != appID {
		t.Fatalf("applicationIds = %v", svc.updated.ApplicationIDs)
	}
}

func TestUpdateApplications_invalidID(t *testing.T) {
	handler := updateApplications(&missionServiceStub{}, stubAuthorizer{module: "ssii", action: authx.ActionWrite, allow: true})
	rec := httptest.NewRecorder()
	req := withMissionID(requestWithIdentity(t, http.MethodPut, "/missions/not-a-uuid/applications", map[string]any{
		"applicationIds": []string{},
	}), "not-a-uuid")
	handler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUpdateApplications_mapsInvalidApplication(t *testing.T) {
	svc := &missionServiceStub{updateErr: domain.ErrInvalidApplication}
	handler := updateApplications(svc, stubAuthorizer{module: "ssii", action: authx.ActionWrite, allow: true})

	missionID := uuid.New()
	rec := httptest.NewRecorder()
	req := withMissionID(requestWithIdentity(t, http.MethodPut, "/missions/"+missionID.String()+"/applications", map[string]any{
		"applicationIds": []string{uuid.New().String()},
	}), missionID.String())
	handler(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	var env httpx.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error == nil || env.Error.Code != httpx.ErrCodeValidation {
		t.Fatalf("envelope = %+v", env.Error)
	}
}
