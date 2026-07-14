package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool   *db.Pool
	schema craSchema
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{
		pool:   pool,
		schema: probeCraSchema(context.Background(), pool),
	}
}

func (r *Repository) Save(ctx context.Context, ts domain.Timesheet) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		commercial, err := json.Marshal(ts.CommercialInfo)
		if err != nil {
			return err
		}
		if err := r.execSaveTimesheet(ctx, tx, ts, commercial); err != nil {
			return err
		}

		var timesheetID uuid.UUID
		if err := tx.QueryRow(ctx, `
			SELECT id FROM cra.timesheets WHERE tenant_id = $1 AND user_id = $2 AND month = $3
		`, ts.TenantID.UUID(), ts.UserID, string(ts.Month)).Scan(&timesheetID); err != nil {
			return err
		}
		ts.ID = timesheetID

		for _, week := range ts.Weeks {
			weekID := week.ID
			if weekID == uuid.Nil {
				weekID = uuid.New()
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO cra.week_entries (id, tenant_id, timesheet_id, week_number, submitted_at)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (timesheet_id, week_number) DO UPDATE SET submitted_at = EXCLUDED.submitted_at
			`, weekID, ts.TenantID.UUID(), timesheetID, int(week.WeekNumber), week.SubmittedAt)
			if err != nil {
				return err
			}
			if err := tx.QueryRow(ctx, `
				SELECT id FROM cra.week_entries WHERE timesheet_id = $1 AND week_number = $2
			`, timesheetID, int(week.WeekNumber)).Scan(&weekID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM cra.time_lines WHERE week_entry_id = $1`, weekID); err != nil {
				return err
			}
			for _, line := range week.Lines {
				if err := r.execSaveTimeLine(ctx, tx, ts.TenantID, weekID, line); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *Repository) execSaveTimesheet(ctx context.Context, tx pgx.Tx, ts domain.Timesheet, commercial []byte) error {
	if r.schema.hasRejectReason {
		_, err := tx.Exec(ctx, `
			INSERT INTO cra.timesheets (
				id, tenant_id, user_id, month, status, commercial_info, validated_at, validated_by,
				rejected_at, rejected_by, reject_reason, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
			ON CONFLICT (tenant_id, user_id, month) DO UPDATE SET
				status = EXCLUDED.status,
				commercial_info = EXCLUDED.commercial_info,
				validated_at = EXCLUDED.validated_at,
				validated_by = EXCLUDED.validated_by,
				rejected_at = EXCLUDED.rejected_at,
				rejected_by = EXCLUDED.rejected_by,
				reject_reason = EXCLUDED.reject_reason,
				updated_at = NOW()
		`, ts.ID, ts.TenantID.UUID(), ts.UserID, string(ts.Month), string(ts.Status),
			commercial, ts.ValidatedAt, ts.ValidatedBy, ts.RejectedAt, ts.RejectedBy, ts.RejectReason)
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO cra.timesheets (
			id, tenant_id, user_id, month, status, commercial_info, validated_at, validated_by, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (tenant_id, user_id, month) DO UPDATE SET
			status = EXCLUDED.status,
			commercial_info = EXCLUDED.commercial_info,
			validated_at = EXCLUDED.validated_at,
			validated_by = EXCLUDED.validated_by,
			updated_at = NOW()
	`, ts.ID, ts.TenantID.UUID(), ts.UserID, string(ts.Month), string(ts.Status),
		commercial, ts.ValidatedAt, ts.ValidatedBy)
	return err
}

func (r *Repository) execSaveTimeLine(ctx context.Context, tx pgx.Tx, tenant kernel.TenantID, weekID uuid.UUID, line domain.TimeLine) error {
	lineID := line.ID
	if lineID == uuid.Nil {
		lineID = uuid.New()
	}
	if r.schema.hasLineBillable {
		_, err := tx.Exec(ctx, `
			INSERT INTO cra.time_lines (
				id, tenant_id, week_entry_id, source_type, source_id, day, duration, comment, origin, billable
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, lineID, tenant.UUID(), weekID, line.Source.Type, line.Source.ID,
			line.Day, line.Duration.Minutes, line.Comment, string(line.Origin), line.Billable)
		return err
	}
	origin := string(line.Origin)
	if origin == "" {
		origin = string(domain.OriginManual)
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO cra.time_lines (
			id, tenant_id, week_entry_id, source_type, source_id, day, duration, comment, origin
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, lineID, tenant.UUID(), weekID, line.Source.Type, line.Source.ID,
		line.Day, line.Duration.Minutes, line.Comment, origin)
	return err
}

func (r *Repository) Get(ctx context.Context, tenant kernel.TenantID, userID ports.UserID, month domain.Month) (domain.Timesheet, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM cra.timesheets WHERE tenant_id = $1 AND user_id = $2 AND month = $3
	`, tenant.UUID(), userID, string(month)).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Timesheet{}, domain.ErrTimesheetNotFound
		}
		return domain.Timesheet{}, err
	}
	return r.GetByID(ctx, tenant, id)
}

func (r *Repository) GetByID(ctx context.Context, tenant kernel.TenantID, id ports.TimesheetID) (domain.Timesheet, error) {
	var ts domain.Timesheet
	var tenantID uuid.UUID
	var commercial []byte
	var month string
	var status string
	if r.schema.hasRejectReason {
		err := r.pool.QueryRow(ctx, `
			SELECT id, tenant_id, user_id, month, status, commercial_info, validated_at, validated_by,
				rejected_at, rejected_by, COALESCE(reject_reason, '')
			FROM cra.timesheets WHERE tenant_id = $1 AND id = $2
		`, tenant.UUID(), id).Scan(&ts.ID, &tenantID, &ts.UserID, &month, &status, &commercial, &ts.ValidatedAt, &ts.ValidatedBy,
			&ts.RejectedAt, &ts.RejectedBy, &ts.RejectReason)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.Timesheet{}, domain.ErrTimesheetNotFound
			}
			return domain.Timesheet{}, err
		}
	} else {
		err := r.pool.QueryRow(ctx, `
			SELECT id, tenant_id, user_id, month, status, commercial_info, validated_at, validated_by
			FROM cra.timesheets WHERE tenant_id = $1 AND id = $2
		`, tenant.UUID(), id).Scan(&ts.ID, &tenantID, &ts.UserID, &month, &status, &commercial, &ts.ValidatedAt, &ts.ValidatedBy)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.Timesheet{}, domain.ErrTimesheetNotFound
			}
			return domain.Timesheet{}, err
		}
	}
	ts.TenantID = kernel.NewTenantID(tenantID)
	ts.Month = domain.Month(month)
	ts.Status = domain.TimesheetStatus(status)
	if len(commercial) > 0 {
		_ = json.Unmarshal(commercial, &ts.CommercialInfo)
	}

	weekRows, err := r.pool.Query(ctx, `
		SELECT id, week_number, submitted_at FROM cra.week_entries WHERE timesheet_id = $1 ORDER BY week_number
	`, ts.ID)
	if err != nil {
		return domain.Timesheet{}, err
	}
	defer weekRows.Close()
	for weekRows.Next() {
		var week domain.WeekEntry
		var weekNum int
		if err := weekRows.Scan(&week.ID, &weekNum, &week.SubmittedAt); err != nil {
			return domain.Timesheet{}, err
		}
		week.TenantID = ts.TenantID
		week.TimesheetID = ts.ID
		week.WeekNumber = domain.WeekNumber(weekNum)

		lineQuery := `
			SELECT id, source_type, source_id, day, duration, comment, origin, billable
			FROM cra.time_lines WHERE week_entry_id = $1 ORDER BY day`
		if !r.schema.hasLineBillable {
			lineQuery = `
			SELECT id, source_type, source_id, day, duration, comment, origin
			FROM cra.time_lines WHERE week_entry_id = $1 ORDER BY day`
		}
		lineRows, err := r.pool.Query(ctx, lineQuery, week.ID)
		if err != nil {
			return domain.Timesheet{}, err
		}
		for lineRows.Next() {
			var line domain.TimeLine
			var origin string
			var minutes int
			var billable bool
			if r.schema.hasLineBillable {
				if err := lineRows.Scan(&line.ID, &line.Source.Type, &line.Source.ID, &line.Day, &minutes, &line.Comment, &origin, &billable); err != nil {
					lineRows.Close()
					return domain.Timesheet{}, err
				}
			} else {
				if err := lineRows.Scan(&line.ID, &line.Source.Type, &line.Source.ID, &line.Day, &minutes, &line.Comment, &origin); err != nil {
					lineRows.Close()
					return domain.Timesheet{}, err
				}
				billable = true
			}
			line.TenantID = ts.TenantID
			line.WeekEntryID = week.ID
			line.Duration = kernel.Duration{Minutes: minutes}
			line.Origin = domain.LineOrigin(origin)
			line.Billable = billable
			week.Lines = append(week.Lines, line)
		}
		lineRows.Close()
		if err := lineRows.Err(); err != nil {
			return domain.Timesheet{}, err
		}
		ts.Weeks = append(ts.Weeks, week)
	}
	return ts, weekRows.Err()
}

func (r *Repository) ListByUser(ctx context.Context, tenant kernel.TenantID, userID ports.UserID, limit int) ([]domain.Timesheet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id FROM cra.timesheets
		WHERE tenant_id = $1 AND user_id = $2
		ORDER BY month DESC
		LIMIT $3
	`, tenant.UUID(), userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanTimesheetIDs(ctx, tenant, rows)
}

func (r *Repository) ListByTenant(ctx context.Context, tenant kernel.TenantID, limit int) ([]domain.Timesheet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id FROM cra.timesheets
		WHERE tenant_id = $1
		ORDER BY month DESC
		LIMIT $2
	`, tenant.UUID(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanTimesheetIDs(ctx, tenant, rows)
}

func (r *Repository) ListSummariesByUser(ctx context.Context, tenant kernel.TenantID, userID ports.UserID, limit int) ([]domain.TimesheetSummary, error) {
	return r.queryTimesheetSummaries(ctx, `
		WHERE t.tenant_id = $1 AND t.user_id = $2
		GROUP BY t.id, u.login, u.prenom, u.nom
		ORDER BY t.month DESC, u.login ASC
		LIMIT $3
	`, tenant.UUID(), userID, limit)
}

func (r *Repository) ListSummariesByTenant(ctx context.Context, tenant kernel.TenantID, limit int) ([]domain.TimesheetSummary, error) {
	return r.queryTimesheetSummaries(ctx, `
		WHERE t.tenant_id = $1
		GROUP BY t.id, u.login, u.prenom, u.nom
		ORDER BY t.month DESC, u.login ASC
		LIMIT $2
	`, tenant.UUID(), limit)
}

func (r *Repository) ListSummariesByTenantMonth(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]domain.TimesheetSummary, error) {
	return r.queryTimesheetSummaries(ctx, `
		WHERE t.tenant_id = $1 AND t.month = $2 AND u.cra_requis = TRUE
		GROUP BY t.id, u.login, u.prenom, u.nom
		ORDER BY u.nom ASC, u.prenom ASC
	`, tenant.UUID(), string(month))
}

func (r *Repository) queryTimesheetSummaries(ctx context.Context, suffix string, args ...any) ([]domain.TimesheetSummary, error) {
	rows, err := r.pool.Query(ctx, r.timesheetSummarySelect()+suffix, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTimesheetSummaries(rows)
}

func (r *Repository) timesheetSummarySelect() string {
	rejectReasonCol := `''`
	if r.schema.hasRejectReason {
		rejectReasonCol = `COALESCE(t.reject_reason, '')`
	}
	prefillCol := `0 AS prefill_minutes`
	if r.schema.hasLineOrigin {
		prefillCol = `COALESCE(SUM(tl.duration) FILTER (WHERE tl.duration > 0 AND tl.origin = 'prefill'), 0) AS prefill_minutes`
	}
	missionCol := commercialInfoMissionUUID + ` AS mission_id`
	if r.schema.hasSSMissions {
		missionCol = timesheetSummaryMissionLookup
	}
	return timesheetSummarySelectBase(rejectReasonCol, prefillCol) + missionCol + timesheetSummaryFrom
}

const timesheetSummaryFrom = `
		FROM cra.timesheets t
		JOIN org.users u ON u.id = t.user_id AND u.tenant_id = t.tenant_id
		LEFT JOIN cra.week_entries we ON we.timesheet_id = t.id
		LEFT JOIN cra.time_lines tl ON tl.week_entry_id = we.id
`

const commercialInfoClientUUID = `
			CASE
				WHEN (t.commercial_info->>'clientId') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
				THEN (t.commercial_info->>'clientId')::uuid
			END`

const commercialInfoMissionUUID = `
			CASE
				WHEN (t.commercial_info->>'missionId') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
				THEN (t.commercial_info->>'missionId')::uuid
			END`

const timesheetSummaryMissionLookup = `
			COALESCE(
				` + commercialInfoMissionUUID + `,
				(
					SELECT m.id FROM ssii.missions m
					INNER JOIN ssii.mission_collaborators mc ON mc.mission_id = m.id AND mc.user_id = t.user_id
					INNER JOIN org.clients c ON c.id = m.client_id
					  AND c.raison_sociale = (t.commercial_info->>'client')
					WHERE m.tenant_id = t.tenant_id
					ORDER BY m.created_at DESC
					LIMIT 1
				)
			) AS mission_id`

func timesheetSummarySelectBase(rejectReasonCol, prefillCol string) string {
	return `
		SELECT
			t.id,
			t.user_id,
			u.login,
			u.prenom,
			u.nom,
			t.month,
			t.status,
			t.commercial_info,
			` + rejectReasonCol + `,
			t.updated_at,
			` + prefillCol + `,
			COALESCE(SUM(tl.duration), 0) AS total_minutes,
			COUNT(we.id) FILTER (WHERE we.submitted_at IS NOT NULL) AS weeks_submitted,
			COUNT(DISTINCT we.id) AS weeks_total,
			COALESCE(
				` + commercialInfoClientUUID + `,
				(
					SELECT c.id FROM org.clients c
					WHERE c.tenant_id = t.tenant_id
					  AND c.raison_sociale = (t.commercial_info->>'client')
					  AND NOT c.archived
					LIMIT 1
				)
			) AS client_id,
`
}

func scanTimesheetSummaries(rows pgx.Rows) ([]domain.TimesheetSummary, error) {
	var out []domain.TimesheetSummary
	for rows.Next() {
		var summary domain.TimesheetSummary
		var month string
		var status string
		var commercial []byte
		var clientID *uuid.UUID
		var missionID *uuid.UUID
		var prefillMinutes int
		if err := rows.Scan(
			&summary.ID,
			&summary.UserID,
			&summary.UserLogin,
			&summary.UserPrenom,
			&summary.UserNom,
			&month,
			&status,
			&commercial,
			&summary.RejectReason,
			&summary.UpdatedAt,
			&prefillMinutes,
			&summary.TotalMinutes,
			&summary.WeeksSubmitted,
			&summary.WeeksTotal,
			&clientID,
			&missionID,
		); err != nil {
			return nil, err
		}
		summary.Month = domain.Month(month)
		summary.Status = domain.TimesheetStatus(status)
		if len(commercial) > 0 {
			_ = json.Unmarshal(commercial, &summary.CommercialInfo)
		}
		summary.ClientID = clientID
		summary.MissionID = missionID
		if summary.ClientID == nil && summary.CommercialInfo.ClientID != nil {
			summary.ClientID = summary.CommercialInfo.ClientID
		}
		if summary.MissionID == nil && summary.CommercialInfo.MissionID != nil {
			summary.MissionID = summary.CommercialInfo.MissionID
		}
		if summary.TotalMinutes > 0 {
			summary.PrefillRatio = int((float64(prefillMinutes) / float64(summary.TotalMinutes)) * 100)
		}
		out = append(out, summary)
	}
	return out, rows.Err()
}

func (r *Repository) scanTimesheetIDs(ctx context.Context, tenant kernel.TenantID, rows pgx.Rows) ([]domain.Timesheet, error) {
	var out []domain.Timesheet
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ts, err := r.GetByID(ctx, tenant, id)
		if err != nil {
			return nil, err
		}
		out = append(out, ts)
	}
	return out, rows.Err()
}

func (r *Repository) FindConsumption(ctx context.Context, tenant kernel.TenantID, appID ports.ApplicationID, period kernel.Period) ([]domain.Consumption, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.user_id, tl.source_type, tl.source_id, tl.day, tl.duration
		FROM cra.time_lines tl
		JOIN cra.week_entries we ON we.id = tl.week_entry_id
		JOIN cra.timesheets t ON t.id = we.timesheet_id
		WHERE tl.tenant_id = $1
		  AND tl.source_type = 'application'
		  AND tl.source_id = $2
		  AND tl.day >= $3 AND tl.day <= $4
		  AND tl.billable = TRUE
		ORDER BY tl.day
	`, tenant.UUID(), appID.String(), period.Start, period.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Consumption
	for rows.Next() {
		var c domain.Consumption
		var minutes int
		if err := rows.Scan(&c.UserID, &c.Source.Type, &c.Source.ID, &c.Day, &minutes); err != nil {
			return nil, err
		}
		c.Duration = kernel.Duration{Minutes: minutes}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) DeleteFutureLines(ctx context.Context, tenant kernel.TenantID, source domain.SourceRef, from time.Time) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM cra.time_lines tl
		USING cra.week_entries we, cra.timesheets t
		WHERE tl.week_entry_id = we.id
		  AND we.timesheet_id = t.id
		  AND t.tenant_id = $1
		  AND tl.source_type = $2
		  AND tl.source_id = $3
		  AND tl.day >= $4
		  AND tl.origin = 'prefill'
	`, tenant.UUID(), source.Type, source.ID, from)
	return err
}

func (r *Repository) ListDailyActivityInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]ports.DailyActivityRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.user_id, COALESCE(u.prenom, ''), COALESCE(u.nom, ''), tl.day,
		       SUM(tl.duration)::int,
		       COALESCE(CASE WHEN tl.source_type = 'mission' THEN tl.source_id ELSE '' END, ''),
		       COALESCE(MAX(c.raison_sociale), ''),
		       COALESCE(MAX(NULLIF(TRIM(ts.commercial_info->>'mission'), '')), MAX(c.raison_sociale), '')
		FROM cra.time_lines tl
		INNER JOIN cra.week_entries we ON we.id = tl.week_entry_id
		INNER JOIN cra.timesheets t ON t.id = we.timesheet_id
		INNER JOIN org.users u ON u.id = t.user_id AND u.tenant_id = t.tenant_id
		LEFT JOIN ssii.missions m ON tl.source_type = 'mission'
			AND m.id::text = tl.source_id AND m.tenant_id = t.tenant_id
		LEFT JOIN org.clients c ON c.id = m.client_id AND c.tenant_id = t.tenant_id
		WHERE t.tenant_id = $1 AND tl.day >= $2 AND tl.day <= $3 AND tl.duration > 0
		GROUP BY t.user_id, u.prenom, u.nom, tl.day,
			COALESCE(CASE WHEN tl.source_type = 'mission' THEN tl.source_id ELSE '' END, '')
		ORDER BY tl.day, u.nom, u.prenom
	`, tenant.UUID(), period.Start, period.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.DailyActivityRow
	for rows.Next() {
		var row ports.DailyActivityRow
		if err := rows.Scan(&row.UserID, &row.UserPrenom, &row.UserNom, &row.Day, &row.Minutes, &row.MissionID, &row.ClientLabel, &row.MissionLabel); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

var _ ports.CRARepository = (*Repository)(nil)
