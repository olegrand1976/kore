package postgres

import (
	"context"
	"encoding/json"
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
	pays := s.Pays
	if pays == "" {
		pays = "FR"
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.societes (
			id, tenant_id, raison_sociale, logo, devise, pays, week_start_day,
			day_capacity_minutes, cra_mail_auto, week_submit_policy,
			adresse, siret, url_tenant
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			raison_sociale = EXCLUDED.raison_sociale,
			devise = EXCLUDED.devise,
			pays = EXCLUDED.pays,
			week_start_day = EXCLUDED.week_start_day,
			day_capacity_minutes = EXCLUDED.day_capacity_minutes,
			cra_mail_auto = EXCLUDED.cra_mail_auto,
			week_submit_policy = EXCLUDED.week_submit_policy,
			adresse = EXCLUDED.adresse,
			siret = EXCLUDED.siret,
			url_tenant = EXCLUDED.url_tenant
	`, s.ID, s.TenantID.UUID(), s.RaisonSociale, nullString(s.Logo), s.Devise, pays,
		normalizeWeekStartDay(s.WeekStartDay),
		normalizeDayCapacityMinutes(s.DayCapacityMinutes),
		s.CraMailAuto,
		normalizeWeekSubmitPolicy(s.WeekSubmitPolicy),
		s.Adresse, s.Siret, s.URLTenant)
	return err
}

func (r *Repository) UpdateSociete(ctx context.Context, s domain.Societe) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE org.societes
		SET raison_sociale = $3, logo = $4, adresse = $5, siret = $6, url_tenant = $7,
			week_start_day = $8, day_capacity_minutes = $9, cra_mail_auto = $10, week_submit_policy = $11
		WHERE tenant_id = $1 AND id = $2
	`, s.TenantID.UUID(), s.ID, s.RaisonSociale, nullString(s.Logo),
		s.Adresse, s.Siret, s.URLTenant,
		normalizeWeekStartDay(s.WeekStartDay),
		normalizeDayCapacityMinutes(s.DayCapacityMinutes),
		s.CraMailAuto,
		normalizeWeekSubmitPolicy(s.WeekSubmitPolicy))
	return err
}

func (r *Repository) GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Societe, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, raison_sociale, COALESCE(logo, ''), devise, COALESCE(pays, 'FR'),
		       COALESCE(week_start_day, 1),
		       COALESCE(day_capacity_minutes, 480),
		       COALESCE(cra_mail_auto, FALSE),
		       COALESCE(week_submit_policy, 'warn'),
		       COALESCE(adresse, ''), COALESCE(siret, ''), COALESCE(url_tenant, '')
		FROM org.societes WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id)
	return scanSociete(row)
}

func (r *Repository) ListSocietes(ctx context.Context, tenant kernel.TenantID) ([]domain.Societe, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, raison_sociale, COALESCE(logo, ''), devise, COALESCE(pays, 'FR'),
		       COALESCE(week_start_day, 1),
		       COALESCE(day_capacity_minutes, 480),
		       COALESCE(cra_mail_auto, FALSE),
		       COALESCE(week_submit_policy, 'warn'),
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
	var logo, adresse, siret, urlTenant, pays string
	var weekStartDay, dayCapacity int
	var craMailAuto bool
	var weekSubmitPolicy string
	err := row.Scan(&s.ID, &tenantID, &s.RaisonSociale, &logo, &s.Devise, &pays,
		&weekStartDay, &dayCapacity, &craMailAuto, &weekSubmitPolicy,
		&adresse, &siret, &urlTenant)
	if err != nil {
		return domain.Societe{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantID)
	s.Logo = logo
	s.Pays = pays
	s.WeekStartDay = normalizeWeekStartDay(weekStartDay)
	s.DayCapacityMinutes = normalizeDayCapacityMinutes(dayCapacity)
	s.CraMailAuto = craMailAuto
	s.WeekSubmitPolicy = normalizeWeekSubmitPolicy(weekSubmitPolicy)
	s.Adresse = adresse
	s.Siret = siret
	s.URLTenant = urlTenant
	return s, nil
}

func scanSocieteRow(rows pgx.Rows) (domain.Societe, error) {
	var s domain.Societe
	var tenantID uuid.UUID
	var logo, adresse, siret, urlTenant, pays string
	var weekStartDay, dayCapacity int
	var craMailAuto bool
	var weekSubmitPolicy string
	if err := rows.Scan(&s.ID, &tenantID, &s.RaisonSociale, &logo, &s.Devise, &pays,
		&weekStartDay, &dayCapacity, &craMailAuto, &weekSubmitPolicy,
		&adresse, &siret, &urlTenant); err != nil {
		return domain.Societe{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantID)
	s.Logo = logo
	s.Pays = pays
	s.WeekStartDay = normalizeWeekStartDay(weekStartDay)
	s.DayCapacityMinutes = normalizeDayCapacityMinutes(dayCapacity)
	s.CraMailAuto = craMailAuto
	s.WeekSubmitPolicy = normalizeWeekSubmitPolicy(weekSubmitPolicy)
	s.Adresse = adresse
	s.Siret = siret
	s.URLTenant = urlTenant
	return s, nil
}

func normalizeWeekStartDay(day int) int {
	if day < 0 || day > 6 {
		return domain.DefaultWeekStartDay
	}
	return day
}

func normalizeDayCapacityMinutes(minutes int) int {
	if minutes <= 0 || minutes > 1440 {
		return domain.DefaultDayCapacityMinutes
	}
	return minutes
}

func normalizeWeekSubmitPolicy(policy string) string {
	switch policy {
	case "block", "warn", "none":
		return policy
	default:
		return domain.DefaultWeekSubmitPolicy
	}
}

func nullString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func (r *Repository) SaveSite(ctx context.Context, s domain.Site) error {
	pays := s.Pays
	if pays == "" {
		pays = "FR"
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.sites (id, tenant_id, societe_id, libelle, pays) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET libelle = EXCLUDED.libelle, pays = EXCLUDED.pays
	`, s.ID, s.TenantID.UUID(), s.SocieteID, s.Libelle, pays)
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

func (r *Repository) ListApplications(ctx context.Context, tenant kernel.TenantID) ([]domain.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, service_id, libelle,
		       COALESCE(proprietaire, ''), COALESCE(mode_facturation, 'temps_passe'), COALESCE(uo_activee, FALSE)
		FROM org.applications
		WHERE tenant_id = $1
		ORDER BY libelle
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Application
	for rows.Next() {
		app, err := scanApplicationRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, app)
	}
	return out, rows.Err()
}

func (r *Repository) GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Application, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, service_id, libelle,
		       COALESCE(proprietaire, ''), COALESCE(mode_facturation, 'temps_passe'), COALESCE(uo_activee, FALSE)
		FROM org.applications
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id)
	return scanApplication(row)
}

func scanApplication(row pgx.Row) (domain.Application, error) {
	var app domain.Application
	var tenantID uuid.UUID
	var proprietaire, modeFacturation string
	if err := row.Scan(
		&app.ID, &tenantID, &app.ServiceID, &app.Libelle,
		&proprietaire, &modeFacturation, &app.UOActivee,
	); err != nil {
		return domain.Application{}, err
	}
	app.TenantID = kernel.NewTenantID(tenantID)
	app.Proprietaire = proprietaire
	app.ModeFacturation = modeFacturation
	return app, nil
}

func scanApplicationRow(rows pgx.Rows) (domain.Application, error) {
	var app domain.Application
	var tenantID uuid.UUID
	var proprietaire, modeFacturation string
	if err := rows.Scan(
		&app.ID, &tenantID, &app.ServiceID, &app.Libelle,
		&proprietaire, &modeFacturation, &app.UOActivee,
	); err != nil {
		return domain.Application{}, err
	}
	app.TenantID = kernel.NewTenantID(tenantID)
	app.Proprietaire = proprietaire
	app.ModeFacturation = modeFacturation
	return app, nil
}

func (r *Repository) SaveUser(ctx context.Context, u domain.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.users (
			id, tenant_id, equipe_id, login, prenom, nom, email, password_hash, profil, date_activation, date_expiration, active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, u.ID, u.TenantID.UUID(), u.EquipeID, string(u.Login), u.Prenom, u.Nom, nullString(u.Email), u.PasswordHash, string(u.Profile),
		u.Period.Activation, u.Period.Expiration, u.Active)
	return err
}

func (r *Repository) FindUserByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), id))
}

func (r *Repository) FindUserDetailByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (ports.UserDetail, error) {
	var detail ports.UserDetail
	var profile string
	var expiration *time.Time
	var activation time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT
			u.id, u.login, u.prenom, u.nom, COALESCE(u.email, ''), u.profil, u.active,
			u.langue, u.type_compte, u.cra_requis, u.salarie_ett,
			u.equipe_id, COALESCE(e.libelle, ''), u.date_activation, u.date_expiration
		FROM org.users u
		LEFT JOIN org.equipes e ON e.id = u.equipe_id
		WHERE u.tenant_id = $1 AND u.id = $2 AND u.deleted_at IS NULL
	`, tenant.UUID(), id).Scan(
		&detail.ID, &detail.Login, &detail.Prenom, &detail.Nom, &detail.Email, &profile, &detail.Active,
		&detail.Langue, &detail.TypeCompte, &detail.CraRequis, &detail.SalarieETT,
		&detail.EquipeID, &detail.EquipeLibelle, &activation, &expiration,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.UserDetail{}, fmt.Errorf("user not found: %w", err)
		}
		return ports.UserDetail{}, err
	}
	detail.Profile = profile
	detail.DateActivation = activation.Format("2006-01-02")
	if expiration != nil {
		formatted := expiration.Format("2006-01-02")
		detail.DateExpiration = &formatted
	}
	return detail, nil
}

func (r *Repository) GetReleaseNotesPreferences(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.ReleaseNotesPreferences, error) {
	var prefs ports.ReleaseNotesPreferences
	var lastSeen *string
	err := r.pool.QueryRow(ctx, `
		SELECT last_seen_version, COALESCE(release_notes_auto_show, TRUE)
		FROM org.users
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), userID).Scan(&lastSeen, &prefs.AutoShowEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ReleaseNotesPreferences{}, fmt.Errorf("user not found: %w", err)
		}
		return ports.ReleaseNotesPreferences{}, err
	}
	prefs.LastSeenVersion = lastSeen
	return prefs, nil
}

func (r *Repository) SetReleaseNotesAutoShow(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, enabled bool) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.users
		SET release_notes_auto_show = $3
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), userID, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %w", pgx.ErrNoRows)
	}
	return nil
}

func (r *Repository) SetLastSeenVersion(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, version string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.users
		SET last_seen_version = $3
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), userID, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %w", pgx.ErrNoRows)
	}
	return nil
}

func (r *Repository) UpdateUser(ctx context.Context, u domain.User) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.users
		SET profil = $3, password_hash = $4, active = $5, email = $6
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, u.TenantID.UUID(), u.ID, string(u.Profile), u.PasswordHash, u.Active, nullString(u.Email))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %w", pgx.ErrNoRows)
	}
	return nil
}

func (r *Repository) SoftDeleteUser(ctx context.Context, tenant kernel.TenantID, id uuid.UUID, deletedAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.users
		SET active = FALSE, deleted_at = $3
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), id, deletedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %w", pgx.ErrNoRows)
	}
	return nil
}

func (r *Repository) FindUserByLogin(ctx context.Context, tenant kernel.TenantID, login string) (domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE tenant_id = $1 AND login = $2 AND deleted_at IS NULL
	`, tenant.UUID(), login))
}

func (r *Repository) FindUserByLoginGlobal(ctx context.Context, login string) (domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE login = $1 AND deleted_at IS NULL LIMIT 1
	`, login))
}

func (r *Repository) ExistsLogin(ctx context.Context, tenant kernel.TenantID, login string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM org.users WHERE tenant_id = $1 AND login = $2)`, tenant.UUID(), login).Scan(&exists)
	return exists, err
}

func (r *Repository) CountActiveUsers(ctx context.Context, tenant kernel.TenantID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM org.users
		WHERE tenant_id = $1 AND active = TRUE AND deleted_at IS NULL
	`, tenant.UUID()).Scan(&count)
	return count, err
}

func (r *Repository) ListUsers(ctx context.Context, tenant kernel.TenantID) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY login
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

func (r *Repository) GetClient(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Client, error) {
	var c domain.Client
	var tenantID uuid.UUID
	var contacts []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, raison_sociale, tva, contacts, archived, created_at
		FROM org.clients WHERE tenant_id = $1 AND id = $2 AND archived = FALSE
	`, tenant.UUID(), id).Scan(&c.ID, &tenantID, &c.RaisonSociale, &c.TVA, &contacts, &c.Archived, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, fmt.Errorf("client not found: %w", err)
		}
		return domain.Client{}, err
	}
	c.TenantID = kernel.NewTenantID(tenantID)
	if len(contacts) > 0 {
		_ = json.Unmarshal(contacts, &c.Contacts)
	}
	return c, nil
}

func (r *Repository) ListClients(ctx context.Context, tenant kernel.TenantID) ([]domain.Client, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, raison_sociale, tva, contacts, archived, created_at
		FROM org.clients WHERE tenant_id = $1 AND archived = FALSE
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Client
	for rows.Next() {
		var c domain.Client
		var tenantID uuid.UUID
		var contacts []byte
		if err := rows.Scan(&c.ID, &tenantID, &c.RaisonSociale, &c.TVA, &contacts, &c.Archived, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.TenantID = kernel.NewTenantID(tenantID)
		if len(contacts) > 0 {
			_ = json.Unmarshal(contacts, &c.Contacts)
		}
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

func (r *Repository) ResolveEquipeUserEmails(ctx context.Context, tenant kernel.TenantID, equipeID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT login FROM org.users
		WHERE tenant_id = $1 AND equipe_id = $2 AND active = TRUE
	`, tenant.UUID(), equipeID)
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

func (r *Repository) ResolveApplicationUserEmails(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.login
		FROM org.users u
		JOIN org.equipes e ON e.id = u.equipe_id AND e.tenant_id = u.tenant_id
		WHERE u.tenant_id = $1 AND e.application_id = $2 AND u.active = TRUE
	`, tenant.UUID(), applicationID)
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

func (r *Repository) ResolveServiceUserEmails(ctx context.Context, tenant kernel.TenantID, serviceID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.login
		FROM org.users u
		JOIN org.equipes e ON e.id = u.equipe_id AND e.tenant_id = u.tenant_id
		JOIN org.applications a ON a.id = e.application_id AND a.tenant_id = u.tenant_id
		WHERE u.tenant_id = $1 AND a.service_id = $2 AND u.active = TRUE
	`, tenant.UUID(), serviceID)
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

func (r *Repository) ResolveSocieteIDForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (uuid.UUID, error) {
	var societeID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT st.societe_id
		FROM org.users u
		LEFT JOIN org.equipes e ON e.id = u.equipe_id
		LEFT JOIN org.applications a ON a.id = e.application_id
		LEFT JOIN org.services sv ON sv.id = a.service_id
		LEFT JOIN org.sites st ON st.id = sv.site_id
		WHERE u.tenant_id = $1 AND u.id = $2
	`, tenant.UUID(), userID).Scan(&societeID)
	if err == nil && societeID != uuid.Nil {
		return societeID, nil
	}
	err = r.pool.QueryRow(ctx, `
		SELECT id FROM org.societes WHERE tenant_id = $1 ORDER BY raison_sociale LIMIT 1
	`, tenant.UUID()).Scan(&societeID)
	if err != nil {
		return uuid.Nil, err
	}
	return societeID, nil
}

func (r *Repository) scanUser(row pgx.Row) (domain.User, error) {
	var u domain.User
	var tenantID uuid.UUID
	var login string
	var profile string
	var email *string
	var expiration *time.Time
	err := row.Scan(&u.ID, &tenantID, &u.EquipeID, &login, &u.Prenom, &u.Nom, &email, &u.PasswordHash, &profile,
		&u.Period.Activation, &expiration, &u.Active, &u.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf("user not found: %w", err)
		}
		return domain.User{}, err
	}
	u.TenantID = kernel.NewTenantID(tenantID)
	u.Login = domain.Login(login)
	if email != nil {
		u.Email = *email
	}
	u.Profile = domain.Profile(profile)
	u.Period.Expiration = expiration
	return u, nil
}

const userSelectCols = `id, tenant_id, equipe_id, login, prenom, nom, email, password_hash, profil, date_activation, date_expiration, active, deleted_at`

func (r *Repository) SaveIdentityProvider(ctx context.Context, idp domain.IdentityProvider) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.identity_providers (
			id, tenant_id, name, issuer, client_id, client_secret, jwks_uri, scopes, default_profile, enabled, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			name = EXCLUDED.name,
			issuer = EXCLUDED.issuer,
			client_id = EXCLUDED.client_id,
			client_secret = CASE WHEN EXCLUDED.client_secret = '' THEN org.identity_providers.client_secret ELSE EXCLUDED.client_secret END,
			jwks_uri = EXCLUDED.jwks_uri,
			scopes = EXCLUDED.scopes,
			default_profile = EXCLUDED.default_profile,
			enabled = EXCLUDED.enabled,
			updated_at = NOW()
	`, idp.ID, idp.TenantID.UUID(), idp.Name, idp.Issuer, idp.ClientID, idp.ClientSecret,
		idp.JWKSURI, idp.Scopes, string(idp.DefaultProfile), idp.Enabled)
	return err
}

func (r *Repository) GetIdentityProvider(ctx context.Context, tenant kernel.TenantID) (domain.IdentityProvider, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, issuer, client_id, client_secret, jwks_uri, scopes, default_profile, enabled
		FROM org.identity_providers WHERE tenant_id = $1
	`, tenant.UUID())
	var idp domain.IdentityProvider
	var tenantID uuid.UUID
	var profile string
	err := row.Scan(&idp.ID, &tenantID, &idp.Name, &idp.Issuer, &idp.ClientID, &idp.ClientSecret,
		&idp.JWKSURI, &idp.Scopes, &profile, &idp.Enabled)
	if err != nil {
		return domain.IdentityProvider{}, err
	}
	idp.TenantID = kernel.NewTenantID(tenantID)
	idp.DefaultProfile = domain.Profile(profile)
	return idp, nil
}

func (r *Repository) ListIdentityProviders(ctx context.Context, tenant kernel.TenantID) ([]domain.IdentityProvider, error) {
	idp, err := r.GetIdentityProvider(ctx, tenant)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return []domain.IdentityProvider{idp}, nil
}

func (r *Repository) LinkUserIdentity(ctx context.Context, link domain.UserIdentityLink) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.user_identities (id, tenant_id, user_id, idp_id, subject, email)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, idp_id, subject) DO NOTHING
	`, link.ID, link.TenantID.UUID(), link.UserID, link.IdPID, link.Subject, link.Email)
	return err
}

func (r *Repository) FindUserIdentityBySubject(ctx context.Context, tenant kernel.TenantID, idpID uuid.UUID, subject string) (domain.UserIdentityLink, error) {
	var link domain.UserIdentityLink
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, idp_id, subject, email
		FROM org.user_identities WHERE tenant_id = $1 AND idp_id = $2 AND subject = $3
	`, tenant.UUID(), idpID, subject).Scan(&link.ID, &tenantID, &link.UserID, &link.IdPID, &link.Subject, &link.Email)
	if err != nil {
		return domain.UserIdentityLink{}, err
	}
	link.TenantID = kernel.NewTenantID(tenantID)
	return link, nil
}

func (r *Repository) FindUserByEmail(ctx context.Context, tenant kernel.TenantID, email string) (domain.User, error) {
	return r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE tenant_id = $1 AND lower(email) = lower($2) AND deleted_at IS NULL
	`, tenant.UUID(), email))
}

func (r *Repository) FindTenantIDsByEmail(ctx context.Context, email string) ([]kernel.TenantID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT tenant_id
		FROM org.users
		WHERE lower(email) = lower($1) AND deleted_at IS NULL
	`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]kernel.TenantID, 0, 1)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, kernel.NewTenantID(id))
	}
	return out, rows.Err()
}

func (r *Repository) SaveAccessToken(ctx context.Context, tokenHash string, tenant kernel.TenantID, email, kind string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.access_tokens (token_hash, tenant_id, email, kind, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, tokenHash, tenant.UUID(), email, kind, expiresAt)
	return err
}

func (r *Repository) ConsumeAccessToken(ctx context.Context, tokenHash string, now time.Time) (ports.AccessTokenRow, bool, error) {
	var row ports.AccessTokenRow
	var tenantID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		UPDATE org.access_tokens
		SET used_at = $2
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > $2
		RETURNING token_hash, tenant_id, email, kind, expires_at, used_at, created_at
	`, tokenHash, now).Scan(&row.TokenHash, &tenantID, &row.Email, &row.Kind, &row.ExpiresAt, &row.UsedAt, &row.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Try to load the row to distinguish invalid / used / expired.
			var out ports.AccessTokenRow
			var tID uuid.UUID
			selErr := r.pool.QueryRow(ctx, `
				SELECT token_hash, tenant_id, email, kind, expires_at, used_at, created_at
				FROM org.access_tokens
				WHERE token_hash = $1
			`, tokenHash).Scan(&out.TokenHash, &tID, &out.Email, &out.Kind, &out.ExpiresAt, &out.UsedAt, &out.CreatedAt)
			if selErr != nil {
				if errors.Is(selErr, pgx.ErrNoRows) {
					return ports.AccessTokenRow{}, false, nil
				}
				return ports.AccessTokenRow{}, false, selErr
			}
			out.TenantID = kernel.NewTenantID(tID)
			return out, false, nil
		}
		return ports.AccessTokenRow{}, false, err
	}
	row.TenantID = kernel.NewTenantID(tenantID)
	return row, true, nil
}

var _ ports.OrganizationRepository = (*Repository)(nil)
