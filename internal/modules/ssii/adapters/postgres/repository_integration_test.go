//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	orgpostgres "github.com/kore/kore/internal/modules/org/adapters/postgres"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/ssii/adapters/postgres"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func seedMissionAppsFixture(t *testing.T) (
	repo *postgres.Repository,
	tenant kernel.TenantID,
	clientID, activeAppID, inactiveAppID uuid.UUID,
) {
	t.Helper()
	pool := dbtest.NewPostgres(t)
	orgRepo := orgpostgres.NewRepository(pool)
	repo = postgres.NewRepository(pool)
	ctx := context.Background()

	tenant = kernel.NewTenantID(uuid.New())
	require.NoError(t, orgRepo.SaveTenant(ctx, orgdomain.Tenant{ID: tenant.UUID(), Name: "SSII Apps"}))

	societeID := uuid.New()
	require.NoError(t, orgRepo.SaveSociete(ctx, orgdomain.Societe{
		ID: societeID, TenantID: tenant, RaisonSociale: "SSII SAS", Devise: "EUR",
	}))
	siteID := uuid.New()
	require.NoError(t, orgRepo.SaveSite(ctx, orgdomain.Site{
		ID: siteID, TenantID: tenant, SocieteID: societeID, Libelle: "HQ", Pays: "FR",
	}))

	activeAppID = uuid.New()
	require.NoError(t, orgRepo.SaveApplication(ctx, orgdomain.Application{
		ID: activeAppID, TenantID: tenant, Libelle: "App Active", Active: true, SiteIDs: []uuid.UUID{siteID},
	}))
	inactiveAppID = uuid.New()
	require.NoError(t, orgRepo.SaveApplication(ctx, orgdomain.Application{
		ID: inactiveAppID, TenantID: tenant, Libelle: "App Inactive", Active: false, SiteIDs: []uuid.UUID{siteID},
	}))

	clientID = uuid.New()
	require.NoError(t, orgRepo.SaveClient(ctx, orgdomain.Client{
		ID: clientID, TenantID: tenant, RaisonSociale: "Client Demo", Pays: "FR",
	}))

	return repo, tenant, clientID, activeAppID, inactiveAppID
}

func TestSSII_MissionApplications_RoundTripAndUnique(t *testing.T) {
	repo, tenant, clientID, activeAppID, _ := seedMissionAppsFixture(t)
	ctx := context.Background()

	mission := domain.NewMission(tenant, clientID, time.Now().UTC(), 45000)
	mission.Title = "Mission Apps"
	mission.Technologies = []string{}
	require.NoError(t, repo.CreateMissionWithRelations(ctx, mission, nil, []uuid.UUID{activeAppID}))

	apps, err := repo.ListMissionApplications(ctx, tenant, mission.ID)
	require.NoError(t, err)
	require.Len(t, apps, 1)
	require.Equal(t, activeAppID, apps[0].ApplicationID)
	require.True(t, apps[0].Active)

	require.NoError(t, repo.SaveMissionApplications(ctx, tenant, mission.ID, []uuid.UUID{activeAppID, activeAppID}))
	apps, err = repo.ListMissionApplications(ctx, tenant, mission.ID)
	require.NoError(t, err)
	require.Len(t, apps, 1)

	require.NoError(t, repo.SaveMissionApplications(ctx, tenant, mission.ID, nil))
	apps, err = repo.ListMissionApplications(ctx, tenant, mission.ID)
	require.NoError(t, err)
	require.Empty(t, apps)
}

func TestSSII_ValidateApplicationIDs_ActiveAndAlreadyLinked(t *testing.T) {
	repo, tenant, clientID, activeAppID, inactiveAppID := seedMissionAppsFixture(t)
	ctx := context.Background()

	valid, err := repo.ValidateApplicationIDs(ctx, tenant, []uuid.UUID{activeAppID, inactiveAppID}, uuid.Nil)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{activeAppID}, valid)

	mission := domain.NewMission(tenant, clientID, time.Now().UTC(), 45000)
	mission.Technologies = []string{}
	require.NoError(t, repo.CreateMissionWithRelations(ctx, mission, nil, []uuid.UUID{activeAppID}))

	// Persist inactive link at repo level (historical), then Validate must accept it for this mission.
	require.NoError(t, repo.SaveMissionApplications(ctx, tenant, mission.ID, []uuid.UUID{activeAppID, inactiveAppID}))

	valid, err = repo.ValidateApplicationIDs(ctx, tenant, []uuid.UUID{activeAppID, inactiveAppID}, mission.ID)
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{activeAppID, inactiveAppID}, valid)

	validCreate, err := repo.ValidateApplicationIDs(ctx, tenant, []uuid.UUID{inactiveAppID}, uuid.Nil)
	require.NoError(t, err)
	require.Empty(t, validCreate)
}
