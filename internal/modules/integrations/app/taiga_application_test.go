package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type stubTaigaGateway struct {
	projects []ports.TaigaProject
	err      error
}

func (g stubTaigaGateway) ListProjects(context.Context) ([]ports.TaigaProject, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.projects, nil
}

type stubApplicationCreator struct {
	created []ports.CreateApplicationInput
}

func (s *stubApplicationCreator) CreateApplication(_ context.Context, cmd ports.CreateApplicationInput) (ports.ApplicationSummary, error) {
	s.created = append(s.created, cmd)
	if cmd.Libelle == "fail" {
		return ports.ApplicationSummary{}, fmt.Errorf("create failed")
	}
	return ports.ApplicationSummary{ID: uuid.New(), Libelle: cmd.Libelle}, nil
}

func (s *stubApplicationCreator) DeactivateApplication(context.Context, kernel.TenantID, uuid.UUID) error {
	return nil
}

func TestListUnlinkedProjects_FiltersLinked(t *testing.T) {
	repo := &taigaRepoFake{}
	repo.links = []domain.ExternalLink{{
		Provider:     "taiga",
		ExternalType: "project",
		ExternalID:   "2",
	}}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, stubTaigaGateway{projects: []ports.TaigaProject{
		{ID: 1, Name: "A", Slug: "a"},
		{ID: 2, Name: "B", Slug: "b"},
	}}, nil)
	tenant := kernel.NewTenantID(uuid.New())

	out, err := svc.ListUnlinkedProjects(context.Background(), tenant)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, 1, out[0].ID)
}

func TestImportApplications_PartialErrors(t *testing.T) {
	repo := &taigaRepoFake{}
	apps := &stubApplicationCreator{}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, stubTaigaGateway{projects: []ports.TaigaProject{
		{ID: 1, Name: "OK", Slug: "ok"},
		{ID: 2, Name: "fail", Slug: "fail"},
	}}, apps)
	tenant := kernel.NewTenantID(uuid.New())

	result, err := svc.ImportApplications(context.Background(), ImportApplicationsCommand{
		TenantID: tenant,
		Projects: []ImportApplicationProject{
			{TaigaProjectID: 1, Libelle: "OK"},
			{TaigaProjectID: 2, Libelle: "fail"},
		},
		SiteIDs: []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)
	require.Len(t, result.Created, 1)
	require.Len(t, result.Errors, 1)
	require.Equal(t, 2, result.Errors[0].TaigaProjectID)
}

func TestCreateApplicationWithTaiga_LinksProject(t *testing.T) {
	repo := &taigaRepoFake{}
	apps := &stubApplicationCreator{}
	svc := NewTaigaService(repo, TaigaConfig{BaseURL: "https://taiga.example.com"}, nil, stubTaigaGateway{projects: []ports.TaigaProject{
		{ID: 5, Name: "Kore", Slug: "kore"},
	}}, apps)
	tenant := kernel.NewTenantID(uuid.New())

	got, err := svc.CreateApplicationWithTaiga(context.Background(), ports.CreateApplicationInput{
		TenantID: tenant,
		Libelle:  "Kore",
		SiteIDs:  []uuid.UUID{uuid.New()},
	}, 5)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, got.ID)
	require.Len(t, repo.links, 1)
	require.Equal(t, "application", repo.links[0].KoreEntityType)
	require.Equal(t, "5", repo.links[0].ExternalID)
}

func TestImportApplications_TaigaUnavailable(t *testing.T) {
	svc := NewTaigaService(&taigaRepoFake{}, TaigaConfig{}, nil, stubTaigaGateway{err: domain.ErrTaigaUnavailable}, &stubApplicationCreator{})
	_, err := svc.ImportApplications(context.Background(), ImportApplicationsCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		Projects: []ImportApplicationProject{{TaigaProjectID: 1}},
		SiteIDs:  []uuid.UUID{uuid.New()},
	})
	require.ErrorIs(t, err, domain.ErrTaigaUnavailable)
}

func TestImportApplications_DuplicateProjectRejected(t *testing.T) {
	repo := &taigaRepoFake{}
	repo.links = []domain.ExternalLink{{
		Provider:     "taiga",
		ExternalType: "project",
		ExternalID:   "1",
		TenantID:     kernel.NewTenantID(uuid.New()),
	}}
	apps := &stubApplicationCreator{}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, stubTaigaGateway{projects: []ports.TaigaProject{
		{ID: 1, Name: "Taken", Slug: "taken"},
	}}, apps)
	tenant := repo.links[0].TenantID

	result, err := svc.ImportApplications(context.Background(), ImportApplicationsCommand{
		TenantID: tenant,
		Projects: []ImportApplicationProject{{TaigaProjectID: 1, Libelle: "Taken"}},
		SiteIDs:  []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)
	require.Empty(t, result.Created)
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0].Message, domain.ErrTaigaProjectLinked.Error())
}

func TestLinkExistingApplication_LinksProject(t *testing.T) {
	repo := &taigaRepoFake{}
	appID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	svc := NewTaigaService(repo, TaigaConfig{BaseURL: "https://taiga.example.com"}, nil, stubTaigaGateway{projects: []ports.TaigaProject{
		{ID: 3, Name: "Legacy", Slug: "legacy"},
	}}, nil)

	err := svc.LinkExistingApplication(context.Background(), tenant, appID, 3)
	require.NoError(t, err)
	require.Len(t, repo.links, 1)
	require.Equal(t, appID, repo.links[0].KoreEntityID)
	require.Equal(t, "3", repo.links[0].ExternalID)
}

func TestLinkExistingApplication_AlreadyLinked(t *testing.T) {
	repo := &taigaRepoFake{}
	appID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	repo.links = []domain.ExternalLink{{
		Provider:       "taiga",
		ExternalType:   "project",
		ExternalID:     "1",
		KoreEntityType: "application",
		KoreEntityID:   appID,
		TenantID:       tenant,
	}}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, stubTaigaGateway{projects: []ports.TaigaProject{
		{ID: 2, Name: "Other", Slug: "other"},
	}}, nil)

	err := svc.LinkExistingApplication(context.Background(), tenant, appID, 2)
	require.ErrorIs(t, err, domain.ErrTaigaApplicationAlreadyLinked)
}

func TestLinkExistingApplication_ProjectNotFound(t *testing.T) {
	repo := &taigaRepoFake{}
	svc := NewTaigaService(repo, TaigaConfig{}, nil, stubTaigaGateway{projects: []ports.TaigaProject{}}, nil)

	err := svc.LinkExistingApplication(context.Background(), kernel.NewTenantID(uuid.New()), uuid.New(), 99)
	require.ErrorIs(t, err, domain.ErrTaigaProjectNotFound)
}
