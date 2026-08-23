package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
)

func (r *Repository) MergeApplications(
	ctx context.Context,
	tenant kernel.TenantID,
	absorbedApplicationID,
	referenceApplicationID uuid.UUID,
) (domain.Application, error) {
	if absorbedApplicationID == referenceApplicationID {
		return domain.Application{}, domain.ErrApplicationsMergeInvalid
	}

	tenantUUID := tenant.UUID()
	err := r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		absorbed, err := getApplicationTx(ctx, tx, tenantUUID, absorbedApplicationID)
		if err != nil {
			if errors.Is(err, domain.ErrApplicationNotFound) {
				return domain.ErrApplicationsMergeInvalid
			}
			return err
		}
		reference, err := getApplicationTx(ctx, tx, tenantUUID, referenceApplicationID)
		if err != nil {
			if errors.Is(err, domain.ErrApplicationNotFound) {
				return domain.ErrApplicationsMergeInvalid
			}
			return err
		}

		if err := assertNoActiveSprintConflict(ctx, tx, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}

		if err := assertNoDuplicateDefaultBudgets(ctx, tx, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}

		if err := mergeApplicationShareRows(ctx, tx, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE org.equipes
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if err := ensureHomeEquipeShares(ctx, tx, tenantUUID, referenceApplicationID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE budget.budgets
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE tma.demands
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE tma.releases
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE maintenance.work_requests
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE support.tickets
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM ssii.mission_applications
			WHERE tenant_id = $1 AND application_id = $2
			  AND mission_id IN (
			    SELECT mission_id FROM ssii.mission_applications
			    WHERE tenant_id = $4 AND application_id = $3
			  )
		`, tenantUUID, absorbedApplicationID, referenceApplicationID, tenantUUID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE ssii.mission_applications
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE project.epics
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE project.sprints
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}
		if err := mergeKanbanConfigs(ctx, tx, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE cra.time_lines
			SET source_id = $3::text
			WHERE tenant_id = $1 AND source_type = 'application' AND source_id = $2::text
		`, tenantUUID, absorbedApplicationID, referenceApplicationID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			DELETE FROM integrations.external_links
			WHERE tenant_id = $1 AND kore_entity_type = 'application' AND kore_entity_id = $2
		`, tenantUUID, absorbedApplicationID); err != nil {
			return err
		}

		proprietaire := reference.Proprietaire
		if proprietaire == "" && absorbed.Proprietaire != "" {
			proprietaire = absorbed.Proprietaire
		}
		chefID := reference.ChefUtilisateurID
		if chefID == nil && absorbed.ChefUtilisateurID != nil {
			chefID = absorbed.ChefUtilisateurID
		}
		budgetDefautID := reference.BudgetDefautID
		if budgetDefautID == nil && absorbed.BudgetDefautID != nil {
			budgetDefautID = absorbed.BudgetDefautID
		}

		tag, err := tx.Exec(ctx, `
			UPDATE org.applications
			SET proprietaire = $3,
			    chef_utilisateur_id = $4,
			    budget_defaut_id = $5
			WHERE tenant_id = $1 AND id = $2
		`, tenantUUID, referenceApplicationID, nullIfEmpty(proprietaire), chefID, budgetDefautID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrApplicationNotFound
		}

		tag, err = tx.Exec(ctx, `
			UPDATE org.applications
			SET active = FALSE
			WHERE tenant_id = $1 AND id = $2
		`, tenantUUID, absorbedApplicationID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrApplicationNotFound
		}

		if err := deleteApplicationShareRows(ctx, tx, tenantUUID, absorbedApplicationID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return domain.Application{}, err
	}
	return r.GetApplication(ctx, tenant, referenceApplicationID)
}

func getApplicationTx(ctx context.Context, tx pgx.Tx, tenantID, applicationID uuid.UUID) (domain.Application, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, tenant_id, libelle,
		       COALESCE(proprietaire, ''), COALESCE(mode_facturation, 'temps_passe'), COALESCE(uo_activee, FALSE),
		       chef_utilisateur_id, budget_defaut_id, active, COALESCE(default_tjm_cents, 0),
		       COALESCE(methodology_profile, 'psa')
		FROM org.applications
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, applicationID)
	app, err := scanApplication(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Application{}, domain.ErrApplicationNotFound
		}
		return domain.Application{}, err
	}
	return app, nil
}

func assertNoActiveSprintConflict(ctx context.Context, tx pgx.Tx, tenantID, absorbedID, referenceID uuid.UUID) error {
	var count int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT 1 FROM project.sprints
			WHERE tenant_id = $1 AND application_id = $2 AND status = 'active'
			UNION ALL
			SELECT 1 FROM project.sprints
			WHERE tenant_id = $1 AND application_id = $3 AND status = 'active'
		) active_sprints
	`, tenantID, absorbedID, referenceID).Scan(&count)
	if err != nil {
		return err
	}
	if count >= 2 {
		return domain.ErrApplicationsMergeActiveSprintConflict
	}
	return nil
}

func mergeApplicationShareRows(ctx context.Context, tx pgx.Tx, tenantID, absorbedID, referenceID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO org.application_sites (tenant_id, application_id, site_id)
		SELECT tenant_id, $3, site_id
		FROM org.application_sites
		WHERE tenant_id = $1 AND application_id = $2
		ON CONFLICT DO NOTHING
	`, tenantID, absorbedID, referenceID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO org.application_services (tenant_id, application_id, service_id)
		SELECT tenant_id, $3, service_id
		FROM org.application_services
		WHERE tenant_id = $1 AND application_id = $2
		ON CONFLICT DO NOTHING
	`, tenantID, absorbedID, referenceID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO org.application_equipes (tenant_id, application_id, equipe_id)
		SELECT tenant_id, $3, equipe_id
		FROM org.application_equipes
		WHERE tenant_id = $1 AND application_id = $2
		ON CONFLICT DO NOTHING
	`, tenantID, absorbedID, referenceID); err != nil {
		return err
	}
	return nil
}

func assertNoDuplicateDefaultBudgets(ctx context.Context, tx pgx.Tx, tenantID, absorbedID, referenceID uuid.UUID) error {
	refCount, err := countDefaultBudgetsTx(ctx, tx, tenantID, referenceID)
	if err != nil {
		return err
	}
	absCount, err := countDefaultBudgetsTx(ctx, tx, tenantID, absorbedID)
	if err != nil {
		return err
	}
	if refCount > 0 && absCount > 0 {
		return domain.ErrApplicationsMergeDuplicateDefaultBudget
	}
	return nil
}

func countDefaultBudgetsTx(ctx context.Context, tx pgx.Tx, tenantID, applicationID uuid.UUID) (int, error) {
	var n int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM budget.budgets
		WHERE tenant_id = $1 AND application_id = $2
		  AND type IN ('defaut', 'default')
	`, tenantID, applicationID).Scan(&n)
	return n, err
}

func deleteApplicationShareRows(ctx context.Context, tx pgx.Tx, tenantID, applicationID uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM org.application_sites WHERE tenant_id = $1 AND application_id = $2`, tenantID, applicationID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM org.application_services WHERE tenant_id = $1 AND application_id = $2`, tenantID, applicationID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM org.application_equipes WHERE tenant_id = $1 AND application_id = $2`, tenantID, applicationID); err != nil {
		return err
	}
	return nil
}

func mergeKanbanConfigs(ctx context.Context, tx pgx.Tx, tenantID, absorbedID, referenceID uuid.UUID) error {
	var referenceHas, absorbedHas bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM project.kanban_configs
			WHERE tenant_id = $1 AND application_id = $2
		)
	`, tenantID, referenceID).Scan(&referenceHas)
	if err != nil {
		return err
	}
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM project.kanban_configs
			WHERE tenant_id = $1 AND application_id = $2
		)
	`, tenantID, absorbedID).Scan(&absorbedHas)
	if err != nil {
		return err
	}
	if !absorbedHas {
		return nil
	}
	if !referenceHas {
		_, err = tx.Exec(ctx, `
			UPDATE project.kanban_configs
			SET application_id = $3
			WHERE tenant_id = $1 AND application_id = $2
		`, tenantID, absorbedID, referenceID)
		return err
	}
	_, err = tx.Exec(ctx, `
		DELETE FROM project.kanban_configs
		WHERE tenant_id = $1 AND application_id = $2
	`, tenantID, absorbedID)
	return err
}
