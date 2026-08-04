//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/adapters/postgres"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

// seedStructure crée la chaîne minimale Société → Site → Service → Application
// exigée par les clés étrangères de org.equipes.
func seedStructure(t *testing.T, repo *postgres.Repository, tenant kernel.TenantID) (siteID, serviceID, appID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	require.NoError(t, repo.SaveTenant(ctx, domain.Tenant{ID: tenant.UUID(), Name: "Structure"}))

	societeID := uuid.New()
	require.NoError(t, repo.SaveSociete(ctx, domain.Societe{
		ID:            societeID,
		TenantID:      tenant,
		RaisonSociale: "Structure SAS",
		Devise:        "EUR",
	}))

	siteID = uuid.New()
	require.NoError(t, repo.SaveSite(ctx, domain.Site{
		ID:        siteID,
		TenantID:  tenant,
		SocieteID: societeID,
		Libelle:   "Paris HQ",
		Pays:      "FR",
	}))

	responsable := uuid.New()
	serviceID = uuid.New()
	require.NoError(t, repo.SaveService(ctx, domain.Service{
		ID:            serviceID,
		TenantID:      tenant,
		SiteID:        siteID,
		Libelle:       "Delivery",
		Type:          "interne",
		ResponsableID: &responsable,
	}))

	appID = uuid.New()
	require.NoError(t, repo.SaveApplication(ctx, domain.Application{
		ID:        appID,
		TenantID:  tenant,
		ServiceID: serviceID,
		Libelle:   "Portail Client",
	}))

	return siteID, serviceID, appID
}

func TestOrg_SaveEquipeRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	_, _, appID := seedStructure(t, repo, tenant)

	responsable := uuid.New()
	equipeID := uuid.New()
	require.NoError(t, repo.SaveEquipe(ctx, domain.Equipe{
		ID:            equipeID,
		TenantID:      tenant,
		ApplicationID: appID,
		Libelle:       "Équipe Dev",
		ResponsableID: &responsable,
	}))

	equipes, err := repo.ListEquipes(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, equipes, 1)
	require.Equal(t, equipeID, equipes[0].ID)
	require.Equal(t, appID, equipes[0].ApplicationID)
	require.Equal(t, "Équipe Dev", equipes[0].Libelle)
	require.NotNil(t, equipes[0].ResponsableID)
	require.Equal(t, responsable, *equipes[0].ResponsableID)

	// Isolation multi-tenant : une autre organisation ne voit pas cette équipe.
	other := kernel.NewTenantID(uuid.New())
	require.NoError(t, repo.SaveTenant(ctx, domain.Tenant{ID: other.UUID(), Name: "Autre"}))
	otherEquipes, err := repo.ListEquipes(ctx, other)
	require.NoError(t, err)
	require.Empty(t, otherEquipes)
}

func TestOrg_SaveEquipeWithoutResponsable(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	_, _, appID := seedStructure(t, repo, tenant)

	require.NoError(t, repo.SaveEquipe(ctx, domain.Equipe{
		ID:            uuid.New(),
		TenantID:      tenant,
		ApplicationID: appID,
		Libelle:       "Équipe sans responsable",
	}))

	equipes, err := repo.ListEquipes(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, equipes, 1)
	require.Nil(t, equipes[0].ResponsableID)
}

func TestOrg_ListSites(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	siteID, _, _ := seedStructure(t, repo, tenant)

	sites, err := repo.ListSites(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, sites, 1)
	require.Equal(t, siteID, sites[0].ID)
	require.Equal(t, "Paris HQ", sites[0].Libelle)
	require.Equal(t, "FR", sites[0].Pays)
	require.NotEqual(t, uuid.Nil, sites[0].SocieteID)
}

// Couvre la migration 0018 : org.services.libelle doit être persisté et relu.
func TestOrg_ListServicesCarriesLibelleAndType(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	siteID, serviceID, _ := seedStructure(t, repo, tenant)

	services, err := repo.ListServices(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, services, 1)
	require.Equal(t, serviceID, services[0].ID)
	require.Equal(t, siteID, services[0].SiteID)
	require.Equal(t, "Paris HQ", services[0].SiteLabel)
	require.Equal(t, "Delivery", services[0].Libelle)
	require.Equal(t, "interne", services[0].Type)
	require.NotNil(t, services[0].ResponsableID)
}

func TestOrg_SaveServiceDefaultsTypeWhenEmpty(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	siteID, _, _ := seedStructure(t, repo, tenant)

	responsable := uuid.New()
	require.NoError(t, repo.SaveService(ctx, domain.Service{
		ID:            uuid.New(),
		TenantID:      tenant,
		SiteID:        siteID,
		Libelle:       "Sans type",
		ResponsableID: &responsable,
	}))

	services, err := repo.ListServices(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, services, 2)
	for _, s := range services {
		require.NotEmpty(t, s.Type, "chaque service doit porter un type")
	}
}

func TestOrg_UpdateUserPersistsEquipe(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	_, _, appID := seedStructure(t, repo, tenant)

	equipeID := uuid.New()
	require.NoError(t, repo.SaveEquipe(ctx, domain.Equipe{
		ID:            equipeID,
		TenantID:      tenant,
		ApplicationID: appID,
		Libelle:       "Équipe Dev",
	}))

	userID := uuid.New()
	require.NoError(t, repo.SaveUser(ctx, domain.User{
		ID:           userID,
		TenantID:     tenant,
		Login:        "COL_collab",
		PasswordHash: "hash",
		Profile:      domain.ProfileCollaborateur,
		Active:       true,
	}))

	stored, err := repo.FindUserByID(ctx, tenant, userID)
	require.NoError(t, err)
	require.Nil(t, stored.EquipeID)

	// Rattachement.
	stored.EquipeID = &equipeID
	require.NoError(t, repo.UpdateUser(ctx, stored))
	attached, err := repo.FindUserByID(ctx, tenant, userID)
	require.NoError(t, err)
	require.NotNil(t, attached.EquipeID)
	require.Equal(t, equipeID, *attached.EquipeID)

	// Détachement.
	attached.EquipeID = nil
	require.NoError(t, repo.UpdateUser(ctx, attached))
	detached, err := repo.FindUserByID(ctx, tenant, userID)
	require.NoError(t, err)
	require.Nil(t, detached.EquipeID)
}
