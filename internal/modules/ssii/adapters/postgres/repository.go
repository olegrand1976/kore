package postgres

import (
	"context"
	"encoding/json"
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
	rateUnit := string(m.RateUnit)
	if rateUnit == "" {
		rateUnit = string(domain.RateUnitTJM)
	}
	contactIDs := m.ClientContactIDs
	if contactIDs == nil {
		contactIDs = []uuid.UUID{}
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ssii.missions (
			id, tenant_id, client_id, status, start_date, end_date,
			title, rate_unit, tjm_amount, currency, technologies, client_contact,
			client_contact_ids, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			end_date = EXCLUDED.end_date,
			title = EXCLUDED.title,
			rate_unit = EXCLUDED.rate_unit,
			tjm_amount = EXCLUDED.tjm_amount,
			technologies = EXCLUDED.technologies,
			client_contact = EXCLUDED.client_contact,
			client_contact_ids = EXCLUDED.client_contact_ids
	`, m.ID, m.TenantID.UUID(), m.ClientID, string(m.Status), m.StartDate, m.EndDate,
		m.Title, rateUnit, m.TJMAmount, m.Currency, m.Technologies, m.ClientContact,
		contactIDs, m.CreatedAt)
	return err
}

func (r *Repository) GetMission(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error) {
	return r.scanMission(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, client_id, status, start_date, end_date,
			COALESCE(title, ''), COALESCE(rate_unit, 'tjm'),
			tjm_amount, currency, technologies, client_contact,
			COALESCE(client_contact_ids, '{}'), created_at
		FROM ssii.missions WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) ListMissions(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, client_id, status, start_date, end_date,
			COALESCE(title, ''), COALESCE(rate_unit, 'tjm'),
			tjm_amount, currency, technologies, client_contact,
			COALESCE(client_contact_ids, '{}'), created_at
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
	var status, rateUnit string
	err := row.Scan(&m.ID, &tenantID, &m.ClientID, &status, &m.StartDate, &m.EndDate,
		&m.Title, &rateUnit, &m.TJMAmount, &m.Currency, &m.Technologies, &m.ClientContact,
		&m.ClientContactIDs, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Mission{}, domain.ErrMissionNotFound
		}
		return domain.Mission{}, err
	}
	m.TenantID = kernel.NewTenantID(tenantID)
	m.Status = domain.MissionStatus(status)
	normalized, nErr := domain.NormalizeRateUnit(rateUnit)
	if nErr != nil {
		normalized = domain.RateUnitTJM
	}
	m.RateUnit = normalized
	if m.ClientContactIDs == nil {
		m.ClientContactIDs = []uuid.UUID{}
	}
	return m, nil
}

func (r *Repository) ListMissionSummaries(ctx context.Context, tenant kernel.TenantID) ([]ports.MissionSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.id, m.client_id, COALESCE(c.raison_sociale, ''), m.status,
			m.start_date, m.end_date, COALESCE(m.title, ''), COALESCE(m.rate_unit, 'tjm'),
			m.tjm_amount, m.currency
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
		var status, rateUnit string
		if err := rows.Scan(&s.ID, &s.ClientID, &s.ClientName, &status, &s.StartDate, &s.EndDate,
			&s.Title, &rateUnit, &s.TJMAmount, &s.Currency); err != nil {
			return nil, err
		}
		s.Status = status
		normalized, nErr := domain.NormalizeRateUnit(rateUnit)
		if nErr != nil {
			normalized = domain.RateUnitTJM
		}
		s.RateUnit = string(normalized)
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

func (r *Repository) SaveMissionCollaborators(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID, userIDs []uuid.UUID) error {
	if _, err := r.pool.Exec(ctx, `
		DELETE FROM ssii.mission_collaborators WHERE tenant_id = $1 AND mission_id = $2
	`, tenant.UUID(), missionID); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if _, err := r.pool.Exec(ctx, `
			INSERT INTO ssii.mission_collaborators (id, tenant_id, mission_id, user_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (mission_id, user_id) DO NOTHING
		`, uuid.New(), tenant.UUID(), missionID, userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetClientName(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx, `
		SELECT raison_sociale FROM org.clients
		WHERE tenant_id = $1 AND id = $2 AND archived = FALSE
	`, tenant.UUID(), clientID).Scan(&name)
	return name, err
}

func (r *Repository) GetClientPays(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID) (string, error) {
	var pays string
	err := r.pool.QueryRow(ctx, `
		SELECT pays FROM org.clients
		WHERE tenant_id = $1 AND id = $2 AND archived = FALSE
	`, tenant.UUID(), clientID).Scan(&pays)
	return pays, err
}

func (r *Repository) ListClientContacts(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID) ([]ports.ClientContactSnapshot, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT contacts FROM org.clients
		WHERE tenant_id = $1 AND id = $2 AND archived = FALSE
	`, tenant.UUID(), clientID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if len(raw) == 0 {
		return []ports.ClientContactSnapshot{}, nil
	}
	var contacts []struct {
		ID        uuid.UUID `json:"id"`
		Nom       string    `json:"nom"`
		Prenom    string    `json:"prenom"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
		Telephone string    `json:"telephone"`
	}
	if err := json.Unmarshal(raw, &contacts); err != nil {
		return nil, err
	}
	out := make([]ports.ClientContactSnapshot, 0, len(contacts))
	for _, c := range contacts {
		if c.ID == uuid.Nil {
			continue
		}
		out = append(out, ports.ClientContactSnapshot{
			ID:        c.ID,
			Nom:       c.Nom,
			Prenom:    c.Prenom,
			Email:     c.Email,
			Role:      c.Role,
			Telephone: c.Telephone,
		})
	}
	return out, nil
}

func (r *Repository) PurgeClientContactsFromMissions(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID, removedIDs []uuid.UUID) error {
	if len(removedIDs) == 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE ssii.missions
		SET
			client_contact_ids = COALESCE((
				SELECT ARRAY_AGG(x ORDER BY ord)
				FROM UNNEST(client_contact_ids) WITH ORDINALITY AS t(x, ord)
				WHERE NOT (x = ANY($3::uuid[]))
			), '{}'::uuid[]),
			client_contact = CASE
				WHEN NOT EXISTS (
					SELECT 1
					FROM UNNEST(client_contact_ids) AS x
					WHERE NOT (x = ANY($3::uuid[]))
				) THEN ''
				ELSE client_contact
			END
		WHERE tenant_id = $1
			AND client_id = $2
			AND client_contact_ids && $3::uuid[]
	`, tenant.UUID(), clientID, removedIDs)
	return err
}

var _ ports.SSIIRepository = (*Repository)(nil)
