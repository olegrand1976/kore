package seed

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

const (
	devTenantName = "Kore Demo"
	devAdminLogin = "ADM_admin"
	devAdminEmail = "lalouviere.it.sc@gmail.com"
	devAdminPass  = "Admin123!"
)

var devTenantID = uuid.MustParse("00000000-0000-4000-8000-000000000001")

type Seeder struct {
	repo    ports.OrganizationRepository
	users   ports.UserService
	enabled bool
}

func NewSeeder(repo ports.OrganizationRepository, users ports.UserService, enabled bool) *Seeder {
	return &Seeder{repo: repo, users: users, enabled: enabled}
}

func (s *Seeder) Run(ctx context.Context) error {
	if !s.enabled {
		return nil
	}
	tenant := domain.Tenant{ID: devTenantID, Name: devTenantName}
	if err := s.repo.SaveTenant(ctx, tenant); err != nil {
		return err
	}
	tenantID := kernel.NewTenantID(devTenantID)
	societes, err := s.repo.ListSocietes(ctx, tenantID)
	if err != nil {
		return err
	}
	if len(societes) == 0 {
		err = s.repo.SaveSociete(ctx, domain.Societe{
			ID:            uuid.New(),
			TenantID:      tenantID,
			RaisonSociale: "Kore Demo SAS",
			Devise:        "EUR",
			Adresse:       "1 rue de la Démo, 75001 Paris",
			Siret:         "12345678901234",
			URLTenant:     "demo.kore.local",
		})
		if err != nil {
			return err
		}
		log.Println("dev seed: demo societe created")
	}
	exists, err := s.repo.ExistsLogin(ctx, tenantID, devAdminLogin)
	if err != nil {
		return err
	}
	if exists {
		if err := s.ensureAdminEmail(ctx, tenantID); err != nil {
			return err
		}
		log.Println("dev seed: admin already exists, email ensured")
		return nil
	}
	_, err = s.users.CreateUser(ctx, ports.CreateUserCommand{
		TenantID: tenantID,
		Login:    devAdminLogin,
		Password: devAdminPass,
		Profile:  domain.ProfileAdmin,
	})
	if err != nil {
		return err
	}
	if err := s.ensureAdminEmail(ctx, tenantID); err != nil {
		return err
	}
	log.Printf("dev seed: tenant %s admin %s created", devTenantID, devAdminLogin)
	return nil
}

func (s *Seeder) ensureAdminEmail(ctx context.Context, tenantID kernel.TenantID) error {
	user, err := s.repo.FindUserByLogin(ctx, tenantID, devAdminLogin)
	if err != nil {
		return err
	}
	user.Email = devAdminEmail
	return s.repo.UpdateUser(ctx, user)
}
