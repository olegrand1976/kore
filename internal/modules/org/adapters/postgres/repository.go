package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveTenant(ctx context.Context, tenant domain.Tenant) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.tenants (id, name) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
	`, tenant.ID, tenant.Name)
	return err
}

func (r *Repository) GetTenant(ctx context.Context, id kernel.TenantID) (domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.pool.QueryRow(ctx, `SELECT id, name FROM org.tenants WHERE id = $1`, id.UUID()).Scan(&tenant.ID, &tenant.Name)
	if err != nil {
		return domain.Tenant{}, err
	}
	return tenant, nil
}

func (r *Repository) SaveSociete(ctx context.Context, s domain.Societe) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.societes (id, tenant_id, raison_sociale, logo, devise, adresse, siret, url_tenant)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, s.ID, s.TenantID.UUID(), s.RaisonSociale, nullString(s.Logo), s.Devise,
		nullString(s.Adresse), nullString(s.Siret), nullString(s.URLTenant))
	return err
}

func (r *Repository) UpdateSociete(ctx context.Context, s domain.Societe) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE org.societes
		SET raison_sociale = $3, logo = $4, adresse = $5, siret = $6, url_tenant = $7
		WHERE tenant_id = $1 AND id = $2
	`, s.TenantID.UUID(), s.ID, s.RaisonSociale, nullString(s.Logo),
		nullString(s.Adresse), nullString(s.Siret), nullString(s.URLTenant))
	return err
}

func (r *Repository) GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Societe, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, raison_sociale, COALESCE(logo, ''), devise,
		       COALESCE(adresse, ''), COALESCE(siret, ''), COALESCE(url_tenant, '')
		FROM org.societes WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id)
	return scanSociete(row)
}

func (r *Repository) ListSocietes(ctx context.Context, tenant kernel.TenantID) ([]domain.Societe, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, raison_sociale, COALESCE(logo, ''), devise,
		       COALESCE(adresse, ''), COALESCE(siret, ''), COALESCE(url_tenant, '')
		FROM org.societes WHERE tenant_id = $1 ORDER BY raison_sociale
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Societe
	for rows.Next() {
		s, err := scanSocieteRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func scanSociete(row pgx.Row) (domain.Societe, error) {
	var s domain.Societe
	var tenantID uuid.UUID
	var logo, adresse, siret, urlTenant string
	err := row.Scan(&s.ID, &tenantID, &s.RaisonSociale, &logo, &s.Devise, &adresse, &siret, &urlTenant)
	if err != nil {
		return domain.Societe{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantID)
	s.Logo = logo
	s.Adresse = adresse
	s.Siret = siret
	s.URLTenant = urlTenant
	return s, nil
}

func scanSocieteRow(rows pgx.Rows) (domain.Societe, error) {
	var s domain.Societe
	var tenantID uuid.UUID
	var logo, adresse, siret, urlTenant string
	if err := rows.Scan(&s.ID, &tenantID, &s.RaisonSociale, &logo, &s.Devise, &adresse, &siret, &urlTenant); err != nil {
		return domain.Societe{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantID)
	s.Logo = logo
	s.Adresse = adresse
	s.Siret = siret
	s.URLTenant = urlTenant
	return s, nil
}

func nullString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func (r *Repository) SaveSite(ctx context.Context, s domain.Site) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.sites (id, tenant_id, societe_id, libelle) VALUES ($1, $2, $3, $4)
	`, s.ID, s.TenantID.UUID(), s.SocieteID, s.Libelle)
	return err
}

func (r *Repository) SaveService(ctx context.Context, s domain.Service) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.services (id, tenant_id, site_id, responsable_id) VALUES ($1, $2, $3, $4)
	`, s.ID, s.TenantID.UUID(), s.SiteID, s.ResponsableID)
	return err
}

func (r *Repository) SaveApplication(ctx context.Context, a domain.Application) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.applications (id, tenant_id, service_id, libelle) VALUES ($1, $2, $3, $4)
	`, a.ID, a.TenantID.UUID(), a.ServiceID, a.Libelle)
	return err
}

func (r *Repository) SaveUser(ctx context.Context, u domain.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.users (
			id, tenant_id, equipe_id, login, password_hash, profil, date_activation, date_expiration, active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, u.ID, u.TenantID.UUID(), u.EquipeID, string(u.Login), u.PasswordHash, string(u.Profile),
		u.Period.Activation, u.Period.Expiration, u.Active)
	return err
}

func (r *Repository) FindUserByLogin(ctx context.Context, tenant kernel.TenantID, login string) (domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, equipe_id, login, password_hash, profil, date_activation, date_expiration, active
		FROM org.users WHERE tenant_id = $1 AND login = $2
	`, tenant.UUID(), login))
}

func (r *Repository) FindUserByLoginGlobal(ctx context.Context, login string) (domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, equipe_id, login, password_hash, profil, date_activation, date_expiration, active
		FROM org.users WHERE login = $1 LIMIT 1
	`, login))
}

func (r *Repository) ExistsLogin(ctx context.Context, tenant kernel.TenantID, login string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM org.users WHERE tenant_id = $1 AND login = $2)`, tenant.UUID(), login).Scan(&exists)
	return exists, err
}

func (r *Repository) CountActiveUsers(ctx context.Context, tenant kernel.TenantID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM org.users WHERE tenant_id = $1 AND active = TRUE`, tenant.UUID()).Scan(&count)
	return count, err
}

func (r *Repository) ListUsers(ctx context.Context, tenant kernel.TenantID) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, equipe_id, login, password_hash, profil, date_activation, date_expiration, active
		FROM org.users WHERE tenant_id = $1 ORDER BY login
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *Repository) SaveClient(ctx context.Context, c domain.Client) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.clients (id, tenant_id, raison_sociale, tva, archived)
		VALUES ($1, $2, $3, $4, $5)
	`, c.ID, c.TenantID.UUID(), c.RaisonSociale, c.TVA, c.Archived)
	return err
}

func (r *Repository) ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, raison_sociale, tva, archived FROM org.clients WHERE tenant_id = $1 AND archived = FALSE
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Client
	for rows.Next() {
		var c domain.Client
		var tenantID uuid.UUID
		if err := rows.Scan(&c.ID, &tenantID, &c.RaisonSociale, &c.TVA, &c.Archived); err != nil {
			return nil, err
		}
		c.TenantID = kernel.NewTenantID(tenantID)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) GetPermissions(ctx context.Context) (map[string]map[authx.Module]map[authx.Action]bool, error) {
	rows, err := r.pool.Query(ctx, `SELECT profile, module, action FROM org.authx_permissions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]map[authx.Module]map[authx.Action]bool)
	for rows.Next() {
		var profile, module, action string
		if err := rows.Scan(&profile, &module, &action); err != nil {
			return nil, err
		}
		if out[profile] == nil {
			out[profile] = make(map[authx.Module]map[authx.Action]bool)
		}
		if out[profile][authx.Module(module)] == nil {
			out[profile][authx.Module(module)] = make(map[authx.Action]bool)
		}
		out[profile][authx.Module(module)][authx.Action(action)] = true
	}
	return out, rows.Err()
}

func (r *Repository) ResolveUserEmails(ctx context.Context, tenant kernel.TenantID, userIDs []uuid.UUID) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT login FROM org.users
		WHERE tenant_id = $1 AND id = ANY($2) AND active = TRUE
	`, tenant.UUID(), userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var emails []string
	for rows.Next() {
		var login string
		if err := rows.Scan(&login); err != nil {
			return nil, err
		}
		emails = append(emails, login+"@kore.local")
	}
	return emails, rows.Err()
}

func (r *Repository) scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	var tenantID uuid.UUID
	var login string
	var profile string
	var expiration *time.Time
	err := row.Scan(&u.ID, &tenantID, &u.EquipeID, &login, &u.PasswordHash, &profile,
		&u.Period.Activation, &expiration, &u.Active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user not found: %w", err)
		}
		return domain.User{}, err
	}
	u.TenantID = kernel.NewTenantID(tenantID)
	u.Login = domain.Login(login)
	u.Profile = domain.Profile(profile)
	u.Period.Expiration = expiration
	return u, nil
}

var _ ports.OrganizationRepository = (*Repository)(nil)
