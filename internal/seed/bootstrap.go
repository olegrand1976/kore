package seed

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

var ErrSeedProtected = errors.New("seed-reset refusé : organisation protégée (seed_protected)")

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// shouldRestoreDemoSociete reports whether the demo société was overwritten by bootstrap v1
// (renamed to the LL-IT label on the demo tenant).
func shouldRestoreDemoSociete(raisonSociale string) bool {
	return raisonSociale == LLITSocieteName
}

// BootstrapLLIT crée un tenant dédié LL-IT (isolé du démo), une société seed_protected
// et le compte ADM_olivier. Idempotent. Nettoie un éventuel ADM_olivier créé par erreur
// sur le tenant démo (bootstrap v1).
func (r *Runner) BootstrapLLIT(ctx context.Context) error {
	llit := kernel.NewTenantID(LLITTenantID)
	demo := kernel.NewTenantID(DemoTenantID)

	if err := r.deps.OrgRepo.SaveTenant(ctx, orgdomain.Tenant{ID: LLITTenantID, Name: LLITTenantName}); err != nil {
		return err
	}
	// Restore demo tenant label if a previous bootstrap overwrote it.
	if err := r.deps.OrgRepo.SaveTenant(ctx, orgdomain.Tenant{ID: DemoTenantID, Name: TenantName}); err != nil {
		return err
	}
	if err := r.ensureTrial(ctx, llit); err != nil {
		return err
	}
	if err := r.ensureWorkflows(ctx, llit); err != nil {
		return err
	}
	if err := r.ensureLLITSociete(ctx, llit); err != nil {
		return err
	}
	if r.deps.LeaveTypes != nil {
		if err := r.deps.LeaveTypes.BootstrapDefaults(ctx, llit, LLITSocieteID); err != nil {
			return err
		}
	}
	if err := r.cleanupMisplacedProdAdmin(ctx, demo); err != nil {
		return err
	}
	if err := r.restoreDemoSocieteAfterMisbootstrap(ctx); err != nil {
		return err
	}
	return r.ensureProdAdmin(ctx, llit)
}

func (r *Runner) ensureLLITSociete(ctx context.Context, tenant kernel.TenantID) error {
	existing, err := r.deps.OrgRepo.GetSociete(ctx, tenant, LLITSocieteID)
	if err == nil && existing.ID != uuid.Nil {
		existing.RaisonSociale = LLITSocieteName
		existing.SeedProtected = true
		if err := r.deps.OrgRepo.UpdateSociete(ctx, existing); err != nil {
			return err
		}
		log.Printf("bootstrap-llit: société protégée mise à jour sur tenant LL-IT (%s)", LLITSocieteName)
		return nil
	}
	if err != nil && !isNotFound(err) {
		return err
	}

	if err := r.deps.OrgRepo.SaveSociete(ctx, orgdomain.Societe{
		ID:            LLITSocieteID,
		TenantID:      tenant,
		RaisonSociale: LLITSocieteName,
		Devise:        "EUR",
		Pays:          "BE",
		SeedProtected: true,
	}); err != nil {
		return err
	}
	log.Printf("bootstrap-llit: société créée sur tenant LL-IT (%s)", LLITSocieteName)
	return nil
}

func (r *Runner) cleanupMisplacedProdAdmin(ctx context.Context, demoTenant kernel.TenantID) error {
	user, err := r.deps.OrgRepo.FindUserByLogin(ctx, demoTenant, ProdAdminLogin)
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := r.deps.OrgRepo.SoftDeleteUser(ctx, demoTenant, user.ID, time.Now().UTC()); err != nil {
		return fmt.Errorf("cleanup ADM_olivier on demo tenant: %w", err)
	}
	log.Printf("bootstrap-llit: %s retiré du tenant démo (isolation)", ProdAdminLogin)
	return nil
}

func (r *Runner) restoreDemoSocieteAfterMisbootstrap(ctx context.Context) error {
	demo := kernel.NewTenantID(DemoTenantID)
	societe, err := r.deps.OrgRepo.GetSociete(ctx, demo, DemoSocieteID)
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !shouldRestoreDemoSociete(societe.RaisonSociale) {
		return nil
	}
	societe.RaisonSociale = DemoSocieteName
	societe.SeedProtected = false
	if err := r.deps.OrgRepo.UpdateSociete(ctx, societe); err != nil {
		return err
	}
	log.Printf("bootstrap-llit: société démo restaurée (%s, seed_protected=false)", DemoSocieteName)
	return nil
}

func (r *Runner) ensureProdAdmin(ctx context.Context, tenant kernel.TenantID) error {
	_, err := r.deps.OrgRepo.FindUserByLogin(ctx, tenant, ProdAdminLogin)
	if err == nil {
		log.Printf("bootstrap-llit: compte %s déjà présent sur tenant LL-IT (mot de passe inchangé)", ProdAdminLogin)
		return nil
	}
	if !isNotFound(err) {
		return err
	}

	password := strings.TrimSpace(os.Getenv("KORE_PROD_ADMIN_PASSWORD"))
	generated := false
	if password == "" {
		password, err = generateStrongPassword()
		if err != nil {
			return err
		}
		generated = true
	}

	if _, err := r.deps.Users.CreateUser(ctx, orgports.CreateUserCommand{
		TenantID: tenant,
		Login:    ProdAdminLogin,
		Password: password,
		Profile:  orgdomain.ProfileAdmin,
	}); err != nil {
		return err
	}

	if generated {
		log.Printf("bootstrap-llit: compte %s créé sur tenant LL-IT — mot de passe (une seule fois) : %s", ProdAdminLogin, password)
		log.Println("bootstrap-llit: enregistrez ce mot de passe puis changez-le après connexion")
	} else {
		log.Printf("bootstrap-llit: compte %s créé sur tenant LL-IT (mot de passe depuis KORE_PROD_ADMIN_PASSWORD)", ProdAdminLogin)
	}
	return nil
}

func (r *Runner) hasSeedProtectedSociete(ctx context.Context, tenant kernel.TenantID) (bool, error) {
	var exists bool
	err := r.deps.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM org.societes
			WHERE tenant_id = $1 AND seed_protected = TRUE
		)
	`, tenant.UUID()).Scan(&exists)
	return exists, err
}

func generateStrongPassword() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
