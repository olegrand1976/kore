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

	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

var ErrSeedProtected = errors.New("seed-reset refusé : organisation protégée (seed_protected)")

// BootstrapLLIT ensures the LL-IT société (seed-protected) and ADM_olivier admin exist.
// Password comes from KORE_PROD_ADMIN_PASSWORD, or is generated once when creating the user.
func (r *Runner) BootstrapLLIT(ctx context.Context) error {
	tenant := kernel.NewTenantID(DemoTenantID)
	if err := r.ensureTenant(ctx); err != nil {
		return err
	}
	if err := r.ensureLLITSociete(ctx, tenant); err != nil {
		return err
	}
	return r.ensureProdAdmin(ctx, tenant)
}

func (r *Runner) ensureLLITSociete(ctx context.Context, tenant kernel.TenantID) error {
	existing, err := r.deps.OrgRepo.GetSociete(ctx, tenant, DemoSocieteID)
	if err == nil && existing.ID != uuid.Nil {
		existing.RaisonSociale = DemoSocieteName
		existing.SeedProtected = true
		if err := r.deps.OrgRepo.UpdateSociete(ctx, existing); err != nil {
			return err
		}
		log.Printf("bootstrap-llit: société protégée mise à jour (%s)", DemoSocieteName)
		return nil
	}

	if err := r.deps.OrgRepo.SaveSociete(ctx, orgdomain.Societe{
		ID:            DemoSocieteID,
		TenantID:      tenant,
		RaisonSociale: DemoSocieteName,
		Devise:        "EUR",
		Pays:          "FR",
		SeedProtected: true,
	}); err != nil {
		return err
	}
	log.Printf("bootstrap-llit: société créée et protégée (%s)", DemoSocieteName)
	return nil
}

func (r *Runner) ensureProdAdmin(ctx context.Context, tenant kernel.TenantID) error {
	exists, err := r.deps.OrgRepo.ExistsLogin(ctx, tenant, ProdAdminLogin)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("bootstrap-llit: compte %s déjà présent (mot de passe inchangé)", ProdAdminLogin)
		return nil
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
		log.Printf("bootstrap-llit: compte %s créé — mot de passe (affiché une seule fois) : %s", ProdAdminLogin, password)
		log.Println("bootstrap-llit: enregistrez ce mot de passe puis changez-le après connexion")
	} else {
		log.Printf("bootstrap-llit: compte %s créé (mot de passe depuis KORE_PROD_ADMIN_PASSWORD)", ProdAdminLogin)
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
