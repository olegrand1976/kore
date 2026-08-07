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

func (r *Repository) ProvisionCore(ctx context.Context, tenant domain.Tenant, societe domain.Societe, admin domain.User) error {
	hydrateEquipeIDsFromPrimary(&admin)
	admin.SyncPrimaryMemberships()
	pays := societe.Pays
	if pays == "" {
		pays = "FR"
	}
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.tenants (id, name) VALUES ($1, $2)
		`, tenant.ID, tenant.Name); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.societes (
				id, tenant_id, raison_sociale, logo, devise, pays, week_start_day,
				day_capacity_minutes, cra_mail_auto, week_submit_policy, cra_gate_mode,
				adresse, adresse_numero, adresse_boite, code_postal, ville,
				siret, url_tenant, cra_mail_recipients,
				totp_default_enabled, totp_user_configurable, task_types_enabled, seed_protected,
				default_tjm_cents
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		`, societe.ID, societe.TenantID.UUID(), societe.RaisonSociale, nullString(societe.Logo), societe.Devise, pays,
			normalizeWeekStartDay(societe.WeekStartDay),
			normalizeDayCapacityMinutes(societe.DayCapacityMinutes),
			societe.CraMailAuto,
			normalizeWeekSubmitPolicy(societe.WeekSubmitPolicy),
			normalizeCraGateMode(societe.CraGateMode),
			societe.Adresse, societe.AdresseNumero, societe.AdresseBoite, societe.CodePostal, societe.Ville,
			societe.Siret, societe.URLTenant, encodeMailRecipients(societe.CraMailRecipients),
			societe.TotpDefaultEnabled, societe.TotpUserConfigurable, encodeTaskTypes(societe.TaskTypesEnabled),
			societe.SeedProtected, societe.DefaultTJMCents); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.users (
				id, tenant_id, equipe_id, login, prenom, nom, email, password_hash, profil,
				date_activation, date_expiration, active,
				totp_enabled, totp_enrollment_required
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`, admin.ID, admin.TenantID.UUID(), admin.EquipeID, string(admin.Login), admin.Prenom, admin.Nom, nullString(admin.Email), admin.PasswordHash, string(admin.Profile),
			admin.Period.Activation, admin.Period.Expiration, admin.Active, admin.TotpEnabled, admin.TotpEnrollmentRequired); err != nil {
			return err
		}
		return replaceUserMemberships(ctx, tx, admin)
	})
}

func (r *Repository) RollbackProvision(ctx context.Context, tenantID kernel.TenantID) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM org.user_profiles WHERE user_id IN (SELECT id FROM org.users WHERE tenant_id = $1)`, tenantID.UUID()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM org.user_equipes WHERE user_id IN (SELECT id FROM org.users WHERE tenant_id = $1)`, tenantID.UUID()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM org.users WHERE tenant_id = $1`, tenantID.UUID()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM org.societes WHERE tenant_id = $1`, tenantID.UUID()); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM org.tenants WHERE id = $1`, tenantID.UUID()); err != nil {
			return err
		}
		return nil
	})
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
			day_capacity_minutes, cra_mail_auto, week_submit_policy, cra_gate_mode,
			adresse, adresse_numero, adresse_boite, code_postal, ville,
			siret, url_tenant, cra_mail_recipients,
			totp_default_enabled, totp_user_configurable, task_types_enabled, seed_protected,
			default_tjm_cents
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		ON CONFLICT (id) DO UPDATE SET
			raison_sociale = EXCLUDED.raison_sociale,
			devise = EXCLUDED.devise,
			pays = EXCLUDED.pays,
			week_start_day = EXCLUDED.week_start_day,
			day_capacity_minutes = EXCLUDED.day_capacity_minutes,
			cra_mail_auto = EXCLUDED.cra_mail_auto,
			week_submit_policy = EXCLUDED.week_submit_policy,
			cra_gate_mode = EXCLUDED.cra_gate_mode,
			adresse = EXCLUDED.adresse,
			adresse_numero = EXCLUDED.adresse_numero,
			adresse_boite = EXCLUDED.adresse_boite,
			code_postal = EXCLUDED.code_postal,
			ville = EXCLUDED.ville,
			siret = EXCLUDED.siret,
			url_tenant = EXCLUDED.url_tenant,
			cra_mail_recipients = EXCLUDED.cra_mail_recipients,
			totp_default_enabled = EXCLUDED.totp_default_enabled,
			totp_user_configurable = EXCLUDED.totp_user_configurable,
			task_types_enabled = EXCLUDED.task_types_enabled,
			seed_protected = EXCLUDED.seed_protected,
			default_tjm_cents = EXCLUDED.default_tjm_cents
	`, s.ID, s.TenantID.UUID(), s.RaisonSociale, nullString(s.Logo), s.Devise, pays,
		normalizeWeekStartDay(s.WeekStartDay),
		normalizeDayCapacityMinutes(s.DayCapacityMinutes),
		s.CraMailAuto,
		normalizeWeekSubmitPolicy(s.WeekSubmitPolicy),
		normalizeCraGateMode(s.CraGateMode),
		s.Adresse, s.AdresseNumero, s.AdresseBoite, s.CodePostal, s.Ville,
		s.Siret, s.URLTenant, encodeMailRecipients(s.CraMailRecipients),
		s.TotpDefaultEnabled, s.TotpUserConfigurable, encodeTaskTypes(s.TaskTypesEnabled),
		s.SeedProtected, s.DefaultTJMCents)
	return err
}

func (r *Repository) UpdateSociete(ctx context.Context, s domain.Societe) error {
	pays := s.Pays
	if pays == "" {
		pays = "FR"
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE org.societes
		SET raison_sociale = $3, logo = $4, adresse = $5, adresse_numero = $6, adresse_boite = $7,
			code_postal = $8, ville = $9, siret = $10, url_tenant = $11, pays = $12,
			week_start_day = $13, day_capacity_minutes = $14, cra_mail_auto = $15, week_submit_policy = $16,
			cra_gate_mode = $17, cra_mail_recipients = $18, totp_default_enabled = $19, totp_user_configurable = $20,
			task_types_enabled = $21, seed_protected = $22, default_tjm_cents = $23
		WHERE tenant_id = $1 AND id = $2
	`, s.TenantID.UUID(), s.ID, s.RaisonSociale, nullString(s.Logo),
		s.Adresse, s.AdresseNumero, s.AdresseBoite, s.CodePostal, s.Ville,
		s.Siret, s.URLTenant, pays,
		normalizeWeekStartDay(s.WeekStartDay),
		normalizeDayCapacityMinutes(s.DayCapacityMinutes),
		s.CraMailAuto,
		normalizeWeekSubmitPolicy(s.WeekSubmitPolicy),
		normalizeCraGateMode(s.CraGateMode),
		encodeMailRecipients(s.CraMailRecipients),
		s.TotpDefaultEnabled, s.TotpUserConfigurable, encodeTaskTypes(s.TaskTypesEnabled),
		s.SeedProtected, s.DefaultTJMCents)
	return err
}

func (r *Repository) SaveSocieteLogo(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID, content []byte, contentType string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.societes
		SET logo_content = $3, logo_content_type = $4
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), societeID, content, nullString(contentType))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSocieteNotFound
	}
	return nil
}

func (r *Repository) GetTenantLogo(ctx context.Context, tenant kernel.TenantID) ([]byte, string, error) {
	var content []byte
	var contentType *string
	err := r.pool.QueryRow(ctx, `
		SELECT logo_content, logo_content_type
		FROM org.societes
		WHERE tenant_id = $1 AND logo_content IS NOT NULL
		ORDER BY created_at ASC
		LIMIT 1
	`, tenant.UUID()).Scan(&content, &contentType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", domain.ErrLogoNotFound
		}
		return nil, "", err
	}
	ct := "application/octet-stream"
	if contentType != nil && *contentType != "" {
		ct = *contentType
	}
	return content, ct, nil
}

func (r *Repository) ListSocietesCraMailAuto(ctx context.Context) ([]ports.CraMailReminderTarget, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tenant_id, id, COALESCE(pays, 'FR'), COALESCE(cra_mail_recipients, '[]')
		FROM org.societes
		WHERE cra_mail_auto = TRUE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ports.CraMailReminderTarget
	for rows.Next() {
		var target ports.CraMailReminderTarget
		var tenantID, societeID uuid.UUID
		var pays string
		var recipientsRaw []byte
		if err := rows.Scan(&tenantID, &societeID, &pays, &recipientsRaw); err != nil {
			return nil, err
		}
		target.TenantID = kernel.NewTenantID(tenantID)
		target.SocieteID = societeID
		target.Pays = pays
		target.Recipients = decodeMailRecipients(recipientsRaw)
		out = append(out, target)
	}
	return out, rows.Err()
}

func (r *Repository) GetSociete(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Societe, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, raison_sociale, COALESCE(logo, ''), devise, COALESCE(pays, 'FR'),
		       COALESCE(week_start_day, 1),
		       COALESCE(day_capacity_minutes, 480),
		       COALESCE(cra_mail_auto, FALSE),
		       COALESCE(week_submit_policy, 'warn'),
		       COALESCE(cra_gate_mode, 'warn'),
		       COALESCE(cra_mail_recipients, '[]'),
		       COALESCE(adresse, ''), COALESCE(adresse_numero, ''), COALESCE(adresse_boite, ''),
		       COALESCE(code_postal, ''), COALESCE(ville, ''),
		       COALESCE(siret, ''), COALESCE(url_tenant, ''),
		       COALESCE(totp_default_enabled, FALSE), COALESCE(totp_user_configurable, TRUE),
		       COALESCE(task_types_enabled, '[]'),
		       COALESCE(seed_protected, FALSE),
		       COALESCE(default_tjm_cents, 0)
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
		       COALESCE(cra_gate_mode, 'warn'),
		       COALESCE(cra_mail_recipients, '[]'),
		       COALESCE(adresse, ''), COALESCE(adresse_numero, ''), COALESCE(adresse_boite, ''),
		       COALESCE(code_postal, ''), COALESCE(ville, ''),
		       COALESCE(siret, ''), COALESCE(url_tenant, ''),
		       COALESCE(totp_default_enabled, FALSE), COALESCE(totp_user_configurable, TRUE),
		       COALESCE(task_types_enabled, '[]'),
		       COALESCE(seed_protected, FALSE),
		       COALESCE(default_tjm_cents, 0)
		FROM org.societes WHERE tenant_id = $1
		ORDER BY raison_sociale
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
	var logo, adresse, adresseNumero, adresseBoite, codePostal, ville, siret, urlTenant, pays string
	var weekStartDay, dayCapacity int
	var craMailAuto, totpDefaultEnabled, totpUserConfigurable, seedProtected bool
	var weekSubmitPolicy, craGateMode string
	var recipientsRaw []byte
	var taskTypesRaw []byte
	err := row.Scan(&s.ID, &tenantID, &s.RaisonSociale, &logo, &s.Devise, &pays,
		&weekStartDay, &dayCapacity, &craMailAuto, &weekSubmitPolicy, &craGateMode, &recipientsRaw,
		&adresse, &adresseNumero, &adresseBoite, &codePostal, &ville,
		&siret, &urlTenant, &totpDefaultEnabled, &totpUserConfigurable, &taskTypesRaw,
		&seedProtected, &s.DefaultTJMCents)
	if err != nil {
		return domain.Societe{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantID)
	s.Logo = logo
	s.Pays = pays
	s.WeekStartDay = normalizeWeekStartDay(weekStartDay)
	s.DayCapacityMinutes = normalizeDayCapacityMinutes(dayCapacity)
	s.CraMailAuto = craMailAuto
	s.CraMailRecipients = decodeMailRecipients(recipientsRaw)
	s.TaskTypesEnabled = decodeTaskTypes(taskTypesRaw)
	s.WeekSubmitPolicy = normalizeWeekSubmitPolicy(weekSubmitPolicy)
	s.CraGateMode = normalizeCraGateMode(craGateMode)
	s.Adresse = adresse
	s.AdresseNumero = adresseNumero
	s.AdresseBoite = adresseBoite
	s.CodePostal = codePostal
	s.Ville = ville
	s.Siret = siret
	s.URLTenant = urlTenant
	s.TotpDefaultEnabled = totpDefaultEnabled
	s.TotpUserConfigurable = totpUserConfigurable
	s.SeedProtected = seedProtected
	return s, nil
}

func scanSocieteRow(rows pgx.Rows) (domain.Societe, error) {
	var s domain.Societe
	var tenantID uuid.UUID
	var logo, adresse, adresseNumero, adresseBoite, codePostal, ville, siret, urlTenant, pays string
	var weekStartDay, dayCapacity int
	var craMailAuto, totpDefaultEnabled, totpUserConfigurable, seedProtected bool
	var weekSubmitPolicy, craGateMode string
	var recipientsRaw []byte
	var taskTypesRaw []byte
	if err := rows.Scan(&s.ID, &tenantID, &s.RaisonSociale, &logo, &s.Devise, &pays,
		&weekStartDay, &dayCapacity, &craMailAuto, &weekSubmitPolicy, &craGateMode, &recipientsRaw,
		&adresse, &adresseNumero, &adresseBoite, &codePostal, &ville,
		&siret, &urlTenant, &totpDefaultEnabled, &totpUserConfigurable, &taskTypesRaw,
		&seedProtected, &s.DefaultTJMCents); err != nil {
		return domain.Societe{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantID)
	s.Logo = logo
	s.Pays = pays
	s.WeekStartDay = normalizeWeekStartDay(weekStartDay)
	s.DayCapacityMinutes = normalizeDayCapacityMinutes(dayCapacity)
	s.CraMailAuto = craMailAuto
	s.CraMailRecipients = decodeMailRecipients(recipientsRaw)
	s.TaskTypesEnabled = decodeTaskTypes(taskTypesRaw)
	s.WeekSubmitPolicy = normalizeWeekSubmitPolicy(weekSubmitPolicy)
	s.CraGateMode = normalizeCraGateMode(craGateMode)
	s.Adresse = adresse
	s.AdresseNumero = adresseNumero
	s.AdresseBoite = adresseBoite
	s.CodePostal = codePostal
	s.Ville = ville
	s.Siret = siret
	s.URLTenant = urlTenant
	s.TotpDefaultEnabled = totpDefaultEnabled
	s.TotpUserConfigurable = totpUserConfigurable
	s.SeedProtected = seedProtected
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

func normalizeCraGateMode(mode string) string {
	switch mode {
	case "block", "warn":
		return mode
	default:
		return domain.DefaultCraGateMode
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

func (r *Repository) UpdateSite(ctx context.Context, tenant kernel.TenantID, siteID uuid.UUID, libelle string) (domain.SiteSummary, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.sites SET libelle = $3
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), siteID, libelle)
	if err != nil {
		return domain.SiteSummary{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.SiteSummary{}, domain.ErrSiteNotFound
	}
	var item domain.SiteSummary
	err = r.pool.QueryRow(ctx, `
		SELECT id, societe_id, libelle, COALESCE(pays, '')
		FROM org.sites
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), siteID).Scan(&item.ID, &item.SocieteID, &item.Libelle, &item.Pays)
	if err != nil {
		return domain.SiteSummary{}, err
	}
	return item, nil
}

func (r *Repository) ListSites(ctx context.Context, tenant kernel.TenantID) ([]domain.SiteSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, societe_id, libelle, COALESCE(pays, '')
		FROM org.sites
		WHERE tenant_id = $1
		ORDER BY libelle
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SiteSummary
	for rows.Next() {
		var item domain.SiteSummary
		if err := rows.Scan(&item.ID, &item.SocieteID, &item.Libelle, &item.Pays); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) SaveService(ctx context.Context, s domain.Service) error {
	serviceType := s.Type
	if serviceType == "" {
		serviceType = domain.DefaultServiceType
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO org.services (id, tenant_id, site_id, libelle, type, responsable_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, s.ID, s.TenantID.UUID(), s.SiteID, s.Libelle, serviceType, s.ResponsableID)
	return err
}

func (r *Repository) UpdateService(ctx context.Context, tenant kernel.TenantID, serviceID uuid.UUID, libelle string) (domain.ServiceSummary, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.services SET libelle = $3
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), serviceID, libelle)
	if err != nil {
		return domain.ServiceSummary{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.ServiceSummary{}, domain.ErrServiceNotFound
	}
	var item domain.ServiceSummary
	err = r.pool.QueryRow(ctx, `
		SELECT s.id, s.site_id, COALESCE(st.libelle, ''), COALESCE(st.societe_id, '00000000-0000-0000-0000-000000000000'::uuid),
		       COALESCE(s.libelle, ''), COALESCE(s.type, ''), s.responsable_id
		FROM org.services s
		LEFT JOIN org.sites st ON st.id = s.site_id
		WHERE s.tenant_id = $1 AND s.id = $2
	`, tenant.UUID(), serviceID).Scan(
		&item.ID, &item.SiteID, &item.SiteLabel, &item.SocieteID,
		&item.Libelle, &item.Type, &item.ResponsableID,
	)
	if err != nil {
		return domain.ServiceSummary{}, err
	}
	return item, nil
}

func (r *Repository) SaveEquipe(ctx context.Context, e domain.Equipe) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO org.equipes (id, tenant_id, application_id, libelle, responsable_id)
			VALUES ($1, $2, $3, $4, $5)
		`, e.ID, e.TenantID.UUID(), e.ApplicationID, e.Libelle, e.ResponsableID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO org.application_equipes (tenant_id, application_id, equipe_id)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
		`, e.TenantID.UUID(), e.ApplicationID, e.ID)
		return err
	})
}

func (r *Repository) SaveApplication(ctx context.Context, a domain.Application) error {
	mode := a.ModeFacturation
	if mode == "" {
		mode = domain.DefaultModeFacturation
	}
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO org.applications (
				id, tenant_id, libelle, proprietaire, mode_facturation,
				uo_activee, chef_utilisateur_id, budget_defaut_id, active, default_tjm_cents
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, a.ID, a.TenantID.UUID(), a.Libelle, nullIfEmpty(a.Proprietaire), mode,
			a.UOActivee, a.ChefUtilisateurID, a.BudgetDefautID, a.Active, a.DefaultTJMCents)
		if err != nil {
			return err
		}
		return replaceApplicationShares(ctx, tx, a)
	})
}

func (r *Repository) UpdateApplication(ctx context.Context, a domain.Application, replaceShares bool) error {
	mode := a.ModeFacturation
	if mode == "" {
		mode = domain.DefaultModeFacturation
	}
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE org.applications
			SET libelle = $3, active = $4, proprietaire = $5, mode_facturation = $6,
			    uo_activee = $7, chef_utilisateur_id = $8, budget_defaut_id = $9, default_tjm_cents = $10
			WHERE tenant_id = $1 AND id = $2
		`, a.TenantID.UUID(), a.ID, a.Libelle, a.Active, nullIfEmpty(a.Proprietaire), mode,
			a.UOActivee, a.ChefUtilisateurID, a.BudgetDefautID, a.DefaultTJMCents)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrApplicationNotFound
		}
		if !replaceShares {
			return ensureHomeEquipeShares(ctx, tx, a.TenantID.UUID(), a.ID)
		}
		return replaceApplicationShares(ctx, tx, a)
	})
}

func replaceApplicationShares(ctx context.Context, tx pgx.Tx, a domain.Application) error {
	tenantID := a.TenantID.UUID()
	if _, err := tx.Exec(ctx, `DELETE FROM org.application_sites WHERE application_id = $1`, a.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM org.application_services WHERE application_id = $1`, a.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM org.application_equipes WHERE application_id = $1`, a.ID); err != nil {
		return err
	}
	for _, siteID := range a.SiteIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.application_sites (tenant_id, application_id, site_id)
			VALUES ($1, $2, $3)
		`, tenantID, a.ID, siteID); err != nil {
			return err
		}
	}
	for _, serviceID := range a.ServiceIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.application_services (tenant_id, application_id, service_id)
			VALUES ($1, $2, $3)
		`, tenantID, a.ID, serviceID); err != nil {
			return err
		}
	}
	for _, equipeID := range a.EquipeIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.application_equipes (tenant_id, application_id, equipe_id)
			VALUES ($1, $2, $3)
		`, tenantID, a.ID, equipeID); err != nil {
			return err
		}
	}
	return ensureHomeEquipeShares(ctx, tx, tenantID, a.ID)
}

// ensureHomeEquipeShares keeps equipes.application_id home links materialized in application_equipes.
func ensureHomeEquipeShares(ctx context.Context, tx pgx.Tx, tenantID, applicationID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO org.application_equipes (tenant_id, application_id, equipe_id)
		SELECT tenant_id, application_id, id
		FROM org.equipes
		WHERE application_id = $1 AND tenant_id = $2
		ON CONFLICT DO NOTHING
	`, applicationID, tenantID)
	return err
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *Repository) AssertApplicationSharesExist(ctx context.Context, tenant kernel.TenantID, siteIDs, serviceIDs, equipeIDs []uuid.UUID) error {
	if err := assertIDsInTenant(ctx, r.pool, `
		SELECT COUNT(*) FROM org.sites WHERE tenant_id = $1 AND id = ANY($2)
	`, tenant, siteIDs); err != nil {
		return err
	}
	if err := assertIDsInTenant(ctx, r.pool, `
		SELECT COUNT(*) FROM org.services WHERE tenant_id = $1 AND id = ANY($2)
	`, tenant, serviceIDs); err != nil {
		return err
	}
	return assertIDsInTenant(ctx, r.pool, `
		SELECT COUNT(*) FROM org.equipes WHERE tenant_id = $1 AND id = ANY($2)
	`, tenant, equipeIDs)
}

func assertIDsInTenant(ctx context.Context, db interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, query string, tenant kernel.TenantID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	var n int
	if err := db.QueryRow(ctx, query, tenant.UUID(), ids).Scan(&n); err != nil {
		return err
	}
	if n != len(ids) {
		return domain.ErrInvalidApplicationShare
	}
	return nil
}

func (r *Repository) ListApplications(ctx context.Context, tenant kernel.TenantID, filter ports.ApplicationListFilter) ([]domain.Application, error) {
	query := `
		SELECT id, tenant_id, libelle,
		       COALESCE(proprietaire, ''), COALESCE(mode_facturation, 'temps_passe'), COALESCE(uo_activee, FALSE),
		       chef_utilisateur_id, budget_defaut_id, active, COALESCE(default_tjm_cents, 0)
		FROM org.applications
		WHERE tenant_id = $1`
	args := []any{tenant.UUID()}
	if filter.Active != nil {
		query += ` AND active = $2`
		args = append(args, *filter.Active)
	}
	query += ` ORDER BY libelle`

	rows, err := r.pool.Query(ctx, query, args...)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.attachApplicationShares(ctx, tenant, out)
}

func (r *Repository) ListEquipes(ctx context.Context, tenant kernel.TenantID, filter ports.EquipeListFilter) ([]domain.Equipe, error) {
	query := `
		SELECT id, tenant_id, application_id, libelle, responsable_id
		FROM org.equipes
		WHERE tenant_id = $1`
	args := []any{tenant.UUID()}
	if filter.ApplicationID != nil {
		query += ` AND (
			application_id = $2
			OR EXISTS (
				SELECT 1 FROM org.application_equipes ae
				WHERE ae.equipe_id = org.equipes.id AND ae.application_id = $2
			)
		)`
		args = append(args, *filter.ApplicationID)
	}
	query += ` ORDER BY libelle`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Equipe
	for rows.Next() {
		var e domain.Equipe
		var tenantID uuid.UUID
		if err := rows.Scan(&e.ID, &tenantID, &e.ApplicationID, &e.Libelle, &e.ResponsableID); err != nil {
			return nil, err
		}
		e.TenantID = kernel.NewTenantID(tenantID)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) ListServices(ctx context.Context, tenant kernel.TenantID) ([]domain.ServiceSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.site_id, COALESCE(st.libelle, ''), COALESCE(st.societe_id, '00000000-0000-0000-0000-000000000000'::uuid),
		       COALESCE(s.libelle, ''), COALESCE(s.type, ''), s.responsable_id
		FROM org.services s
		LEFT JOIN org.sites st ON st.id = s.site_id
		WHERE s.tenant_id = $1
		ORDER BY st.libelle, s.libelle, s.id
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ServiceSummary
	for rows.Next() {
		var item domain.ServiceSummary
		if err := rows.Scan(&item.ID, &item.SiteID, &item.SiteLabel, &item.SocieteID,
			&item.Libelle, &item.Type, &item.ResponsableID); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *Repository) GetApplication(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Application, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, libelle,
		       COALESCE(proprietaire, ''), COALESCE(mode_facturation, 'temps_passe'), COALESCE(uo_activee, FALSE),
		       chef_utilisateur_id, budget_defaut_id, active, COALESCE(default_tjm_cents, 0)
		FROM org.applications
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id)
	app, err := scanApplication(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Application{}, domain.ErrApplicationNotFound
		}
		return domain.Application{}, err
	}
	apps, err := r.attachApplicationShares(ctx, tenant, []domain.Application{app})
	if err != nil {
		return domain.Application{}, err
	}
	return apps[0], nil
}

func scanApplication(row pgx.Row) (domain.Application, error) {
	var app domain.Application
	var tenantID uuid.UUID
	var proprietaire, modeFacturation string
	if err := row.Scan(
		&app.ID, &tenantID, &app.Libelle,
		&proprietaire, &modeFacturation, &app.UOActivee,
		&app.ChefUtilisateurID, &app.BudgetDefautID, &app.Active, &app.DefaultTJMCents,
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
		&app.ID, &tenantID, &app.Libelle,
		&proprietaire, &modeFacturation, &app.UOActivee,
		&app.ChefUtilisateurID, &app.BudgetDefautID, &app.Active, &app.DefaultTJMCents,
	); err != nil {
		return domain.Application{}, err
	}
	app.TenantID = kernel.NewTenantID(tenantID)
	app.Proprietaire = proprietaire
	app.ModeFacturation = modeFacturation
	return app, nil
}

func (r *Repository) attachApplicationShares(ctx context.Context, tenant kernel.TenantID, apps []domain.Application) ([]domain.Application, error) {
	if len(apps) == 0 {
		return apps, nil
	}
	ids := make([]uuid.UUID, len(apps))
	index := make(map[uuid.UUID]int, len(apps))
	for i, a := range apps {
		ids[i] = a.ID
		index[a.ID] = i
		apps[i].SiteIDs = nil
		apps[i].ServiceIDs = nil
		apps[i].EquipeIDs = nil
	}
	siteRows, err := r.pool.Query(ctx, `
		SELECT application_id, site_id FROM org.application_sites
		WHERE tenant_id = $1 AND application_id = ANY($2)
		ORDER BY site_id
	`, tenant.UUID(), ids)
	if err != nil {
		return nil, err
	}
	defer siteRows.Close()
	for siteRows.Next() {
		var appID, siteID uuid.UUID
		if err := siteRows.Scan(&appID, &siteID); err != nil {
			return nil, err
		}
		if i, ok := index[appID]; ok {
			apps[i].SiteIDs = append(apps[i].SiteIDs, siteID)
		}
	}
	if err := siteRows.Err(); err != nil {
		return nil, err
	}

	svcRows, err := r.pool.Query(ctx, `
		SELECT application_id, service_id FROM org.application_services
		WHERE tenant_id = $1 AND application_id = ANY($2)
		ORDER BY service_id
	`, tenant.UUID(), ids)
	if err != nil {
		return nil, err
	}
	defer svcRows.Close()
	for svcRows.Next() {
		var appID, serviceID uuid.UUID
		if err := svcRows.Scan(&appID, &serviceID); err != nil {
			return nil, err
		}
		if i, ok := index[appID]; ok {
			apps[i].ServiceIDs = append(apps[i].ServiceIDs, serviceID)
		}
	}
	if err := svcRows.Err(); err != nil {
		return nil, err
	}

	eqRows, err := r.pool.Query(ctx, `
		SELECT application_id, equipe_id FROM org.application_equipes
		WHERE tenant_id = $1 AND application_id = ANY($2)
		ORDER BY equipe_id
	`, tenant.UUID(), ids)
	if err != nil {
		return nil, err
	}
	defer eqRows.Close()
	for eqRows.Next() {
		var appID, equipeID uuid.UUID
		if err := eqRows.Scan(&appID, &equipeID); err != nil {
			return nil, err
		}
		if i, ok := index[appID]; ok {
			apps[i].EquipeIDs = append(apps[i].EquipeIDs, equipeID)
		}
	}
	if err := eqRows.Err(); err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *Repository) SaveUser(ctx context.Context, u domain.User) error {
	hydrateEquipeIDsFromPrimary(&u)
	u.SyncPrimaryMemberships()
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO org.users (
				id, tenant_id, equipe_id, login, prenom, nom, email, password_hash, profil,
				date_activation, date_expiration, active,
				totp_enabled, totp_enrollment_required
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`, u.ID, u.TenantID.UUID(), u.EquipeID, string(u.Login), u.Prenom, u.Nom, nullString(u.Email), u.PasswordHash, string(u.Profile),
			u.Period.Activation, u.Period.Expiration, u.Active, u.TotpEnabled, u.TotpEnrollmentRequired)
		if err != nil {
			return err
		}
		return replaceUserMemberships(ctx, tx, u)
	})
}

func (r *Repository) FindUserByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.User, error) {
	u, err := r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, tenant.UUID(), id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}
	if err := r.loadUserMemberships(ctx, &u); err != nil {
		return domain.User{}, err
	}
	return u, nil
}

// BudgetBelongsToApplication reports whether budgetID is the application default
// budget (same tenant + application + type 'defaut' — RG-BUD-01 source of truth).
func (r *Repository) BudgetBelongsToApplication(ctx context.Context, tenant kernel.TenantID, budgetID, applicationID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM budget.budgets
			WHERE tenant_id = $1 AND id = $2 AND application_id = $3 AND type = 'defaut'
		)
	`, tenant.UUID(), budgetID, applicationID).Scan(&exists)
	return exists, err
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
	profiles, equipeIDs, err := r.fetchMemberships(ctx, id)
	if err != nil {
		return ports.UserDetail{}, err
	}
	detail.Profiles = profiles
	if len(detail.Profiles) == 0 && profile != "" {
		detail.Profiles = []string{profile}
	}
	detail.EquipeIDs = equipeIDs
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
	// Callers must clear EquipeIDs (empty non-nil slice) to detach; SyncPrimaryMemberships
	// then sets EquipeID=nil. EquipeID set + empty EquipeIDs hydrates from the primary column.
	if len(u.EquipeIDs) == 0 && u.EquipeID != nil {
		hydrateEquipeIDsFromPrimary(&u)
	}
	u.SyncPrimaryMemberships()
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE org.users
			SET profil = $3, password_hash = $4, active = $5, email = $6, equipe_id = $7
			WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
		`, u.TenantID.UUID(), u.ID, string(u.Profile), u.PasswordHash, u.Active, nullString(u.Email), u.EquipeID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("user not found: %w", pgx.ErrNoRows)
		}
		return replaceUserMemberships(ctx, tx, u)
	})
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
	u, err := r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE tenant_id = $1 AND login = $2 AND deleted_at IS NULL
	`, tenant.UUID(), login))
	if err != nil {
		return domain.User{}, err
	}
	if err := r.loadUserMemberships(ctx, &u); err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (r *Repository) FindUserByLoginGlobal(ctx context.Context, login string) (domain.User, error) {
	u, err := r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE login = $1 AND deleted_at IS NULL LIMIT 1
	`, login))
	if err != nil {
		return domain.User{}, err
	}
	if err := r.loadUserMemberships(ctx, &u); err != nil {
		return domain.User{}, err
	}
	return u, nil
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.loadUsersMemberships(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
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
		SELECT DISTINCT u.login FROM org.users u
		WHERE u.tenant_id = $1 AND u.active = TRUE AND u.deleted_at IS NULL
		  AND (
		    u.equipe_id = $2
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      WHERE ue.user_id = u.id AND ue.equipe_id = $2
		    )
		  )
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
		SELECT DISTINCT u.login
		FROM org.users u
		WHERE u.tenant_id = $1 AND u.active = TRUE AND u.deleted_at IS NULL
		  AND (
		    EXISTS (
		      SELECT 1 FROM org.equipes e
		      WHERE e.id = u.equipe_id AND e.tenant_id = u.tenant_id AND e.application_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.equipes e ON e.id = ue.equipe_id AND e.tenant_id = u.tenant_id
		      WHERE ue.user_id = u.id AND e.application_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.application_equipes ae
		      WHERE ae.application_id = $2 AND ae.equipe_id = u.equipe_id
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.application_equipes ae ON ae.equipe_id = ue.equipe_id AND ae.application_id = $2
		      WHERE ue.user_id = u.id
		    )
		  )
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
		SELECT DISTINCT u.login
		FROM org.users u
		WHERE u.tenant_id = $1 AND u.active = TRUE AND u.deleted_at IS NULL
		  AND (
		    EXISTS (
		      SELECT 1 FROM org.equipes e
		      JOIN org.application_services asv ON asv.application_id = e.application_id
		      WHERE e.id = u.equipe_id AND e.tenant_id = u.tenant_id AND asv.service_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.equipes e ON e.id = ue.equipe_id AND e.tenant_id = u.tenant_id
		      JOIN org.application_services asv ON asv.application_id = e.application_id
		      WHERE ue.user_id = u.id AND asv.service_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.application_equipes ae
		      JOIN org.application_services asv ON asv.application_id = ae.application_id
		      WHERE ae.equipe_id = u.equipe_id AND asv.service_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.application_equipes ae ON ae.equipe_id = ue.equipe_id
		      JOIN org.application_services asv ON asv.application_id = ae.application_id
		      WHERE ue.user_id = u.id AND asv.service_id = $2
		    )
		  )
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

func (r *Repository) ResolveTenantUserEmails(ctx context.Context, tenant kernel.TenantID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT login FROM org.users
		WHERE tenant_id = $1 AND active = TRUE
	`, tenant.UUID())
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

func (r *Repository) ResolveEquipeUserIDs(ctx context.Context, tenant kernel.TenantID, equipeID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT u.id FROM org.users u
		WHERE u.tenant_id = $1 AND u.active = TRUE AND u.deleted_at IS NULL
		  AND (
		    u.equipe_id = $2
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      WHERE ue.user_id = u.id AND ue.equipe_id = $2
		    )
		  )
	`, tenant.UUID(), equipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserIDs(rows)
}

func (r *Repository) ResolveApplicationUserIDs(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT u.id
		FROM org.users u
		WHERE u.tenant_id = $1 AND u.active = TRUE AND u.deleted_at IS NULL
		  AND (
		    EXISTS (
		      SELECT 1 FROM org.equipes e
		      WHERE e.id = u.equipe_id AND e.tenant_id = u.tenant_id AND e.application_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.equipes e ON e.id = ue.equipe_id AND e.tenant_id = u.tenant_id
		      WHERE ue.user_id = u.id AND e.application_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.application_equipes ae
		      WHERE ae.application_id = $2 AND ae.equipe_id = u.equipe_id
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.application_equipes ae ON ae.equipe_id = ue.equipe_id AND ae.application_id = $2
		      WHERE ue.user_id = u.id
		    )
		  )
	`, tenant.UUID(), applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserIDs(rows)
}

func (r *Repository) ResolveServiceUserIDs(ctx context.Context, tenant kernel.TenantID, serviceID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT u.id
		FROM org.users u
		WHERE u.tenant_id = $1 AND u.active = TRUE AND u.deleted_at IS NULL
		  AND (
		    EXISTS (
		      SELECT 1 FROM org.equipes e
		      JOIN org.application_services asv ON asv.application_id = e.application_id
		      WHERE e.id = u.equipe_id AND e.tenant_id = u.tenant_id AND asv.service_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.equipes e ON e.id = ue.equipe_id AND e.tenant_id = u.tenant_id
		      JOIN org.application_services asv ON asv.application_id = e.application_id
		      WHERE ue.user_id = u.id AND asv.service_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.application_equipes ae
		      JOIN org.application_services asv ON asv.application_id = ae.application_id
		      WHERE ae.equipe_id = u.equipe_id AND asv.service_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.user_equipes ue
		      JOIN org.application_equipes ae ON ae.equipe_id = ue.equipe_id
		      JOIN org.application_services asv ON asv.application_id = ae.application_id
		      WHERE ue.user_id = u.id AND asv.service_id = $2
		    )
		  )
	`, tenant.UUID(), serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUserIDs(rows)
}

func scanUserIDs(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]uuid.UUID, error) {
	var out []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *Repository) ResolveSocieteIDForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (uuid.UUID, error) {
	var societeID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT st.societe_id
		FROM org.users u
		LEFT JOIN org.equipes e ON e.id = u.equipe_id
		LEFT JOIN org.application_services asv ON asv.application_id = e.application_id
		LEFT JOIN org.services sv ON sv.id = asv.service_id
		LEFT JOIN org.sites st ON st.id = sv.site_id
		WHERE u.tenant_id = $1 AND u.id = $2 AND st.societe_id IS NOT NULL
		LIMIT 1
	`, tenant.UUID(), userID).Scan(&societeID)
	if err == nil && societeID != uuid.Nil {
		return societeID, nil
	}
	// Via application_sites share on home app.
	err = r.pool.QueryRow(ctx, `
		SELECT st.societe_id
		FROM org.users u
		JOIN org.equipes e ON e.id = u.equipe_id
		JOIN org.application_sites asi ON asi.application_id = e.application_id
		JOIN org.sites st ON st.id = asi.site_id
		WHERE u.tenant_id = $1 AND u.id = $2
		LIMIT 1
	`, tenant.UUID(), userID).Scan(&societeID)
	if err == nil && societeID != uuid.Nil {
		return societeID, nil
	}
	// Fallback: any team membership via junction table (service share).
	err = r.pool.QueryRow(ctx, `
		SELECT st.societe_id
		FROM org.user_equipes ue
		JOIN org.equipes e ON e.id = ue.equipe_id
		JOIN org.application_services asv ON asv.application_id = e.application_id
		JOIN org.services sv ON sv.id = asv.service_id
		JOIN org.sites st ON st.id = sv.site_id
		WHERE ue.user_id = $1 AND e.tenant_id = $2
		ORDER BY ue.equipe_id
		LIMIT 1
	`, userID, tenant.UUID()).Scan(&societeID)
	if err == nil && societeID != uuid.Nil {
		return societeID, nil
	}
	// Fallback: team membership via site share on home app.
	err = r.pool.QueryRow(ctx, `
		SELECT st.societe_id
		FROM org.user_equipes ue
		JOIN org.equipes e ON e.id = ue.equipe_id
		JOIN org.application_sites asi ON asi.application_id = e.application_id
		JOIN org.sites st ON st.id = asi.site_id
		WHERE ue.user_id = $1 AND e.tenant_id = $2
		ORDER BY ue.equipe_id
		LIMIT 1
	`, userID, tenant.UUID()).Scan(&societeID)
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

func (r *Repository) ResolveSocieteIDForEquipe(ctx context.Context, tenant kernel.TenantID, equipeID uuid.UUID) (uuid.UUID, error) {
	var societeID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT st.societe_id
		FROM org.equipes e
		JOIN org.application_services asv ON asv.application_id = e.application_id
		JOIN org.services sv ON sv.id = asv.service_id
		JOIN org.sites st ON st.id = sv.site_id
		WHERE e.tenant_id = $1 AND e.id = $2
		LIMIT 1
	`, tenant.UUID(), equipeID).Scan(&societeID)
	if err == nil && societeID != uuid.Nil {
		return societeID, nil
	}
	err = r.pool.QueryRow(ctx, `
		SELECT st.societe_id
		FROM org.equipes e
		JOIN org.application_sites asi ON asi.application_id = e.application_id
		JOIN org.sites st ON st.id = asi.site_id
		WHERE e.tenant_id = $1 AND e.id = $2
		LIMIT 1
	`, tenant.UUID(), equipeID).Scan(&societeID)
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
	var totpSecret *string
	var totpEnabledAt *time.Time
	err := row.Scan(&u.ID, &tenantID, &u.EquipeID, &login, &u.Prenom, &u.Nom, &email, &u.PasswordHash, &profile,
		&u.Period.Activation, &expiration, &u.Active, &u.DeletedAt,
		&u.TotpEnabled, &u.TotpEnrollmentRequired, &totpSecret, &totpEnabledAt)
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
	if totpSecret != nil {
		u.TotpSecretEncrypted = *totpSecret
	}
	u.TotpEnabledAt = totpEnabledAt
	return u, nil
}

func hydrateEquipeIDsFromPrimary(u *domain.User) {
	if len(u.EquipeIDs) == 0 && u.EquipeID != nil {
		u.EquipeIDs = []uuid.UUID{*u.EquipeID}
	}
	if len(u.Profiles) == 0 && u.Profile != "" {
		u.Profiles = []domain.Profile{u.Profile}
	}
}

func replaceUserMemberships(ctx context.Context, tx pgx.Tx, u domain.User) error {
	if _, err := tx.Exec(ctx, `DELETE FROM org.user_profiles WHERE user_id = $1`, u.ID); err != nil {
		return err
	}
	for _, p := range u.Profiles {
		if p == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.user_profiles (user_id, profil) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, u.ID, string(p)); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM org.user_equipes WHERE user_id = $1`, u.ID); err != nil {
		return err
	}
	for _, eid := range u.EquipeIDs {
		if eid == uuid.Nil {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO org.user_equipes (user_id, equipe_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, u.ID, eid); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) fetchMemberships(ctx context.Context, userID uuid.UUID) ([]string, []uuid.UUID, error) {
	prows, err := r.pool.Query(ctx, `
		SELECT profil FROM org.user_profiles WHERE user_id = $1 ORDER BY profil
	`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer prows.Close()
	var profiles []string
	for prows.Next() {
		var p string
		if err := prows.Scan(&p); err != nil {
			return nil, nil, err
		}
		profiles = append(profiles, p)
	}
	if err := prows.Err(); err != nil {
		return nil, nil, err
	}

	erows, err := r.pool.Query(ctx, `
		SELECT equipe_id FROM org.user_equipes WHERE user_id = $1 ORDER BY equipe_id
	`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer erows.Close()
	var equipeIDs []uuid.UUID
	for erows.Next() {
		var id uuid.UUID
		if err := erows.Scan(&id); err != nil {
			return nil, nil, err
		}
		equipeIDs = append(equipeIDs, id)
	}
	return profiles, equipeIDs, erows.Err()
}

func (r *Repository) loadUserMemberships(ctx context.Context, u *domain.User) error {
	profiles, equipeIDs, err := r.fetchMemberships(ctx, u.ID)
	if err != nil {
		return err
	}
	u.Profiles = make([]domain.Profile, 0, len(profiles))
	for _, p := range profiles {
		u.Profiles = append(u.Profiles, domain.Profile(p))
	}
	if len(u.Profiles) == 0 && u.Profile != "" {
		u.Profiles = []domain.Profile{u.Profile}
	}
	u.EquipeIDs = orderEquipeIDsPrimaryFirst(u.EquipeID, equipeIDs)
	return nil
}

func (r *Repository) loadUsersMemberships(ctx context.Context, users []domain.User) error {
	if len(users) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(users))
	index := make(map[uuid.UUID]int, len(users))
	for i := range users {
		ids[i] = users[i].ID
		index[users[i].ID] = i
		users[i].Profiles = nil
		users[i].EquipeIDs = nil
	}

	prows, err := r.pool.Query(ctx, `
		SELECT user_id, profil FROM org.user_profiles
		WHERE user_id = ANY($1) ORDER BY user_id, profil
	`, ids)
	if err != nil {
		return err
	}
	defer prows.Close()
	for prows.Next() {
		var uid uuid.UUID
		var p string
		if err := prows.Scan(&uid, &p); err != nil {
			return err
		}
		if i, ok := index[uid]; ok {
			users[i].Profiles = append(users[i].Profiles, domain.Profile(p))
		}
	}
	if err := prows.Err(); err != nil {
		return err
	}

	erows, err := r.pool.Query(ctx, `
		SELECT user_id, equipe_id FROM org.user_equipes
		WHERE user_id = ANY($1) ORDER BY user_id, equipe_id
	`, ids)
	if err != nil {
		return err
	}
	defer erows.Close()
	for erows.Next() {
		var uid, eid uuid.UUID
		if err := erows.Scan(&uid, &eid); err != nil {
			return err
		}
		if i, ok := index[uid]; ok {
			users[i].EquipeIDs = append(users[i].EquipeIDs, eid)
		}
	}
	if err := erows.Err(); err != nil {
		return err
	}

	for i := range users {
		if len(users[i].Profiles) == 0 && users[i].Profile != "" {
			users[i].Profiles = []domain.Profile{users[i].Profile}
		}
		users[i].EquipeIDs = orderEquipeIDsPrimaryFirst(users[i].EquipeID, users[i].EquipeIDs)
	}
	return nil
}

func orderEquipeIDsPrimaryFirst(primary *uuid.UUID, ids []uuid.UUID) []uuid.UUID {
	if primary == nil {
		return ids
	}
	if len(ids) == 0 {
		return []uuid.UUID{*primary}
	}
	out := make([]uuid.UUID, 0, len(ids))
	out = append(out, *primary)
	for _, id := range ids {
		if id != *primary {
			out = append(out, id)
		}
	}
	return out
}

const userSelectCols = `id, tenant_id, equipe_id, login, prenom, nom, email, password_hash, profil, date_activation, date_expiration, active, deleted_at,
totp_enabled, totp_enrollment_required, totp_secret_encrypted, totp_enabled_at`

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
	u, err := r.scanUser(r.pool.QueryRow(ctx, `
		SELECT `+userSelectCols+`
		FROM org.users WHERE tenant_id = $1 AND lower(email) = lower($2) AND deleted_at IS NULL
	`, tenant.UUID(), email))
	if err != nil {
		return domain.User{}, err
	}
	if err := r.loadUserMemberships(ctx, &u); err != nil {
		return domain.User{}, err
	}
	return u, nil
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

func decodeMailRecipients(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func encodeMailRecipients(recipients []string) []byte {
	if recipients == nil {
		recipients = []string{}
	}
	data, _ := json.Marshal(recipients)
	return data
}

func decodeTaskTypes(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func encodeTaskTypes(types []string) []byte {
	if types == nil {
		types = []string{}
	}
	data, _ := json.Marshal(types)
	return data
}

func (r *Repository) UpdateUserTotp(ctx context.Context, u domain.User) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.users
		SET totp_enabled = $3,
		    totp_enrollment_required = $4,
		    totp_secret_encrypted = $5,
		    totp_enabled_at = $6
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`, u.TenantID.UUID(), u.ID, u.TotpEnabled, u.TotpEnrollmentRequired, nullString(u.TotpSecretEncrypted), u.TotpEnabledAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %w", pgx.ErrNoRows)
	}
	return nil
}

func (r *Repository) SaveTotpBackupCodes(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, codeHashes []string) error {
	if err := r.DeleteTotpBackupCodes(ctx, tenant, userID); err != nil {
		return err
	}
	for _, hash := range codeHashes {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO org.user_totp_backup_codes (id, tenant_id, user_id, code_hash)
			VALUES ($1, $2, $3, $4)
		`, uuid.New(), tenant.UUID(), userID, hash)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ConsumeTotpBackupCode(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, codeHash string, usedAt time.Time) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.user_totp_backup_codes
		SET used_at = $4
		WHERE tenant_id = $1 AND user_id = $2 AND code_hash = $3 AND used_at IS NULL
	`, tenant.UUID(), userID, codeHash, usedAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) DeleteTotpBackupCodes(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM org.user_totp_backup_codes
		WHERE tenant_id = $1 AND user_id = $2
	`, tenant.UUID(), userID)
	return err
}

func (r *Repository) ListUnusedTotpBackupCodeHashes(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code_hash FROM org.user_totp_backup_codes
		WHERE tenant_id = $1 AND user_id = $2 AND used_at IS NULL
	`, tenant.UUID(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, err
		}
		out = append(out, hash)
	}
	return out, rows.Err()
}

func (r *Repository) MarkTotpEnrollmentRequiredForSocieteUsers(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) (int, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE org.users u
		SET totp_enrollment_required = TRUE
		WHERE u.tenant_id = $1
		  AND u.active = TRUE
		  AND u.deleted_at IS NULL
		  AND COALESCE(u.totp_enabled, FALSE) = FALSE
		  AND (
		    EXISTS (
		      SELECT 1 FROM org.equipes e
		      JOIN org.application_services asv ON asv.application_id = e.application_id
		      JOIN org.services sv ON sv.id = asv.service_id
		      JOIN org.sites st ON st.id = sv.site_id
		      WHERE e.id = u.equipe_id AND st.societe_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.equipes e
		      JOIN org.application_sites asi ON asi.application_id = e.application_id
		      JOIN org.sites st ON st.id = asi.site_id
		      WHERE e.id = u.equipe_id AND st.societe_id = $2
		    )
		    OR (
		      u.equipe_id IS NULL
		      AND (SELECT id FROM org.societes WHERE tenant_id = $1 ORDER BY raison_sociale LIMIT 1) = $2
		    )
		  )
	`, tenant.UUID(), societeID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *Repository) ClearTotpEnrollmentRequiredForSocieteUsers(ctx context.Context, tenant kernel.TenantID, societeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE org.users u
		SET totp_enrollment_required = FALSE
		WHERE u.tenant_id = $1
		  AND COALESCE(u.totp_enabled, FALSE) = FALSE
		  AND (
		    EXISTS (
		      SELECT 1 FROM org.equipes e
		      JOIN org.application_services asv ON asv.application_id = e.application_id
		      JOIN org.services sv ON sv.id = asv.service_id
		      JOIN org.sites st ON st.id = sv.site_id
		      WHERE e.id = u.equipe_id AND st.societe_id = $2
		    )
		    OR EXISTS (
		      SELECT 1 FROM org.equipes e
		      JOIN org.application_sites asi ON asi.application_id = e.application_id
		      JOIN org.sites st ON st.id = asi.site_id
		      WHERE e.id = u.equipe_id AND st.societe_id = $2
		    )
		    OR (
		      u.equipe_id IS NULL
		      AND (SELECT id FROM org.societes WHERE tenant_id = $1 ORDER BY raison_sociale LIMIT 1) = $2
		    )
		  )
	`, tenant.UUID(), societeID)
	return err
}

var _ ports.OrganizationRepository = (*Repository)(nil)
var _ ports.RequestSettingsRepository = (*Repository)(nil)
