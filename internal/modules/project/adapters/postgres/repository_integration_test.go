//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	orgpostgres "github.com/kore/kore/internal/modules/org/adapters/postgres"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/project/adapters/postgres"
	"github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func seedAgileApp(t *testing.T, orgRepo *orgpostgres.Repository, tenant kernel.TenantID) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	require.NoError(t, orgRepo.SaveTenant(ctx, orgdomain.Tenant{ID: tenant.UUID(), Name: "Project IT"}))

	societeID := uuid.New()
	require.NoError(t, orgRepo.SaveSociete(ctx, orgdomain.Societe{
		ID:            societeID,
		TenantID:      tenant,
		RaisonSociale: "Agile SAS",
		Devise:        "EUR",
	}))

	siteID := uuid.New()
	require.NoError(t, orgRepo.SaveSite(ctx, orgdomain.Site{
		ID:        siteID,
		TenantID:  tenant,
		SocieteID: societeID,
		Libelle:   "HQ",
		Pays:      "FR",
	}))

	serviceID := uuid.New()
	resp := uuid.New()
	require.NoError(t, orgRepo.SaveService(ctx, orgdomain.Service{
		ID:            serviceID,
		TenantID:      tenant,
		SiteID:        siteID,
		Libelle:       "Dev",
		Type:          "interne",
		ResponsableID: &resp,
	}))

	appID := uuid.New()
	require.NoError(t, orgRepo.SaveApplication(ctx, orgdomain.Application{
		ID:                 appID,
		TenantID:           tenant,
		Libelle:            "Agile product",
		Active:             true,
		ServiceIDs:         []uuid.UUID{serviceID},
		MethodologyProfile: orgdomain.MethodologyAgileScrum,
	}))
	return appID
}

func insertSprintDemand(t *testing.T, pool *db.Pool, tenant kernel.TenantID, appID, sprintID, authorID uuid.UUID, subject string, storyPoints int16, resolvedAt *time.Time) {
	t.Helper()
	ctx := context.Background()
	status := "en_cours"
	if resolvedAt != nil {
		status = "resolue"
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO tma.demands (
			id, tenant_id, application_id, type, subject, description, priority,
			author_id, status, visible, consumption_active, requires_chef_gate,
			sprint_id, story_points, resolved_at, created_at
		) VALUES (
			$1, $2, $3, 'incident', $4, '', 'normal',
			$5, $6, TRUE, TRUE, FALSE,
			$7, $8, $9, NOW()
		)
	`, uuid.New(), tenant.UUID(), appID, subject, authorID, status, sprintID, storyPoints, resolvedAt)
	require.NoError(t, err)
}

func TestProject_EpicSprintRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	orgRepo := orgpostgres.NewRepository(pool)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	appID := seedAgileApp(t, orgRepo, tenant)

	epic := domain.NewEpic(tenant, appID, "Epic A", "desc", kernel.PriorityHigh)
	require.NoError(t, repo.SaveEpic(ctx, epic))

	got, err := repo.GetEpic(ctx, tenant, appID, epic.ID)
	require.NoError(t, err)
	require.Equal(t, epic.Title, got.Title)

	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 13)
	sprint := domain.NewSprint(tenant, appID, "S1", "Goal", start, end, nil)
	require.NoError(t, repo.SaveSprint(ctx, sprint))

	sprints, err := repo.ListSprints(ctx, tenant, appID)
	require.NoError(t, err)
	require.Len(t, sprints, 1)
	require.Equal(t, domain.SprintStatusPlanned, sprints[0].Status)
}

func TestProject_SprintBurndownResolvedAt(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	orgRepo := orgpostgres.NewRepository(pool)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	appID := seedAgileApp(t, orgRepo, tenant)
	authorID := uuid.New()

	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -3)
	end := start.AddDate(0, 0, 6)
	sprint := domain.NewSprint(tenant, appID, "Burndown", "", start, end, nil)
	require.NoError(t, sprint.Start())
	require.NoError(t, repo.SaveSprint(ctx, sprint))

	day1 := start.AddDate(0, 0, 1)
	day3 := start.AddDate(0, 0, 3)
	insertSprintDemand(t, pool, tenant, appID, sprint.ID, authorID, "US-1", 5, &day1)
	insertSprintDemand(t, pool, tenant, appID, sprint.ID, authorID, "US-2", 8, &day3)
	insertSprintDemand(t, pool, tenant, appID, sprint.ID, authorID, "US-3", 3, nil)

	series, err := repo.GetSprintBurndown(ctx, tenant, sprint)
	require.NoError(t, err)
	require.Equal(t, 16, series.PlannedPoints)
	require.GreaterOrEqual(t, len(series.Points), 4)

	require.Equal(t, 16, series.Points[0].RemainingPoints)
	require.Equal(t, 11, series.Points[1].RemainingPoints)
	require.Equal(t, 3, series.Points[3].RemainingPoints)
}

func TestProject_ActiveSprintUnique(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	orgRepo := orgpostgres.NewRepository(pool)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	appID := seedAgileApp(t, orgRepo, tenant)

	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	s1 := domain.NewSprint(tenant, appID, "S1", "", start, start.AddDate(0, 0, 13), nil)
	s2 := domain.NewSprint(tenant, appID, "S2", "", start.AddDate(0, 0, 14), start.AddDate(0, 0, 27), nil)
	require.NoError(t, s1.Start())
	require.NoError(t, repo.SaveSprint(ctx, s1))
	require.NoError(t, s2.Start())
	err := repo.SaveSprint(ctx, s2)
	require.Error(t, err)
}
