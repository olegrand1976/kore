package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/publicsite/domain"
	"github.com/kore/kore/internal/modules/publicsite/ports"
	"github.com/kore/kore/internal/platform/db"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, l domain.Lead) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO publicsite.leads (id, email, company, size, need, utm_source, consent_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, l.ID, l.Email, l.Company, l.Size, l.Need, l.UTMSource, l.ConsentAt, string(l.Status), l.CreatedAt)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM publicsite.leads WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrLeadNotFound
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (domain.Lead, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, company, size, need, utm_source, consent_at, status, created_at
		FROM publicsite.leads WHERE id = $1
	`, id)
	return scanLead(row)
}

func (r *Repository) ListAvailableSlots(ctx context.Context, filter ports.SlotFilter) ([]domain.BookingSlot, error) {
	query := `
		SELECT id, commercial_id, slot_start, slot_end, status, COALESCE(external_event_id, '')
		FROM publicsite.booking_slots
		WHERE status = 'free' AND slot_start >= $1 AND slot_start < $2
	`
	args := []any{filter.From, filter.To}
	if filter.CommercialID != nil {
		query += ` AND commercial_id = $3`
		args = append(args, *filter.CommercialID)
	}
	query += ` ORDER BY slot_start`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BookingSlot
	for rows.Next() {
		slot, err := scanSlot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, slot)
	}
	return out, rows.Err()
}

func (r *Repository) GetSlot(ctx context.Context, id uuid.UUID) (domain.BookingSlot, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, commercial_id, slot_start, slot_end, status, COALESCE(external_event_id, '')
		FROM publicsite.booking_slots WHERE id = $1
	`, id)
	return scanSlot(row)
}

func (r *Repository) ReserveSlot(ctx context.Context, slotID uuid.UUID) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		var status string
		var slotStart time.Time
		err := tx.QueryRow(ctx, `
			SELECT status, slot_start FROM publicsite.booking_slots WHERE id = $1 FOR UPDATE
		`, slotID).Scan(&status, &slotStart)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrSlotNotFound
			}
			return err
		}
		if status != string(domain.SlotStatusFree) {
			return domain.ErrSlotAlreadyBooked
		}
		if slotStart.Before(time.Now().UTC()) {
			return domain.ErrSlotExpired
		}
		_, err = tx.Exec(ctx, `
			UPDATE publicsite.booking_slots SET status = 'reserved' WHERE id = $1
		`, slotID)
		return err
	})
}

func (r *Repository) ReleaseSlot(ctx context.Context, slotID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE publicsite.booking_slots SET status = 'free' WHERE id = $1
	`, slotID)
	return err
}

func (r *Repository) SaveAppointment(ctx context.Context, a domain.Appointment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO publicsite.appointments (id, lead_id, commercial_id, slot_id, channel, status, cancel_token, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, a.ID, a.LeadID, a.CommercialID, a.SlotID, string(a.Channel), string(a.Status), a.CancelToken, a.CreatedAt)
	return err
}

func (r *Repository) GetAppointmentByToken(ctx context.Context, token string) (domain.Appointment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, lead_id, commercial_id, slot_id, channel, status, cancel_token, created_at
		FROM publicsite.appointments WHERE cancel_token = $1
	`, token)
	var a domain.Appointment
	var channel, status string
	err := row.Scan(&a.ID, &a.LeadID, &a.CommercialID, &a.SlotID, &channel, &status, &a.CancelToken, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Appointment{}, domain.ErrAppointmentNotFound
		}
		return domain.Appointment{}, err
	}
	a.Channel = domain.MeetingChannel(channel)
	a.Status = domain.AppointmentStatus(status)
	return a, nil
}

func (r *Repository) UpdateAppointment(ctx context.Context, a domain.Appointment) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE publicsite.appointments SET slot_id = $2, channel = $3, status = $4 WHERE id = $1
	`, a.ID, a.SlotID, string(a.Channel), string(a.Status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAppointmentNotFound
	}
	return nil
}

func scanLead(row pgx.Row) (domain.Lead, error) {
	var l domain.Lead
	var status string
	err := row.Scan(&l.ID, &l.Email, &l.Company, &l.Size, &l.Need, &l.UTMSource, &l.ConsentAt, &status, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Lead{}, domain.ErrLeadNotFound
		}
		return domain.Lead{}, err
	}
	l.Status = domain.LeadStatus(status)
	return l, nil
}

func scanSlot(row pgx.Row) (domain.BookingSlot, error) {
	var s domain.BookingSlot
	var status string
	err := row.Scan(&s.ID, &s.CommercialID, &s.SlotStart, &s.SlotEnd, &status, &s.ExternalEventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.BookingSlot{}, domain.ErrSlotNotFound
		}
		return domain.BookingSlot{}, err
	}
	s.Status = domain.SlotStatus(status)
	return s, nil
}

type scannable interface {
	Scan(dest ...any) error
}

var _ ports.LeadRepository = (*Repository)(nil)
var _ ports.BookingRepository = (*Repository)(nil)

// SeedSlot inserts a booking slot for dev/demo.
func (r *Repository) SeedSlot(ctx context.Context, commercialID uuid.UUID, start, end time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO publicsite.booking_slots (id, commercial_id, slot_start, slot_end, status)
		VALUES ($1, $2, $3, $4, 'free')
		ON CONFLICT (commercial_id, slot_start) DO NOTHING
	`, uuid.New(), commercialID, start, end)
	if err != nil {
		return fmt.Errorf("seed slot: %w", err)
	}
	return nil
}
