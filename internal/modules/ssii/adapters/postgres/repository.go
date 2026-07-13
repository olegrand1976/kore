package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveMission(ctx context.Context, m domain.Mission) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ssii.missions (
			id, tenant_id, client_id, status, start_date, end_date,
			tjm_amount, currency, technologies, client_contact, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			end_date = EXCLUDED.end_date,
			tjm_amount = EXCLUDED.tjm_amount,
			technologies = EXCLUDED.technologies,
			client_contact = EXCLUDED.client_contact
	`, m.ID, m.TenantID.UUID(), m.ClientID, string(m.Status), m.StartDate, m.EndDate,
		m.TJMAmount, m.Currency, m.Technologies, m.ClientContact, m.CreatedAt)
	return err
}

func (r *Repository) GetMission(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error) {
	return r.scanMission(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, client_id, status, start_date, end_date,
			tjm_amount, currency, technologies, client_contact, created_at
		FROM ssii.missions WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) ListMissions(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, client_id, status, start_date, end_date,
			tjm_amount, currency, technologies, client_contact, created_at
		FROM ssii.missions WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Mission
	for rows.Next() {
		m, err := r.scanMission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) scanMission(row pgx.Row) (domain.Mission, error) {
	var m domain.Mission
	var tenantID uuid.UUID
	var status string
	err := row.Scan(&m.ID, &tenantID, &m.ClientID, &status, &m.StartDate, &m.EndDate,
		&m.TJMAmount, &m.Currency, &m.Technologies, &m.ClientContact, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Mission{}, domain.ErrMissionNotFound
		}
		return domain.Mission{}, err
	}
	m.TenantID = kernel.NewTenantID(tenantID)
	m.Status = domain.MissionStatus(status)
	return m, nil
}

func (r *Repository) ListMissionSummaries(ctx context.Context, tenant kernel.TenantID) ([]ports.MissionSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.id, m.client_id, COALESCE(c.raison_sociale, ''), m.status,
			m.start_date, m.end_date, m.tjm_amount, m.currency
		FROM ssii.missions m
		LEFT JOIN org.clients c ON c.id = m.client_id AND c.tenant_id = m.tenant_id
		WHERE m.tenant_id = $1
		ORDER BY m.created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.MissionSummary
	for rows.Next() {
		var s ports.MissionSummary
		var status string
		if err := rows.Scan(&s.ID, &s.ClientID, &s.ClientName, &status, &s.StartDate, &s.EndDate, &s.TJMAmount, &s.Currency); err != nil {
			return nil, err
		}
		s.Status = status
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) ListMissionCollaborators(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID) ([]ports.MissionCollaborator, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.login, u.prenom, u.nom
		FROM ssii.mission_collaborators mc
		JOIN org.users u ON u.id = mc.user_id AND u.deleted_at IS NULL
		WHERE mc.tenant_id = $1 AND mc.mission_id = $2
		ORDER BY u.nom, u.prenom
	`, tenant.UUID(), missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.MissionCollaborator
	for rows.Next() {
		var c ports.MissionCollaborator
		if err := rows.Scan(&c.UserID, &c.Login, &c.Prenom, &c.Nom); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) GetClientName(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx, `
		SELECT raison_sociale FROM org.clients
		WHERE tenant_id = $1 AND id = $2 AND archived = FALSE
	`, tenant.UUID(), clientID).Scan(&name)
	return name, err
}

var _ ports.SSIIRepository = (*Repository)(nil)
