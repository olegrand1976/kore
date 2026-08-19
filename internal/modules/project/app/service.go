package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/internal/modules/project/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo ports.Repository
	apps ports.ApplicationReader
	now  func() time.Time
}

func NewService(repo ports.Repository, apps ports.ApplicationReader) ports.ProjectService {
	return &service{
		repo: repo,
		apps: apps,
		now:  time.Now,
	}
}

func (s *service) assertAgileApp(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (orgdomain.Application, error) {
	app, err := s.apps.GetApplication(ctx, tenant, appID)
	if err != nil {
		return orgdomain.Application{}, err
	}
	if !app.MethodologyProfile.IsAgile() {
		return orgdomain.Application{}, domain.ErrApplicationNotAgile
	}
	return app, nil
}

func (s *service) CreateEpic(ctx context.Context, cmd ports.CreateEpicCommand) (domain.Epic, error) {
	if _, err := s.assertAgileApp(ctx, cmd.TenantID, cmd.ApplicationID); err != nil {
		return domain.Epic{}, err
	}
	title := strings.TrimSpace(cmd.Title)
	if title == "" {
		return domain.Epic{}, orgdomain.ErrInvalidApplicationLibelle
	}
	epic := domain.NewEpic(cmd.TenantID, cmd.ApplicationID, title, cmd.Description, kernel.NormalizeRequestPriority(cmd.Priority))
	if err := s.repo.SaveEpic(ctx, epic); err != nil {
		return domain.Epic{}, err
	}
	return epic, nil
}

func (s *service) ListEpics(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Epic, error) {
	if _, err := s.assertAgileApp(ctx, tenant, appID); err != nil {
		return nil, err
	}
	return s.repo.ListEpics(ctx, tenant, appID)
}

func (s *service) UpdateEpic(ctx context.Context, cmd ports.UpdateEpicCommand) (domain.Epic, error) {
	if _, err := s.assertAgileApp(ctx, cmd.TenantID, cmd.ApplicationID); err != nil {
		return domain.Epic{}, err
	}
	epic, err := s.repo.GetEpic(ctx, cmd.TenantID, cmd.ApplicationID, cmd.EpicID)
	if err != nil {
		return domain.Epic{}, err
	}
	if cmd.Title != nil {
		title := strings.TrimSpace(*cmd.Title)
		if title == "" {
			return domain.Epic{}, orgdomain.ErrInvalidApplicationLibelle
		}
		epic.Title = title
	}
	if cmd.Description != nil {
		epic.Description = *cmd.Description
	}
	if cmd.Status != nil {
		epic.Status = *cmd.Status
	}
	if cmd.Priority != nil {
		epic.Priority = kernel.NormalizeRequestPriority(*cmd.Priority)
	}
	if cmd.TargetSprint != nil {
		epic.TargetSprintID = *cmd.TargetSprint
	}
	epic.UpdatedAt = s.now().UTC()
	if err := s.repo.SaveEpic(ctx, epic); err != nil {
		return domain.Epic{}, err
	}
	return epic, nil
}

func (s *service) CreateSprint(ctx context.Context, cmd ports.CreateSprintCommand) (domain.Sprint, error) {
	app, err := s.assertAgileApp(ctx, cmd.TenantID, cmd.ApplicationID)
	if err != nil {
		return domain.Sprint{}, err
	}
	if app.MethodologyProfile != orgdomain.MethodologyAgileScrum {
		return domain.Sprint{}, domain.ErrApplicationNotAgile
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return domain.Sprint{}, orgdomain.ErrInvalidApplicationLibelle
	}
	if !cmd.EndDate.After(cmd.StartDate) && !cmd.EndDate.Equal(cmd.StartDate) {
		return domain.Sprint{}, orgdomain.ErrInvalidApplicationLibelle
	}
	sprint := domain.NewSprint(cmd.TenantID, cmd.ApplicationID, name, cmd.Goal, cmd.StartDate, cmd.EndDate, cmd.CapacityPoints)
	if err := s.repo.SaveSprint(ctx, sprint); err != nil {
		return domain.Sprint{}, err
	}
	return sprint, nil
}

func (s *service) ListSprints(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Sprint, error) {
	app, err := s.assertAgileApp(ctx, tenant, appID)
	if err != nil {
		return nil, err
	}
	if app.MethodologyProfile != orgdomain.MethodologyAgileScrum {
		return nil, domain.ErrApplicationNotAgile
	}
	return s.repo.ListSprints(ctx, tenant, appID)
}

func (s *service) StartSprint(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (domain.Sprint, error) {
	app, err := s.assertAgileApp(ctx, tenant, appID)
	if err != nil {
		return domain.Sprint{}, err
	}
	if app.MethodologyProfile != orgdomain.MethodologyAgileScrum {
		return domain.Sprint{}, domain.ErrApplicationNotAgile
	}
	active, err := s.repo.GetActiveSprint(ctx, tenant, appID)
	if err != nil {
		return domain.Sprint{}, err
	}
	if active != nil && active.ID != sprintID {
		return domain.Sprint{}, domain.ErrActiveSprintExists
	}
	sprint, err := s.repo.GetSprint(ctx, tenant, appID, sprintID)
	if err != nil {
		return domain.Sprint{}, err
	}
	if err := sprint.Start(); err != nil {
		return domain.Sprint{}, err
	}
	if err := s.repo.SaveSprint(ctx, sprint); err != nil {
		return domain.Sprint{}, err
	}
	return sprint, nil
}

func (s *service) CloseSprint(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (domain.Sprint, error) {
	app, err := s.assertAgileApp(ctx, tenant, appID)
	if err != nil {
		return domain.Sprint{}, err
	}
	if app.MethodologyProfile != orgdomain.MethodologyAgileScrum {
		return domain.Sprint{}, domain.ErrApplicationNotAgile
	}
	sprint, err := s.repo.GetSprint(ctx, tenant, appID, sprintID)
	if err != nil {
		return domain.Sprint{}, err
	}
	if err := sprint.Close(s.now().UTC()); err != nil {
		return domain.Sprint{}, err
	}
	if err := s.repo.SaveSprint(ctx, sprint); err != nil {
		return domain.Sprint{}, err
	}
	return sprint, nil
}

func (s *service) PlanSprint(ctx context.Context, cmd ports.PlanSprintCommand) error {
	app, err := s.assertAgileApp(ctx, cmd.TenantID, cmd.ApplicationID)
	if err != nil {
		return err
	}
	if app.MethodologyProfile != orgdomain.MethodologyAgileScrum {
		return domain.ErrApplicationNotAgile
	}
	sprint, err := s.repo.GetSprint(ctx, cmd.TenantID, cmd.ApplicationID, cmd.SprintID)
	if err != nil {
		return err
	}
	if sprint.Status == domain.SprintStatusClosed {
		return domain.ErrSprintNotPlanned
	}
	return s.repo.AssignDemandsToSprint(ctx, cmd.TenantID, cmd.ApplicationID, cmd.SprintID, cmd.DemandIDs)
}

func (s *service) ListBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, backlogOnly bool) ([]domain.BacklogItem, error) {
	if _, err := s.assertAgileApp(ctx, tenant, appID); err != nil {
		return nil, err
	}
	return s.repo.ListBacklog(ctx, tenant, appID, backlogOnly)
}

func (s *service) ReorderBacklog(ctx context.Context, cmd ports.ReorderBacklogCommand) error {
	if _, err := s.assertAgileApp(ctx, cmd.TenantID, cmd.ApplicationID); err != nil {
		return err
	}
	return s.repo.ReorderBacklog(ctx, cmd.TenantID, cmd.ApplicationID, cmd.DemandIDs)
}

func (s *service) GetKanbanConfig(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.KanbanConfig, error) {
	app, err := s.assertAgileApp(ctx, tenant, appID)
	if err != nil {
		return domain.KanbanConfig{}, err
	}
	if app.MethodologyProfile != orgdomain.MethodologyAgileKanban {
		return domain.KanbanConfig{}, domain.ErrApplicationNotAgile
	}
	return s.repo.GetKanbanConfig(ctx, tenant, appID)
}

func (s *service) SaveKanbanConfig(ctx context.Context, cmd ports.UpdateKanbanConfigCommand) (domain.KanbanConfig, error) {
	app, err := s.assertAgileApp(ctx, cmd.TenantID, cmd.ApplicationID)
	if err != nil {
		return domain.KanbanConfig{}, err
	}
	if app.MethodologyProfile != orgdomain.MethodologyAgileKanban {
		return domain.KanbanConfig{}, domain.ErrApplicationNotAgile
	}
	cfg := domain.KanbanConfig{
		ApplicationID: cmd.ApplicationID,
		TenantID:      cmd.TenantID,
		Columns:       cmd.Columns,
		UpdatedAt:     s.now().UTC(),
	}
	if err := domain.ValidateKanbanConfig(cfg.Columns); err != nil {
		return domain.KanbanConfig{}, err
	}
	if err := s.repo.SaveKanbanConfig(ctx, cfg); err != nil {
		return domain.KanbanConfig{}, err
	}
	return cfg, nil
}

func (s *service) GetSprintBurndown(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID) (domain.BurndownSeries, error) {
	if _, err := s.assertAgileApp(ctx, tenant, appID); err != nil {
		return domain.BurndownSeries{}, err
	}
	sprint, err := s.repo.GetSprint(ctx, tenant, appID, sprintID)
	if err != nil {
		return domain.BurndownSeries{}, err
	}
	return s.repo.GetSprintBurndown(ctx, tenant, sprint)
}

func (s *service) GetVelocity(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, lastNSprints int) (domain.VelocityReport, error) {
	if _, err := s.assertAgileApp(ctx, tenant, appID); err != nil {
		return domain.VelocityReport{}, err
	}
	if lastNSprints <= 0 {
		lastNSprints = 3
	}
	return s.repo.GetVelocity(ctx, tenant, appID, lastNSprints)
}

func (s *service) ListAgileApplications(ctx context.Context, tenant kernel.TenantID) ([]orgdomain.Application, error) {
	return s.repo.ListAgileApplications(ctx, tenant)
}

func (s *service) HasProjectArtifacts(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (bool, error) {
	n, err := s.repo.CountProjectArtifacts(ctx, tenant, appID)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
