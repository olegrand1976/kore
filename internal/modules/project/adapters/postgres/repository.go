package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/internal/modules/project/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveEpic(ctx context.Context, epic domain.Epic) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO project.epics (
			id, tenant_id, application_id, title, description, status, priority,
			target_sprint_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			target_sprint_id = EXCLUDED.target_sprint_id,
			updated_at = EXCLUDED.updated_at
	`, epic.ID, epic.TenantID.UUID(), epic.ApplicationID, epic.Title, epic.Description,
		string(epic.Status), string(epic.Priority), epic.TargetSprintID, epic.CreatedAt, epic.UpdatedAt)
	return err
}

func (r *Repository) GetEpic(ctx context.Context, tenant kernel.TenantID, appID, id uuid.UUID) (domain.Epic, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, title, description, status, priority,
			target_sprint_id, created_at, updated_at
		FROM project.epics
		WHERE tenant_id = $1 AND application_id = $2 AND id = $3
	`, tenant.UUID(), appID, id)
	return scanEpic(row)
}

func (r *Repository) ListEpics(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Epic, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, application_id, title, description, status, priority,
			target_sprint_id, created_at, updated_at
		FROM project.epics
		WHERE tenant_id = $1 AND application_id = $2
		ORDER BY created_at DESC
	`, tenant.UUID(), appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Epic
	for rows.Next() {
		epic, err := scanEpic(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, epic)
	}
	return out, rows.Err()
}

func (r *Repository) SaveSprint(ctx context.Context, sprint domain.Sprint) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO project.sprints (
			id, tenant_id, application_id, name, goal, start_date, end_date,
			status, capacity_points, created_at, closed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			goal = EXCLUDED.goal,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			status = EXCLUDED.status,
			capacity_points = EXCLUDED.capacity_points,
			closed_at = EXCLUDED.closed_at
	`, sprint.ID, sprint.TenantID.UUID(), sprint.ApplicationID, sprint.Name, sprint.Goal,
		sprint.StartDate, sprint.EndDate, string(sprint.Status), sprint.CapacityPoints,
		sprint.CreatedAt, sprint.ClosedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "idx_project_sprints_one_active" {
			return domain.ErrActiveSprintExists
		}
	}
	return err
}

func (r *Repository) GetSprint(ctx context.Context, tenant kernel.TenantID, appID, id uuid.UUID) (domain.Sprint, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, name, goal, start_date, end_date,
			status, capacity_points, created_at, closed_at
		FROM project.sprints
		WHERE tenant_id = $1 AND application_id = $2 AND id = $3
	`, tenant.UUID(), appID, id)
	return scanSprint(row)
}

func (r *Repository) ListSprints(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) ([]domain.Sprint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, application_id, name, goal, start_date, end_date,
			status, capacity_points, created_at, closed_at
		FROM project.sprints
		WHERE tenant_id = $1 AND application_id = $2
		ORDER BY start_date DESC
	`, tenant.UUID(), appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Sprint
	for rows.Next() {
		s, err := scanSprint(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) GetActiveSprint(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (*domain.Sprint, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, name, goal, start_date, end_date,
			status, capacity_points, created_at, closed_at
		FROM project.sprints
		WHERE tenant_id = $1 AND application_id = $2 AND status = 'active'
		LIMIT 1
	`, tenant.UUID(), appID)
	s, err := scanSprint(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, domain.ErrSprintNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *Repository) SaveKanbanConfig(ctx context.Context, cfg domain.KanbanConfig) error {
	raw, err := json.Marshal(cfg.Columns)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO project.kanban_configs (application_id, tenant_id, columns, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (application_id) DO UPDATE SET
			columns = EXCLUDED.columns,
			updated_at = EXCLUDED.updated_at
	`, cfg.ApplicationID, cfg.TenantID.UUID(), raw, cfg.UpdatedAt)
	return err
}

func (r *Repository) GetKanbanConfig(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (domain.KanbanConfig, error) {
	var cfg domain.KanbanConfig
	var tenantID uuid.UUID
	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT application_id, tenant_id, columns, updated_at
		FROM project.kanban_configs
		WHERE tenant_id = $1 AND application_id = $2
	`, tenant.UUID(), appID).Scan(&cfg.ApplicationID, &tenantID, &raw, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.KanbanConfig{
				ApplicationID: appID,
				TenantID:      tenant,
				Columns:       nil,
				UpdatedAt:     time.Now().UTC(),
			}, nil
		}
		return domain.KanbanConfig{}, err
	}
	cfg.TenantID = kernel.NewTenantID(tenantID)
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg.Columns); err != nil {
			return domain.KanbanConfig{}, err
		}
	}
	return cfg, nil
}

func (r *Repository) AssignDemandsToSprint(ctx context.Context, tenant kernel.TenantID, appID, sprintID uuid.UUID, demandIDs []uuid.UUID) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE tma.demands SET sprint_id = NULL
			WHERE tenant_id = $1 AND application_id = $2 AND sprint_id = $3
		`, tenant.UUID(), appID, sprintID)
		if err != nil {
			return err
		}
		for _, id := range demandIDs {
			tag, err := tx.Exec(ctx, `
				UPDATE tma.demands SET sprint_id = $4
				WHERE tenant_id = $1 AND application_id = $2 AND id = $3
			`, tenant.UUID(), appID, id, sprintID)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return fmt.Errorf("demand not found: %s", id)
			}
		}
		return nil
	})
}

func (r *Repository) ReorderBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, demandIDs []uuid.UUID) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		for i, id := range demandIDs {
			rank := i + 1
			tag, err := tx.Exec(ctx, `
				UPDATE tma.demands SET backlog_rank = $4
				WHERE tenant_id = $1 AND application_id = $2 AND id = $3
			`, tenant.UUID(), appID, id, rank)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return fmt.Errorf("demand not found: %s", id)
			}
		}
		return nil
	})
}

func (r *Repository) ListBacklog(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, backlogOnly bool) ([]domain.BacklogItem, error) {
	query := `
		SELECT id, subject, status, story_points, epic_id, sprint_id, backlog_rank, assignee_id
		FROM tma.demands
		WHERE tenant_id = $1 AND application_id = $2 AND visible = TRUE`
	if backlogOnly {
		query += ` AND sprint_id IS NULL AND status NOT IN ('resolue')`
	}
	query += ` ORDER BY backlog_rank NULLS LAST, created_at ASC`

	rows, err := r.pool.Query(ctx, query, tenant.UUID(), appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BacklogItem
	for rows.Next() {
		var item domain.BacklogItem
		if err := rows.Scan(
			&item.DemandID, &item.Subject, &item.Status,
			&item.StoryPoints, &item.EpicID, &item.SprintID, &item.BacklogRank, &item.AssigneeID,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) CountProjectArtifacts(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT (
			(SELECT COUNT(*) FROM project.epics WHERE tenant_id = $1 AND application_id = $2) +
			(SELECT COUNT(*) FROM project.sprints WHERE tenant_id = $1 AND application_id = $2) +
			(SELECT COUNT(*) FROM project.kanban_configs WHERE tenant_id = $1 AND application_id = $2) +
			(SELECT COUNT(*) FROM tma.demands WHERE tenant_id = $1 AND application_id = $2
				AND (epic_id IS NOT NULL OR sprint_id IS NOT NULL OR story_points IS NOT NULL OR backlog_rank IS NOT NULL))
		)
	`, tenant.UUID(), appID).Scan(&n)
	return n, err
}

func (r *Repository) GetSprintBurndown(ctx context.Context, tenant kernel.TenantID, sprint domain.Sprint) (domain.BurndownSeries, error) {
	var planned int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(story_points), 0)
		FROM tma.demands
		WHERE tenant_id = $1 AND sprint_id = $2 AND story_points IS NOT NULL
	`, tenant.UUID(), sprint.ID).Scan(&planned)
	if err != nil {
		return domain.BurndownSeries{}, err
	}

	startDay := sprint.StartDate.UTC().Truncate(24 * time.Hour)
	endDay := sprint.EndDate.UTC().Truncate(24 * time.Hour)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	lastDay := endDay
	if today.Before(lastDay) {
		lastDay = today
	}
	if lastDay.Before(startDay) {
		lastDay = startDay
	}

	days := int(lastDay.Sub(startDay).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	totalDays := int(endDay.Sub(startDay).Hours()/24) + 1
	if totalDays < 1 {
		totalDays = 1
	}

	resolvedByDay := make(map[string]int)
	rows, err := r.pool.Query(ctx, `
		SELECT (resolved_at AT TIME ZONE 'UTC')::date AS d, COALESCE(SUM(story_points), 0)::int
		FROM tma.demands
		WHERE tenant_id = $1 AND sprint_id = $2
			AND story_points IS NOT NULL AND resolved_at IS NOT NULL
		GROUP BY d
		ORDER BY d
	`, tenant.UUID(), sprint.ID)
	if err != nil {
		return domain.BurndownSeries{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var d time.Time
		var pts int
		if err := rows.Scan(&d, &pts); err != nil {
			return domain.BurndownSeries{}, err
		}
		resolvedByDay[d.Format("2006-01-02")] = pts
	}
	if err := rows.Err(); err != nil {
		return domain.BurndownSeries{}, err
	}

	doneCumulative := 0
	resolvedDays := make([]string, 0, len(resolvedByDay))
	for d := range resolvedByDay {
		resolvedDays = append(resolvedDays, d)
	}
	sort.Strings(resolvedDays)
	resolvedIdx := 0

	points := make([]domain.BurndownPoint, 0, days)
	for i := 0; i < days; i++ {
		date := startDay.AddDate(0, 0, i)
		dayKey := date.Format("2006-01-02")
		for resolvedIdx < len(resolvedDays) && resolvedDays[resolvedIdx] <= dayKey {
			doneCumulative += resolvedByDay[resolvedDays[resolvedIdx]]
			resolvedIdx++
		}
		remaining := planned - doneCumulative
		if remaining < 0 {
			remaining = 0
		}
		ideal := planned - (planned * i / totalDays)
		if ideal < 0 {
			ideal = 0
		}
		points = append(points, domain.BurndownPoint{
			Date:            date,
			RemainingPoints: remaining,
			IdealPoints:     ideal,
		})
	}
	return domain.BurndownSeries{
		SprintID:      sprint.ID,
		PlannedPoints: planned,
		Points:        points,
	}, nil
}

func (r *Repository) GetVelocity(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, lastN int) (domain.VelocityReport, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.name,
			COALESCE((
				SELECT SUM(d.story_points)
				FROM tma.demands d
				WHERE d.tenant_id = s.tenant_id AND d.sprint_id = s.id
					AND d.resolved_at IS NOT NULL AND d.story_points IS NOT NULL
			), 0)
		FROM project.sprints s
		WHERE s.tenant_id = $1 AND s.application_id = $2 AND s.status = 'closed'
		ORDER BY s.closed_at DESC NULLS LAST
		LIMIT $3
	`, tenant.UUID(), appID, lastN)
	if err != nil {
		return domain.VelocityReport{}, err
	}
	defer rows.Close()

	report := domain.VelocityReport{ApplicationID: appID}
	var total int
	for rows.Next() {
		var item domain.VelocitySprint
		if err := rows.Scan(&item.SprintID, &item.SprintName, &item.ClosedPoints); err != nil {
			return domain.VelocityReport{}, err
		}
		report.Sprints = append(report.Sprints, item)
		total += item.ClosedPoints
	}
	if err := rows.Err(); err != nil {
		return domain.VelocityReport{}, err
	}
	if len(report.Sprints) > 0 {
		report.AverageVelocity = float64(total) / float64(len(report.Sprints))
	}
	return report, nil
}

func (r *Repository) ListAgileApplications(ctx context.Context, tenant kernel.TenantID) ([]orgdomain.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, libelle,
			COALESCE(proprietaire, ''), COALESCE(mode_facturation, 'temps_passe'), COALESCE(uo_activee, FALSE),
			chef_utilisateur_id, budget_defaut_id, active, COALESCE(default_tjm_cents, 0),
			COALESCE(methodology_profile, 'psa')
		FROM org.applications
		WHERE tenant_id = $1 AND methodology_profile IN ('agile_scrum', 'agile_kanban') AND active = TRUE
		ORDER BY libelle
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []orgdomain.Application
	for rows.Next() {
		app, err := scanAgileApplication(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, app)
	}
	return out, rows.Err()
}

func scanEpic(row pgx.Row) (domain.Epic, error) {
	var epic domain.Epic
	var tenantID uuid.UUID
	var status, priority string
	err := row.Scan(
		&epic.ID, &tenantID, &epic.ApplicationID, &epic.Title, &epic.Description,
		&status, &priority, &epic.TargetSprintID, &epic.CreatedAt, &epic.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Epic{}, domain.ErrEpicNotFound
		}
		return domain.Epic{}, err
	}
	epic.TenantID = kernel.NewTenantID(tenantID)
	epic.Status = domain.EpicStatus(status)
	epic.Priority = kernel.RequestPriority(priority)
	return epic, nil
}

func scanSprint(row pgx.Row) (domain.Sprint, error) {
	var sprint domain.Sprint
	var tenantID uuid.UUID
	var status string
	err := row.Scan(
		&sprint.ID, &tenantID, &sprint.ApplicationID, &sprint.Name, &sprint.Goal,
		&sprint.StartDate, &sprint.EndDate, &status, &sprint.CapacityPoints,
		&sprint.CreatedAt, &sprint.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Sprint{}, domain.ErrSprintNotFound
		}
		return domain.Sprint{}, err
	}
	sprint.TenantID = kernel.NewTenantID(tenantID)
	sprint.Status = domain.SprintStatus(status)
	return sprint, nil
}

func scanAgileApplication(rows pgx.Rows) (orgdomain.Application, error) {
	var app orgdomain.Application
	var tenantID uuid.UUID
	var proprietaire, modeFacturation, methodology string
	if err := rows.Scan(
		&app.ID, &tenantID, &app.Libelle,
		&proprietaire, &modeFacturation, &app.UOActivee,
		&app.ChefUtilisateurID, &app.BudgetDefautID, &app.Active, &app.DefaultTJMCents,
		&methodology,
	); err != nil {
		return orgdomain.Application{}, err
	}
	app.TenantID = kernel.NewTenantID(tenantID)
	app.Proprietaire = proprietaire
	app.ModeFacturation = modeFacturation
	profile, _ := orgdomain.NormalizeMethodologyProfile(methodology)
	app.MethodologyProfile = profile
	return app, nil
}

var _ ports.Repository = (*Repository)(nil)
