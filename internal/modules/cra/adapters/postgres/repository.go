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
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, ts domain.Timesheet) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		commercial, err := json.Marshal(ts.CommercialInfo)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
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
		if err != nil {
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
				lineID := line.ID
				if lineID == uuid.Nil {
					lineID = uuid.New()
				}
				_, err = tx.Exec(ctx, `
					INSERT INTO cra.time_lines (
						id, tenant_id, week_entry_id, source_type, source_id, day, duration, comment, origin
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				`, lineID, ts.TenantID.UUID(), weekID, line.Source.Type, line.Source.ID,
					line.Day, line.Duration.Minutes, line.Comment, string(line.Origin))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
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

		lineRows, err := r.pool.Query(ctx, `
			SELECT id, source_type, source_id, day, duration, comment, origin
			FROM cra.time_lines WHERE week_entry_id = $1 ORDER BY day
		`, week.ID)
		if err != nil {
			return domain.Timesheet{}, err
		}
		for lineRows.Next() {
			var line domain.TimeLine
			var origin string
			var minutes int
			if err := lineRows.Scan(&line.ID, &line.Source.Type, &line.Source.ID, &line.Day, &minutes, &line.Comment, &origin); err != nil {
				lineRows.Close()
				return domain.Timesheet{}, err
			}
			line.TenantID = ts.TenantID
			line.WeekEntryID = week.ID
			line.Duration = kernel.Duration{Minutes: minutes}
			line.Origin = domain.LineOrigin(origin)
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

var _ ports.CRARepository = (*Repository)(nil)
