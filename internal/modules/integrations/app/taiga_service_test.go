package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
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

func (f *taigaRepoFake) FindExternalLinkByKore(_ context.Context, _ kernel.TenantID, koreEntityType string, koreEntityID uuid.UUID) (domain.ExternalLink, error) {
	for _, link := range f.links {
		if link.KoreEntityType == koreEntityType && link.KoreEntityID == koreEntityID {
			return link, nil
		}
	}
	return domain.ExternalLink{}, domain.ErrExternalLinkNotFound
}

func (f *taigaRepoFake) UpsertUserMapping(_ context.Context, mapping domain.UserMapping) error {
	f.mappings = append(f.mappings, mapping)
	return nil
}

func TestUpsertUserMapping_ValidatesInput(t *testing.T) {
	svc := NewTaigaService(&taigaRepoFake{}, TaigaConfig{})
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
}

func TestUpsertUserMapping_HappyPath(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{})
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

func TestHandleWebhook_CreatesExternalLinkWithURL(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{
		BaseURL:     "https://tree.taiga.io",
		ProjectSlug: "kore-demo",
	})
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
	svc := NewTaigaService(repo, TaigaConfig{})
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
	svc := NewTaigaService(repo, TaigaConfig{})
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
	svc := NewTaigaService(repo, TaigaConfig{})
	tenant := kernel.NewTenantID(uuid.New())

	body := []byte(`{"action":"create","type":"userstory","data":{"id":1}}`)
	err := svc.HandleWebhook(context.Background(), tenant, body)
	require.NoError(t, err)
	require.Empty(t, repo.links)
}

func TestHandleWebhook_InvalidJSON(t *testing.T) {
	svc := NewTaigaService(&taigaRepoFake{}, TaigaConfig{})
	err := svc.HandleWebhook(context.Background(), kernel.NewTenantID(uuid.New()), []byte("{"))
	require.Error(t, err)
}
