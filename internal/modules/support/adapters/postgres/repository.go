package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/internal/modules/support/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveTicket(ctx context.Context, t domain.Ticket) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO support.tickets (
			id, tenant_id, application_id, subject, description, priority, due_at, state, channel,
			reporter_id, assignee_id, analysis_note, created_at, resolved_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			state = EXCLUDED.state,
			assignee_id = EXCLUDED.assignee_id,
			analysis_note = EXCLUDED.analysis_note,
			resolved_at = EXCLUDED.resolved_at,
			priority = EXCLUDED.priority,
			due_at = EXCLUDED.due_at
	`, t.ID, t.TenantID.UUID(), t.ApplicationID, t.Subject, t.Description, string(t.Priority), t.DueAt, string(t.State),
		t.Channel, t.ReporterID, t.AssigneeID, t.AnalysisNote, t.CreatedAt, t.ResolvedAt)
	return err
}

func (r *Repository) GetTicket(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Ticket, error) {
	return r.scanTicket(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, application_id, subject, description, priority, due_at, state, channel,
			reporter_id, assignee_id, analysis_note, created_at, resolved_at
		FROM support.tickets WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) ListTickets(ctx context.Context, tenant kernel.TenantID) ([]domain.Ticket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, application_id, subject, description, priority, due_at, state, channel,
			reporter_id, assignee_id, analysis_note, created_at, resolved_at
		FROM support.tickets WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Ticket
	for rows.Next() {
		t, err := r.scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repository) SaveReply(ctx context.Context, reply domain.TicketReply) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO support.ticket_replies (id, tenant_id, ticket_id, author_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, reply.ID, reply.TenantID.UUID(), reply.TicketID, reply.AuthorID, reply.Content, reply.CreatedAt)
	return err
}

func (r *Repository) ListReplies(ctx context.Context, tenant kernel.TenantID, ticketID uuid.UUID) ([]domain.TicketReply, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, ticket_id, author_id, content, created_at
		FROM support.ticket_replies WHERE tenant_id = $1 AND ticket_id = $2 ORDER BY created_at
	`, tenant.UUID(), ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TicketReply
	for rows.Next() {
		var reply domain.TicketReply
		var tenantID uuid.UUID
		if err := rows.Scan(&reply.ID, &tenantID, &reply.TicketID, &reply.AuthorID, &reply.Content, &reply.CreatedAt); err != nil {
			return nil, err
		}
		reply.TenantID = kernel.NewTenantID(tenantID)
		out = append(out, reply)
	}
	return out, rows.Err()
}

func (r *Repository) scanTicket(row pgx.Row) (domain.Ticket, error) {
	var t domain.Ticket
	var tenantID uuid.UUID
	var state, priority string
	err := row.Scan(&t.ID, &tenantID, &t.ApplicationID, &t.Subject, &t.Description, &priority, &t.DueAt, &state,
		&t.Channel, &t.ReporterID, &t.AssigneeID, &t.AnalysisNote, &t.CreatedAt, &t.ResolvedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Ticket{}, domain.ErrTicketNotFound
		}
		return domain.Ticket{}, err
	}
	t.TenantID = kernel.NewTenantID(tenantID)
	t.State = domain.TicketState(state)
	t.Priority = kernel.RequestPriority(priority)
	return t, nil
}

var _ ports.SupportRepository = (*Repository)(nil)
