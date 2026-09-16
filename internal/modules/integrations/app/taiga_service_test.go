package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type taigaRepoFake struct {
	links    []domain.ExternalLink
	mappings []domain.UserMapping
}

func (f *taigaRepoFake) UpsertExternalLink(_ context.Context, link domain.ExternalLink) error {
	f.links = append(f.links, link)
	return nil
}

func (f *taigaRepoFake) InsertApplicationProjectLink(_ context.Context, link domain.ExternalLink) error {
	for _, existing := range f.links {
		if existing.Provider == link.Provider &&
			existing.ExternalType == link.ExternalType &&
			existing.ExternalID == link.ExternalID &&
			existing.TenantID == link.TenantID {
			return domain.ErrTaigaProjectLinked
		}
	}
	f.links = append(f.links, link)
	return nil
}

func (f *taigaRepoFake) FindExternalLinkByKore(_ context.Context, _ kernel.TenantID, koreEntityType string, koreEntityID uuid.UUID) (domain.ExternalLink, error) {
	for _, link := range f.links {
		if link.KoreEntityType == koreEntityType && link.KoreEntityID == koreEntityID {
			return link, nil
		}
	}
	return domain.ExternalLink{}, domain.ErrExternalLinkNotFound
}

func (f *taigaRepoFake) FindExternalLinkByExternal(_ context.Context, tenant kernel.TenantID, provider, externalType, externalID string) (domain.ExternalLink, error) {
	for _, link := range f.links {
		if link.TenantID == tenant && link.Provider == provider && link.ExternalType == externalType && link.ExternalID == externalID {
			return link, nil
		}
	}
	return domain.ExternalLink{}, domain.ErrExternalLinkNotFound
}

func (f *taigaRepoFake) FindApplicationByTaigaProjectID(ctx context.Context, tenant kernel.TenantID, taigaProjectID string) (domain.ExternalLink, error) {
	return f.FindExternalLinkByExternal(ctx, tenant, "taiga", "project", taigaProjectID)
}

func (f *taigaRepoFake) ListApplicationProjectLinks(_ context.Context, tenant kernel.TenantID) ([]domain.ExternalLink, error) {
	var out []domain.ExternalLink
	for _, link := range f.links {
		if link.TenantID == tenant && link.Provider == "taiga" && link.ExternalType == "project" && link.KoreEntityType == "application" {
			out = append(out, link)
		}
	}
	return out, nil
}

func (f *taigaRepoFake) UpsertUserMapping(_ context.Context, mapping domain.UserMapping) error {
	f.mappings = append(f.mappings, mapping)
	return nil
}

func (f *taigaRepoFake) ListLinkedTaigaProjectIDs(_ context.Context, _ kernel.TenantID) ([]string, error) {
	ids := make([]string, 0)
	seen := make(map[string]struct{})
	for _, link := range f.links {
		if link.ExternalType != "project" || link.Provider != "taiga" {
			continue
		}
		if _, ok := seen[link.ExternalID]; ok {
			continue
		}
		seen[link.ExternalID] = struct{}{}
		ids = append(ids, link.ExternalID)
	}
	return ids, nil
}

func (f *taigaRepoFake) ListLinkedApplicationIDs(_ context.Context, _ kernel.TenantID) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0)
	for _, link := range f.links {
		if link.KoreEntityType == "application" && link.Provider == "taiga" {
			ids = append(ids, link.KoreEntityID)
		}
	}
	return ids, nil
}

func (f *taigaRepoFake) ListUserMappings(_ context.Context, tenant kernel.TenantID, provider string) ([]domain.UserMapping, error) {
	var out []domain.UserMapping
	for _, m := range f.mappings {
		if m.TenantID == tenant && m.Provider == provider {
			out = append(out, m)
		}
	}
	return out, nil
}

func TestUpsertUserMapping_ValidatesInput(t *testing.T) {
	svc := NewTaigaService(&taigaRepoFake{}, TaigaConfig{}, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())

	_, err := svc.UpsertUserMapping(context.Background(), UpsertUserMappingCommand{
		TenantID: tenant, TaigaUserID: 0, KoreUserID: uuid.New(), MatchMethod: "email",
	})
	require.ErrorIs(t, err, domain.ErrInvalidTaigaUserID)

	_, err = svc.UpsertUserMapping(context.Background(), UpsertUserMappingCommand{
		TenantID: tenant, TaigaUserID: 42, KoreUserID: uuid.Nil, MatchMethod: "email",
	})
	require.ErrorIs(t, err, domain.ErrInvalidKoreUserID)

	_, err = svc.UpsertUserMapping(context.Background(), UpsertUserMappingCommand{
		TenantID: tenant, TaigaUserID: 42, KoreUserID: uuid.New(), MatchMethod: " ",
	})
	require.ErrorIs(t, err, domain.ErrInvalidMatchMethod)

	_, err = svc.UpsertUserMapping(context.Background(), UpsertUserMappingCommand{
		TenantID: tenant, TaigaUserID: 42, KoreUserID: uuid.New(), MatchMethod: "foo",
	})
	require.ErrorIs(t, err, domain.ErrInvalidMatchMethod)
}

func TestUpsertUserMapping_HappyPath(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	koreUser := uuid.New()

	got, err := svc.UpsertUserMapping(context.Background(), UpsertUserMappingCommand{
		TenantID:      tenant,
		TaigaUserID:   99,
		TaigaUsername: "alice",
		KoreUserID:    koreUser,
		MatchMethod:   "email",
	})
	require.NoError(t, err)
	require.Equal(t, "taiga", got.Provider)
	require.Equal(t, "99", got.ExternalUserID)
	require.Len(t, repo.mappings, 1)
}

func TestListUserMappings_ReturnsTenantMappings(t *testing.T) {
	repo := &taigaRepoFake{}
	tenant := kernel.NewTenantID(uuid.New())
	other := kernel.NewTenantID(uuid.New())
	repo.mappings = []domain.UserMapping{
		{TenantID: tenant, Provider: "taiga", ExternalUserID: "1", ExternalUsername: "alice", KoreUserID: uuid.New(), MatchMethod: "email"},
		{TenantID: other, Provider: "taiga", ExternalUserID: "2", ExternalUsername: "bob", KoreUserID: uuid.New(), MatchMethod: "email"},
	}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, nil, nil)

	got, err := svc.ListUserMappings(context.Background(), tenant)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "alice", got[0].ExternalUsername)
}

func TestHandleWebhook_CreatesExternalLinkWithURL(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{
		BaseURL:     "https://tree.taiga.io",
		ProjectSlug: "kore-demo",
	}, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()

	body, err := json.Marshal(map[string]any{
		"action": "create",
		"type":   "userstory",
		"data": map[string]any{
			"id":                 123,
			"ref":                7,
			"project":            5,
			"external_reference": []any{"kore", demandID.String()},
		},
	})
	require.NoError(t, err)

	err = svc.HandleWebhook(context.Background(), tenant, body)
	require.NoError(t, err)
	require.Len(t, repo.links, 1)
	require.Equal(t, "https://tree.taiga.io/project/kore-demo/us/7", repo.links[0].ExternalURL)
}

func TestHandleWebhook_UsesPermalinkFromPayload(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()

	body, err := json.Marshal(map[string]any{
		"action": "create",
		"type":   "userstory",
		"data": map[string]any{
			"id":                 123,
			"ref":                7,
			"permalink":          "https://taiga.example/project/acme/us/7",
			"external_reference": []any{"kore", demandID.String()},
		},
	})
	require.NoError(t, err)

	err = svc.HandleWebhook(context.Background(), tenant, body)
	require.NoError(t, err)
	require.Equal(t, "https://taiga.example/project/acme/us/7", repo.links[0].ExternalURL)
}

func TestResolveTaigaExternalURL_ProjectSlugInPayload(t *testing.T) {
	data := map[string]any{
		"project": map[string]any{"slug": "from-payload"},
		"ref":     3,
	}
	ref := 3
	got := resolveTaigaExternalURL(data, TaigaConfig{BaseURL: "https://tree.taiga.io"}, &ref)
	require.Equal(t, "https://tree.taiga.io/project/from-payload/us/3", got)
}

func TestHandleWebhook_CreatesExternalLink(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()

	body, err := json.Marshal(map[string]any{
		"action": "create",
		"type":   "userstory",
		"data": map[string]any{
			"id":                 123,
			"ref":                7,
			"project":            5,
			"external_reference": []any{"kore", demandID.String()},
		},
	})
	require.NoError(t, err)

	err = svc.HandleWebhook(context.Background(), tenant, body)
	require.NoError(t, err)
	require.Len(t, repo.links, 1)
	require.Equal(t, demandID, repo.links[0].KoreEntityID)
	require.Equal(t, "123", repo.links[0].ExternalID)
}

func TestHandleWebhook_IgnoresUnrelatedPayload(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())

	body := []byte(`{"action":"create","type":"userstory","data":{"id":1}}`)
	err := svc.HandleWebhook(context.Background(), tenant, body)
	require.NoError(t, err)
	require.Empty(t, repo.links)
}

func TestHandleWebhook_InvalidJSON(t *testing.T) {
	svc := NewTaigaService(&taigaRepoFake{}, TaigaConfig{}, nil, nil, nil)
	err := svc.HandleWebhook(context.Background(), kernel.NewTenantID(uuid.New()), []byte("{"))
	require.Error(t, err)
}

type stubTaigaDemandGate struct {
	exists bool
}

func (g stubTaigaDemandGate) KoreDemandExists(context.Context, kernel.TenantID, uuid.UUID) (bool, error) {
	return g.exists, nil
}
func (g stubTaigaDemandGate) CreateDemandFromTaiga(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID, string, string) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (g stubTaigaDemandGate) GetDemandSummary(context.Context, kernel.TenantID, uuid.UUID) (ports.TaigaDemandSummary, error) {
	return ports.TaigaDemandSummary{}, nil
}
func (g stubTaigaDemandGate) ListDemandsByApplication(context.Context, kernel.TenantID, uuid.UUID) ([]ports.TaigaDemandSummary, error) {
	return nil, nil
}

func TestHandleWebhook_InvalidKoreDemandID(t *testing.T) {
	svc := NewTaigaService(&taigaRepoFake{}, TaigaConfig{}, nil, nil, nil)
	body, err := json.Marshal(map[string]any{
		"action": "create",
		"type":   "userstory",
		"data": map[string]any{
			"id":                 1,
			"external_reference": []any{"kore", "not-a-uuid"},
		},
	})
	require.NoError(t, err)
	err = svc.HandleWebhook(context.Background(), kernel.NewTenantID(uuid.New()), body)
	require.ErrorIs(t, err, domain.ErrInvalidKoreDemandID)
}

func TestHandleWebhook_KoreDemandNotFound(t *testing.T) {
	svc := NewTaigaService(&taigaRepoFake{}, TaigaConfig{}, stubTaigaDemandGate{exists: false}, nil, nil)
	demandID := uuid.New()
	body, err := json.Marshal(map[string]any{
		"action": "create",
		"type":   "userstory",
		"data": map[string]any{
			"id":                 1,
			"external_reference": []any{"kore", demandID.String()},
		},
	})
	require.NoError(t, err)
	err = svc.HandleWebhook(context.Background(), kernel.NewTenantID(uuid.New()), body)
	require.ErrorIs(t, err, domain.ErrKoreDemandNotFound)
}

func TestHandleWebhook_CreatesDemandFromIssue(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	authorID := uuid.New()
	createdDemand := uuid.New()
	repo := &taigaRepoFake{}
	repo.links = []domain.ExternalLink{{
		TenantID:       tenant,
		Provider:       "taiga",
		ExternalType:   "project",
		ExternalID:     "9",
		KoreEntityType: "application",
		KoreEntityID:   appID,
	}}
	repo.mappings = []domain.UserMapping{{
		TenantID:   tenant,
		Provider:   "taiga",
		KoreUserID: authorID,
	}}
	gate := &recordingDemandGate{createID: createdDemand}
	svc := NewTaigaService(repo, TaigaConfig{BaseURL: "https://taiga.example"}, gate, nil, nil)
	body, err := json.Marshal(map[string]any{
		"action": "create",
		"type":   "issue",
		"data": map[string]any{
			"id": 55, "ref": 3, "project": 9, "subject": "Bug", "description": "d", "version": 1,
		},
	})
	require.NoError(t, err)
	require.NoError(t, svc.HandleWebhook(context.Background(), tenant, body))
	require.Equal(t, 1, gate.createCalls)
	require.True(t, len(repo.links) >= 2)
}

type recordingDemandGate struct {
	createID    uuid.UUID
	createCalls int
}

func (g *recordingDemandGate) KoreDemandExists(context.Context, kernel.TenantID, uuid.UUID) (bool, error) {
	return true, nil
}
func (g *recordingDemandGate) CreateDemandFromTaiga(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID, string, string) (uuid.UUID, error) {
	g.createCalls++
	return g.createID, nil
}
func (g *recordingDemandGate) GetDemandSummary(context.Context, kernel.TenantID, uuid.UUID) (ports.TaigaDemandSummary, error) {
	return ports.TaigaDemandSummary{ID: g.createID, Subject: "s"}, nil
}
func (g *recordingDemandGate) ListDemandsByApplication(context.Context, kernel.TenantID, uuid.UUID) ([]ports.TaigaDemandSummary, error) {
	return nil, nil
}

func TestHandleWebhook_IgnoresExternalReferenceOnlyChange(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{}, stubTaigaDemandGate{exists: true}, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()
	body, err := json.Marshal(map[string]any{
		"action": "change",
		"type":   "issue",
		"data": map[string]any{
			"id":                 9,
			"external_reference": []any{"kore", demandID.String()},
			"values_diff":        map[string]any{"external_reference": []any{nil, []any{"kore", demandID.String()}}},
		},
	})
	require.NoError(t, err)
	require.NoError(t, svc.HandleWebhook(context.Background(), tenant, body))
	require.Empty(t, repo.links)
}

func TestPullIssueToKore_DoesNotCreateExtraIssueViaGateway(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	authorID := uuid.New()
	createdDemand := uuid.New()
	repo := &taigaRepoFake{}
	repo.links = []domain.ExternalLink{{
		TenantID: tenant, Provider: "taiga", ExternalType: "project", ExternalID: "9",
		KoreEntityType: "application", KoreEntityID: appID,
	}}
	gw := &recordingGateway{created: ports.TaigaIssue{ID: 999, Ref: 1, ProjectID: 9}}
	gate := &recordingDemandGate{createID: createdDemand}
	svc := NewTaigaService(repo, TaigaConfig{}, gate, gw, nil)
	_, err := svc.PullIssueToKore(context.Background(), tenant, ports.TaigaIssue{
		ID: 55, Ref: 3, ProjectID: 9, Subject: "Bug", Version: 1,
	}, authorID)
	require.NoError(t, err)
	require.Equal(t, 0, gw.createCalls, "pull must not CreateIssue (auto-push)")
	require.Equal(t, 1, gate.createCalls)
}

func TestPushDemandToTaiga(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()
	appID := uuid.New()
	repo := &taigaRepoFake{}
	repo.links = []domain.ExternalLink{{
		TenantID: tenant, Provider: "taiga", ExternalType: "project", ExternalID: "4",
		KoreEntityType: "application", KoreEntityID: appID,
	}}
	gw := &recordingGateway{created: ports.TaigaIssue{ID: 77, Ref: 8, ProjectID: 4, Version: 1, Permalink: "https://t/i/8"}}
	gateSummary := &pushDemandGate{summary: ports.TaigaDemandSummary{
		ID: demandID, ApplicationID: appID, Subject: "Sub", Description: "Desc",
	}}
	svc := NewTaigaService(repo, TaigaConfig{}, gateSummary, gw, nil)
	link, err := svc.PushDemandToTaiga(context.Background(), tenant, demandID)
	require.NoError(t, err)
	require.Equal(t, "77", link.ExternalID)
	require.Equal(t, 1, gw.createCalls)
}

type pushDemandGate struct {
	summary ports.TaigaDemandSummary
}

func (g *pushDemandGate) KoreDemandExists(context.Context, kernel.TenantID, uuid.UUID) (bool, error) {
	return true, nil
}
func (g *pushDemandGate) CreateDemandFromTaiga(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID, string, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (g *pushDemandGate) GetDemandSummary(context.Context, kernel.TenantID, uuid.UUID) (ports.TaigaDemandSummary, error) {
	return g.summary, nil
}
func (g *pushDemandGate) ListDemandsByApplication(context.Context, kernel.TenantID, uuid.UUID) ([]ports.TaigaDemandSummary, error) {
	return nil, nil
}

type recordingGateway struct {
	created     ports.TaigaIssue
	createCalls int
}

func (g *recordingGateway) ListProjects(context.Context) ([]ports.TaigaProject, error) {
	return nil, nil
}
func (g *recordingGateway) CreateIssue(_ context.Context, projectID int, subject, _ string, ref []string) (ports.TaigaIssue, error) {
	g.createCalls++
	out := g.created
	out.ProjectID = projectID
	out.Subject = subject
	out.ExternalReference = ref
	return out, nil
}
func (g *recordingGateway) ListProjectIssues(context.Context, int) ([]ports.TaigaIssue, error) {
	return nil, nil
}
func (g *recordingGateway) UpdateIssueExternalReference(context.Context, int, int, []string) (ports.TaigaIssue, error) {
	return g.created, nil
}
