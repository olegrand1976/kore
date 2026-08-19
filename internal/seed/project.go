package seed

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
	projectdomain "github.com/kore/kore/internal/modules/project/domain"
	projectports "github.com/kore/kore/internal/modules/project/ports"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

func (r *Runner) ensureAgileMethodologyProfiles(ctx context.Context, tenant kernel.TenantID) error {
	if r.deps.Pool == nil {
		return nil
	}
	_, err := r.deps.Pool.Exec(ctx, `
		UPDATE org.applications SET methodology_profile = 'agile_scrum'
		WHERE id = $1 AND tenant_id = $2 AND methodology_profile = 'psa'
		AND NOT EXISTS (
			SELECT 1 FROM project.epics WHERE application_id = $1 AND tenant_id = $2
		)
	`, DemoAppID, tenant.UUID())
	if err != nil {
		return err
	}
	_, err = r.deps.Pool.Exec(ctx, `
		UPDATE org.applications SET methodology_profile = 'agile_kanban'
		WHERE id = $1 AND tenant_id = $2 AND methodology_profile = 'psa'
		AND NOT EXISTS (
			SELECT 1 FROM project.kanban_configs WHERE application_id = $1 AND tenant_id = $2
		)
	`, DemoApp3ID, tenant.UUID())
	return err
}

func (r *Runner) seedProjectData(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	if r.deps.Project == nil {
		return nil
	}
	if err := r.ensureAgileMethodologyProfiles(ctx, tenant); err != nil {
		return err
	}
	if err := r.seedScrumProjectData(ctx, tenant); err != nil {
		return err
	}
	return r.seedKanbanProjectData(ctx, tenant)
}

func (r *Runner) seedScrumProjectData(ctx context.Context, tenant kernel.TenantID) error {
	var exists bool
	err := r.deps.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM project.epics WHERE tenant_id = $1 AND application_id = $2)
	`, tenant.UUID(), DemoAppID).Scan(&exists)
	if err != nil || exists {
		return err
	}

	if _, err := r.deps.Project.CreateEpic(ctx, projectports.CreateEpicCommand{
		TenantID:      tenant,
		ApplicationID: DemoAppID,
		Title:         "Portail client — stabilisation",
		Description:   "Epic demo : parcours export PDF et performance CRA.",
		Priority:      "high",
	}); err != nil {
		return err
	}
	if _, err := r.deps.Project.CreateEpic(ctx, projectports.CreateEpicCommand{
		TenantID:      tenant,
		ApplicationID: DemoAppID,
		Title:         "SSO et accès mobile",
		Description:   "Epic demo : authentification mobile et portail client.",
		Priority:      "medium",
	}); err != nil {
		return err
	}

	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 13)
	capacity := int16(40)

	planned, err := r.deps.Project.CreateSprint(ctx, projectports.CreateSprintCommand{
		TenantID:       tenant,
		ApplicationID:  DemoAppID,
		Name:           "Sprint demo +2",
		Goal:           "Backlog suivant",
		StartDate:      start.AddDate(0, 0, 14),
		EndDate:        start.AddDate(0, 0, 27),
		CapacityPoints: &capacity,
	})
	if err != nil {
		return err
	}

	active, err := r.deps.Project.CreateSprint(ctx, projectports.CreateSprintCommand{
		TenantID:       tenant,
		ApplicationID:  DemoAppID,
		Name:           "Sprint demo en cours",
		Goal:           "Correctifs portail ACME",
		StartDate:      start,
		EndDate:        end,
		CapacityPoints: &capacity,
	})
	if err != nil {
		return err
	}

	if err := r.seedDemandStoryPoints(ctx, tenant, DemoAppID); err != nil {
		return err
	}

	demandIDs, err := r.openDemandIDs(ctx, tenant, DemoAppID, 4)
	if err != nil {
		return err
	}
	if len(demandIDs) > 0 {
		if err := r.deps.Project.PlanSprint(ctx, projectports.PlanSprintCommand{
			TenantID:      tenant,
			ApplicationID: DemoAppID,
			SprintID:      active.ID,
			DemandIDs:     demandIDs,
		}); err != nil {
			return err
		}
	}

	if _, err := r.deps.Project.StartSprint(ctx, tenant, DemoAppID, active.ID); err != nil {
		return err
	}

	log.Printf("seed: project scrum — epics, sprints (%s actif, %s planifié)", active.Name, planned.Name)
	return nil
}

func (r *Runner) seedKanbanProjectData(ctx context.Context, tenant kernel.TenantID) error {
	var exists bool
	err := r.deps.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM project.kanban_configs WHERE tenant_id = $1 AND application_id = $2)
	`, tenant.UUID(), DemoApp3ID).Scan(&exists)
	if err != nil || exists {
		return err
	}

	if _, err := r.deps.Project.SaveKanbanConfig(ctx, projectports.UpdateKanbanConfigCommand{
		TenantID:      tenant,
		ApplicationID: DemoApp3ID,
		Columns: []projectdomain.KanbanColumn{
			{StateCode: "ouverte", Label: "À faire"},
			{StateCode: "affectee", Label: "Affectée"},
			{StateCode: "en_cours", Label: "En cours", WipLimit: intPtr(3)},
			{StateCode: "rework", Label: "Rework"},
			{StateCode: "resolue", Label: "Terminé"},
		},
	}); err != nil {
		return err
	}

	log.Printf("seed: project kanban — config %s", DemoApp3Label)
	return nil
}

func intPtr(n int) *int {
	return &n
}

func (r *Runner) seedDemandStoryPoints(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) error {
	if r.deps.Pool == nil {
		return nil
	}
	points := []int16{5, 3, 8, 2, 5, 3}
	rows, err := r.deps.Pool.Query(ctx, `
		SELECT id FROM tma.demands
		WHERE tenant_id = $1 AND application_id = $2 AND status NOT IN ('resolue')
		ORDER BY created_at ASC
		LIMIT 6
	`, tenant.UUID(), appID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i, id := range ids {
		sp := points[i%len(points)]
		_, err := r.deps.Pool.Exec(ctx, `
			UPDATE tma.demands SET story_points = $3, backlog_rank = $4
			WHERE tenant_id = $1 AND id = $2
		`, tenant.UUID(), id, sp, i+1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) openDemandIDs(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, limit int) ([]uuid.UUID, error) {
	if r.deps.Pool != nil {
		rows, err := r.deps.Pool.Query(ctx, `
			SELECT id FROM tma.demands
			WHERE tenant_id = $1 AND application_id = $2
				AND visible = TRUE AND status <> 'resolue' AND sprint_id IS NULL
			ORDER BY backlog_rank NULLS LAST, created_at ASC
			LIMIT $3
		`, tenant.UUID(), appID, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var ids []uuid.UUID
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, rows.Err()
	}

	demands, err := r.deps.TMA.List(ctx, tenant, tmaports.ExportFilter{
		ApplicationID: &appID,
		VisibleOnly:   true,
		BacklogOnly:   true,
	})
	if err != nil {
		return nil, err
	}
	var ids []uuid.UUID
	for _, d := range demands {
		if d.Status == tmadomain.DemandStatusResolved {
			continue
		}
		if d.SprintID != nil {
			continue
		}
		ids = append(ids, d.ID)
		if len(ids) >= limit {
			break
		}
	}
	return ids, nil
}
