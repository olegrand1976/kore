package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	projectapp "github.com/kore/kore/internal/modules/project/app"
	"github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/internal/modules/project/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	sprints map[uuid.UUID]domain.Sprint
}

func (f *fakeRepo) SaveEpic(ctx context.Context, epic domain.Epic) error { return nil }
func (f *fakeRepo) GetEpic(ctx context.Context, tenant kernel.TenantID, appID, id uuid.UUID) (domain.Epic, error) {
	return domain.Epic{}, domain.ErrEpicNotFound
}
func (f *fakeRepo) ListEpics(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Epic, error) {
	return nil, nil
}
func (f *fakeRepo) SaveSprint(ctx context.Context, sprint domain.Sprint) error {
	if f.sprints == nil {
		f.sprints = map[uuid.UUID]domain.Sprint{}
	}
	f.sprints[sprint.ID] = sprint
	return nil
}
func (f *fakeRepo) GetSprint(ctx context.Context, tenant kernel.TenantID, appID, id uuid.UUID) (domain.Sprint, error) {
	s, ok := f.sprints[id]
	if !ok {
		return domain.Sprint{}, domain.ErrSprintNotFound
	}
	return s, nil
}
func (f *fakeRepo) ListSprints(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Sprint, error) {
	return nil, nil
}
func (f *fakeRepo) GetActiveSprint(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (*domain.Sprint, error) {
	return nil, nil
}
func (f *fakeRepo) SaveKanbanConfig(ctx context.Context, cfg domain.KanbanConfig) error { return nil }
func (f *fakeRepo) GetKanbanConfig(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.KanbanConfig, error) {
	return domain.KanbanConfig{}, nil
}
func (f *fakeRepo) AssignDemandsToSprint(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID, demandIDs []uuid.UUID) error {
	return nil
}
func (f *fakeRepo) ReorderBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, demandIDs []uuid.UUID) error {
	return nil
}
func (f *fakeRepo) ListBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, backlogOnly bool) ([]domain.BacklogItem, error) {
	return nil, nil
}
func (f *fakeRepo) CountProjectArtifacts(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (int, error) {
	return 0, nil
}
func (f *fakeRepo) GetSprintBurndown(ctx context.Context, tenant kernel.TenantID, sprint domain.Sprint) (domain.BurndownSeries, error) {
	return domain.BurndownSeries{}, nil
}
func (f *fakeRepo) GetVelocity(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, lastN int) (domain.VelocityReport, error) {
	return domain.VelocityReport{}, nil
}
func (f *fakeRepo) ListAgileApplications(ctx context.Context, tenant kernel.TenantID) ([]orgdomain.Application, error) {
	return nil, nil
}

type fakeApps struct {
	profile orgdomain.MethodologyProfile
}

func (f fakeApps) GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (orgdomain.Application, error) {
	return orgdomain.Application{
		ID:                 id,
		TenantID:           tenant,
		MethodologyProfile: f.profile,
	}, nil
}

func TestProjectService_PlanSprint(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{}
	svc := projectapp.NewService(repo, fakeApps{profile: orgdomain.MethodologyAgileScrum})
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	sprintID := uuid.New()
	repo.sprints = map[uuid.UUID]domain.Sprint{
		sprintID: domain.NewSprint(tenant, appID, "S1", "", time.Now(), time.Now().AddDate(0, 0, 14), nil),
	}
	err := svc.PlanSprint(context.Background(), ports.PlanSprintCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		SprintID:      sprintID,
		DemandIDs:     []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)
}

func TestProjectService_SaveKanbanConfigValidation(t *testing.T) {
	t.Parallel()
	repo := &fakeRepo{}
	svc := projectapp.NewService(repo, fakeApps{profile: orgdomain.MethodologyAgileKanban})
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	_, err := svc.SaveKanbanConfig(context.Background(), ports.UpdateKanbanConfigCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		Columns:       []domain.KanbanColumn{{StateCode: "invalid", Label: "X"}},
	})
	require.ErrorIs(t, err, domain.ErrInvalidKanbanConfig)
}
