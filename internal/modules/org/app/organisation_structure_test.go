package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cryptox"
	"github.com/kore/kore/pkg/kernel"
)

// structureRepo réutilise le fake complet de refresh_test.go et n'intercepte que
// les méthodes utiles à la hiérarchie organisation.
type structureRepo struct {
	refreshUserRepo
	savedEquipe   *domain.Equipe
	savedService  *domain.Service
	updatedUser   *domain.User
	sites         []domain.SiteSummary
	saveEquipeErr error
}

func (r *structureRepo) SaveEquipe(_ context.Context, e domain.Equipe) error {
	if r.saveEquipeErr != nil {
		return r.saveEquipeErr
	}
	r.savedEquipe = &e
	return nil
}

func (r *structureRepo) SaveService(_ context.Context, s domain.Service) error {
	r.savedService = &s
	return nil
}

func (r *structureRepo) ListSites(context.Context, kernel.TenantID) ([]domain.SiteSummary, error) {
	return r.sites, nil
}

func (r *structureRepo) UpdateUser(_ context.Context, u domain.User) error {
	r.updatedUser = &u
	return nil
}

func TestCreateEquipe_persistsWithApplication(t *testing.T) {
	repo := &structureRepo{}
	svc := NewOrganizationService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	responsable := uuid.New()

	equipe, err := svc.CreateEquipe(context.Background(), ports.CreateEquipeCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		Libelle:       "Équipe Dev",
		ResponsableID: &responsable,
	})
	if err != nil {
		t.Fatalf("CreateEquipe: %v", err)
	}
	if equipe.ID == uuid.Nil {
		t.Fatal("expected a generated equipe id")
	}
	if repo.savedEquipe == nil {
		t.Fatal("expected SaveEquipe to be called")
	}
	if repo.savedEquipe.ApplicationID != appID {
		t.Fatalf("applicationID = %v, want %v", repo.savedEquipe.ApplicationID, appID)
	}
	if repo.savedEquipe.Libelle != "Équipe Dev" {
		t.Fatalf("libelle = %q", repo.savedEquipe.Libelle)
	}
	if repo.savedEquipe.ResponsableID == nil || *repo.savedEquipe.ResponsableID != responsable {
		t.Fatalf("responsableID = %v, want %v", repo.savedEquipe.ResponsableID, responsable)
	}
}

func TestCreateEquipe_rejectsMissingApplication(t *testing.T) {
	repo := &structureRepo{}
	svc := NewOrganizationService(repo)

	_, err := svc.CreateEquipe(context.Background(), ports.CreateEquipeCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		Libelle:  "Équipe orpheline",
	})
	if !errors.Is(err, domain.ErrEquipeWithoutApplication) {
		t.Fatalf("err = %v, want ErrEquipeWithoutApplication", err)
	}
	if repo.savedEquipe != nil {
		t.Fatal("expected no persistence when the application is missing")
	}
}

func TestCreateService_defaultsTypeToInterne(t *testing.T) {
	repo := &structureRepo{}
	svc := NewOrganizationService(repo)

	_, err := svc.CreateService(context.Background(), ports.CreateServiceCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		SiteID:        uuid.New(),
		Libelle:       "Delivery",
		ResponsableID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("CreateService: %v", err)
	}
	if repo.savedService == nil {
		t.Fatal("expected SaveService to be called")
	}
	if repo.savedService.Type != domain.DefaultServiceType {
		t.Fatalf("type = %q, want %q", repo.savedService.Type, domain.DefaultServiceType)
	}
	if repo.savedService.Libelle != "Delivery" {
		t.Fatalf("libelle = %q", repo.savedService.Libelle)
	}
}

func TestCreateService_rejectsMissingResponsible(t *testing.T) {
	repo := &structureRepo{}
	svc := NewOrganizationService(repo)

	_, err := svc.CreateService(context.Background(), ports.CreateServiceCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		SiteID:   uuid.New(),
		Libelle:  "Sans responsable",
	})
	if !errors.Is(err, domain.ErrServiceWithoutResponsible) {
		t.Fatalf("err = %v, want ErrServiceWithoutResponsible", err)
	}
	if repo.savedService != nil {
		t.Fatal("expected no persistence without a responsible")
	}
}

func TestListSites_delegatesToRepository(t *testing.T) {
	want := []domain.SiteSummary{{ID: uuid.New(), Libelle: "Paris HQ"}}
	svc := NewOrganizationService(&structureRepo{sites: want})

	got, err := svc.ListSites(context.Background(), kernel.NewTenantID(uuid.New()))
	if err != nil {
		t.Fatalf("ListSites: %v", err)
	}
	if len(got) != 1 || got[0].Libelle != "Paris HQ" {
		t.Fatalf("sites = %+v", got)
	}
}

func newUserServiceForEquipe(t *testing.T, repo ports.OrganizationRepository) ports.UserService {
	t.Helper()
	return NewUserService(
		repo,
		NewArgon2Hasher(),
		authx.NewTokenIssuer("test-signing-key", time.Hour, time.Hour),
		nil,
		nil,
		nil,
		nil,
		cryptox.DevKeyFromJWTSigningKey("test-signing-key"),
	)
}

func equipeTestUser(tenant kernel.TenantID, equipeID *uuid.UUID) domain.User {
	return domain.User{
		ID:       uuid.New(),
		TenantID: tenant,
		EquipeID: equipeID,
		Login:    "COL_collab",
		Profile:  domain.ProfileCollaborateur,
		Active:   true,
		Period:   domain.ActivationPeriod{Activation: time.Now().UTC().Add(-time.Hour)},
	}
}

func TestUpdateUser_attachesEquipe(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	user := equipeTestUser(tenant, nil)
	repo := &structureRepo{refreshUserRepo: refreshUserRepo{user: user}}
	svc := newUserServiceForEquipe(t, repo)

	equipeID := uuid.New()
	target := &equipeID
	summary, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID: tenant,
		UserID:   user.ID,
		EquipeID: &target,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if repo.updatedUser == nil {
		t.Fatal("expected UpdateUser to be persisted")
	}
	if repo.updatedUser.EquipeID == nil || *repo.updatedUser.EquipeID != equipeID {
		t.Fatalf("persisted equipeID = %v, want %v", repo.updatedUser.EquipeID, equipeID)
	}
	if summary.EquipeID == nil || *summary.EquipeID != equipeID {
		t.Fatalf("summary equipeID = %v, want %v", summary.EquipeID, equipeID)
	}
}

func TestUpdateUser_detachesEquipe(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	current := uuid.New()
	user := equipeTestUser(tenant, &current)
	repo := &structureRepo{refreshUserRepo: refreshUserRepo{user: user}}
	svc := newUserServiceForEquipe(t, repo)

	// Pointeur vers nil = détachement explicite.
	var none *uuid.UUID
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID: tenant,
		UserID:   user.ID,
		EquipeID: &none,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if repo.updatedUser == nil {
		t.Fatal("expected UpdateUser to be persisted")
	}
	if repo.updatedUser.EquipeID != nil {
		t.Fatalf("expected equipe detached, got %v", repo.updatedUser.EquipeID)
	}
}

func TestUpdateUser_keepsEquipeWhenFieldAbsent(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	current := uuid.New()
	user := equipeTestUser(tenant, &current)
	repo := &structureRepo{refreshUserRepo: refreshUserRepo{user: user}}
	svc := newUserServiceForEquipe(t, repo)

	profile := domain.ProfileAdmin
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID: tenant,
		UserID:   user.ID,
		Profile:  &profile,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if repo.updatedUser == nil {
		t.Fatal("expected UpdateUser to be persisted")
	}
	if repo.updatedUser.EquipeID == nil || *repo.updatedUser.EquipeID != current {
		t.Fatalf("expected equipe unchanged (%v), got %v", current, repo.updatedUser.EquipeID)
	}
}
