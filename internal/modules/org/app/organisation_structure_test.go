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
	savedEquipe      *domain.Equipe
	savedService     *domain.Service
	savedApplication *domain.Application
	applications     map[uuid.UUID]domain.Application
	updatedUser      *domain.User
	sites            []domain.SiteSummary
	equipes          []domain.Equipe
	listedUsers      []domain.User
	saveEquipeErr    error
	users            map[uuid.UUID]domain.User
	budgetOK         map[uuid.UUID]bool
}

func (r *structureRepo) BudgetBelongsToApplication(_ context.Context, _ kernel.TenantID, budgetID, _ uuid.UUID) (bool, error) {
	if r.budgetOK == nil {
		return false, nil
	}
	return r.budgetOK[budgetID], nil
}

func (r *structureRepo) FindUserByID(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.User, error) {
	if r.users != nil {
		if u, ok := r.users[id]; ok {
			return u, nil
		}
		return domain.User{}, domain.ErrUserNotFound
	}
	return r.refreshUserRepo.FindUserByID(context.Background(), kernel.TenantID{}, id)
}

func (r *structureRepo) SaveEquipe(_ context.Context, e domain.Equipe) error {
	if r.saveEquipeErr != nil {
		return r.saveEquipeErr
	}
	r.savedEquipe = &e
	r.equipes = append(r.equipes, e)
	return nil
}

func (r *structureRepo) GetEquipe(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.Equipe, error) {
	for _, e := range r.equipes {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.Equipe{}, domain.ErrEquipeNotFound
}

func (r *structureRepo) UpdateEquipe(_ context.Context, tenant kernel.TenantID, equipeID uuid.UUID, libelle string, responsableID *uuid.UUID) (domain.Equipe, error) {
	for i := range r.equipes {
		if r.equipes[i].ID == equipeID {
			r.equipes[i].Libelle = libelle
			r.equipes[i].ResponsableID = responsableID
			r.equipes[i].TenantID = tenant
			r.savedEquipe = &r.equipes[i]
			return r.equipes[i], nil
		}
	}
	return domain.Equipe{}, domain.ErrEquipeNotFound
}

func (r *structureRepo) ListEquipes(context.Context, kernel.TenantID, ports.EquipeListFilter) ([]domain.Equipe, error) {
	return r.equipes, nil
}

func (r *structureRepo) SaveService(_ context.Context, s domain.Service) error {
	r.savedService = &s
	return nil
}

func (r *structureRepo) UpdateSite(_ context.Context, _ kernel.TenantID, siteID uuid.UUID, libelle string) (domain.SiteSummary, error) {
	for i := range r.sites {
		if r.sites[i].ID == siteID {
			r.sites[i].Libelle = libelle
			return r.sites[i], nil
		}
	}
	return domain.SiteSummary{}, domain.ErrSiteNotFound
}

func (r *structureRepo) UpdateService(_ context.Context, _ kernel.TenantID, serviceID uuid.UUID, libelle string) (domain.ServiceSummary, error) {
	if r.savedService != nil && r.savedService.ID == serviceID {
		r.savedService.Libelle = libelle
		return domain.ServiceSummary{ID: serviceID, Libelle: libelle, SiteID: r.savedService.SiteID}, nil
	}
	return domain.ServiceSummary{}, domain.ErrServiceNotFound
}

func (r *structureRepo) SaveApplication(_ context.Context, a domain.Application) error {
	r.savedApplication = &a
	if r.applications == nil {
		r.applications = map[uuid.UUID]domain.Application{}
	}
	r.applications[a.ID] = a
	return nil
}

func (r *structureRepo) GetApplication(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.Application, error) {
	if app, ok := r.applications[id]; ok {
		return app, nil
	}
	return domain.Application{}, domain.ErrApplicationNotFound
}

func (r *structureRepo) UpdateApplication(_ context.Context, a domain.Application, _ bool) error {
	if r.applications == nil {
		return domain.ErrApplicationNotFound
	}
	if _, ok := r.applications[a.ID]; !ok {
		return domain.ErrApplicationNotFound
	}
	r.applications[a.ID] = a
	r.savedApplication = &a
	return nil
}

func (r *structureRepo) ListSites(context.Context, kernel.TenantID) ([]domain.SiteSummary, error) {
	return r.sites, nil
}

func (r *structureRepo) UpdateUser(_ context.Context, u domain.User) error {
	r.updatedUser = &u
	return nil
}

func (r *structureRepo) ListUsers(context.Context, kernel.TenantID) ([]domain.User, error) {
	if len(r.listedUsers) > 0 {
		return r.listedUsers, nil
	}
	if r.user.ID != uuid.Nil {
		return []domain.User{r.user}, nil
	}
	return nil, nil
}

func TestCreateEquipe_persistsWithApplication(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	responsable := uuid.New()
	repo := &structureRepo{
		users: map[uuid.UUID]domain.User{
			responsable: {ID: responsable, TenantID: tenant, Login: "chef"},
		},
	}
	svc := NewOrganizationService(repo, nil)

	equipe, err := svc.CreateEquipe(context.Background(), ports.CreateEquipeCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		Libelle:       "  Équipe Dev  ",
		ResponsableID: &responsable,
	})
	if err != nil {
		t.Fatalf("CreateEquipe: %v", err)
	}
	if equipe.ID == uuid.Nil {
		t.Fatal("expected a generated equipe id")
	}
	if equipe.Libelle != "Équipe Dev" {
		t.Fatalf("libelle = %q, want trimmed", equipe.Libelle)
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
	svc := NewOrganizationService(repo, nil)

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

func TestCreateEquipe_rejectsEmptyLibelle(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.CreateEquipe(context.Background(), ports.CreateEquipeCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		ApplicationID: uuid.New(),
		Libelle:       "   ",
	})
	if !errors.Is(err, domain.ErrInvalidEquipeLibelle) {
		t.Fatalf("err = %v, want ErrInvalidEquipeLibelle", err)
	}
}

func TestCreateEquipe_rejectsUnknownResponsable(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	repo := &structureRepo{users: map[uuid.UUID]domain.User{}}
	svc := NewOrganizationService(repo, nil)
	unknown := uuid.New()
	_, err := svc.CreateEquipe(context.Background(), ports.CreateEquipeCommand{
		TenantID:      tenant,
		ApplicationID: uuid.New(),
		Libelle:       "Dev",
		ResponsableID: &unknown,
	})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
	if repo.savedEquipe != nil {
		t.Fatal("expected no persistence with unknown responsable")
	}
}

func TestCreateService_defaultsTypeToInterne(t *testing.T) {
	repo := &structureRepo{}
	svc := NewOrganizationService(repo, nil)

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
	svc := NewOrganizationService(repo, nil)

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

func TestUpdateSite_renames(t *testing.T) {
	siteID := uuid.New()
	repo := &structureRepo{sites: []domain.SiteSummary{{ID: siteID, Libelle: "Old"}}}
	svc := NewOrganizationService(repo, nil)
	got, err := svc.UpdateSite(context.Background(), ports.UpdateSiteCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		SiteID:   siteID,
		Libelle:  "  Nouveau site  ",
	})
	if err != nil {
		t.Fatalf("UpdateSite: %v", err)
	}
	if got.Libelle != "Nouveau site" {
		t.Fatalf("libelle = %q", got.Libelle)
	}
}

func TestUpdateSite_rejectsEmptyLibelle(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.UpdateSite(context.Background(), ports.UpdateSiteCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		SiteID:   uuid.New(),
		Libelle:  "   ",
	})
	if !errors.Is(err, domain.ErrInvalidSiteLibelle) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateService_renames(t *testing.T) {
	svcID := uuid.New()
	repo := &structureRepo{savedService: &domain.Service{ID: svcID, Libelle: "Old", SiteID: uuid.New()}}
	svc := NewOrganizationService(repo, nil)
	got, err := svc.UpdateService(context.Background(), ports.UpdateServiceCommand{
		TenantID:  kernel.NewTenantID(uuid.New()),
		ServiceID: svcID,
		Libelle:   "Delivery Renamed",
	})
	if err != nil {
		t.Fatalf("UpdateService: %v", err)
	}
	if got.Libelle != "Delivery Renamed" {
		t.Fatalf("libelle = %q", got.Libelle)
	}
}

func TestUpdateService_notFound(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.UpdateService(context.Background(), ports.UpdateServiceCommand{
		TenantID:  kernel.NewTenantID(uuid.New()),
		ServiceID: uuid.New(),
		Libelle:   "X",
	})
	if !errors.Is(err, domain.ErrServiceNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateEquipe_renamesAndSetsResponsable(t *testing.T) {
	equipeID := uuid.New()
	appID := uuid.New()
	responsable := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	repo := &structureRepo{
		equipes: []domain.Equipe{{
			ID:            equipeID,
			TenantID:      tenant,
			ApplicationID: appID,
			Libelle:       "Old",
		}},
		users: map[uuid.UUID]domain.User{
			responsable: {ID: responsable, TenantID: tenant, Login: "chef"},
		},
	}
	svc := NewOrganizationService(repo, nil)
	got, err := svc.UpdateEquipe(context.Background(), ports.UpdateEquipeCommand{
		TenantID:      tenant,
		EquipeID:      equipeID,
		Libelle:       "  Nouvelle équipe  ",
		ResponsableID: &responsable,
	})
	if err != nil {
		t.Fatalf("UpdateEquipe: %v", err)
	}
	if got.Libelle != "Nouvelle équipe" {
		t.Fatalf("libelle = %q", got.Libelle)
	}
	if got.ResponsableID == nil || *got.ResponsableID != responsable {
		t.Fatalf("responsable = %v", got.ResponsableID)
	}
	if got.ApplicationID != appID {
		t.Fatalf("applicationId changed to %v", got.ApplicationID)
	}
}

func TestUpdateEquipe_clearsResponsable(t *testing.T) {
	equipeID := uuid.New()
	responsable := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	repo := &structureRepo{
		equipes: []domain.Equipe{{
			ID:            equipeID,
			TenantID:      tenant,
			Libelle:       "Dev",
			ResponsableID: &responsable,
		}},
	}
	svc := NewOrganizationService(repo, nil)
	got, err := svc.UpdateEquipe(context.Background(), ports.UpdateEquipeCommand{
		TenantID:      tenant,
		EquipeID:      equipeID,
		Libelle:       "Dev",
		ResponsableID: nil,
	})
	if err != nil {
		t.Fatalf("UpdateEquipe: %v", err)
	}
	if got.ResponsableID != nil {
		t.Fatalf("responsable = %v, want nil", got.ResponsableID)
	}
}

func TestUpdateEquipe_rejectsEmptyLibelle(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.UpdateEquipe(context.Background(), ports.UpdateEquipeCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		EquipeID: uuid.New(),
		Libelle:  "   ",
	})
	if !errors.Is(err, domain.ErrInvalidEquipeLibelle) {
		t.Fatalf("err = %v, want ErrInvalidEquipeLibelle", err)
	}
}

func TestUpdateEquipe_notFound(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.UpdateEquipe(context.Background(), ports.UpdateEquipeCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		EquipeID: uuid.New(),
		Libelle:  "X",
	})
	if !errors.Is(err, domain.ErrEquipeNotFound) {
		t.Fatalf("err = %v, want ErrEquipeNotFound", err)
	}
}

func TestUpdateEquipe_rejectsUnknownResponsable(t *testing.T) {
	equipeID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	repo := &structureRepo{
		equipes: []domain.Equipe{{ID: equipeID, TenantID: tenant, Libelle: "Dev"}},
		users:   map[uuid.UUID]domain.User{},
	}
	svc := NewOrganizationService(repo, nil)
	unknown := uuid.New()
	_, err := svc.UpdateEquipe(context.Background(), ports.UpdateEquipeCommand{
		TenantID:      tenant,
		EquipeID:      equipeID,
		Libelle:       "Dev",
		ResponsableID: &unknown,
	})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
}

func TestListSites_delegatesToRepository(t *testing.T) {
	want := []domain.SiteSummary{{ID: uuid.New(), Libelle: "Paris HQ"}}
	svc := NewOrganizationService(&structureRepo{sites: want}, nil)

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
	equipeID := uuid.New()
	repo := &structureRepo{
		user:    user,
		equipes: []domain.Equipe{{ID: equipeID, TenantID: tenant}},
	}
	svc := newUserServiceForEquipe(t, repo)

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
	repo := &structureRepo{user: user}
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

func TestUpdateUser_allowsSelfEquipeChange(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	current := uuid.New()
	user := equipeTestUser(tenant, &current)
	other := uuid.New()
	repo := &structureRepo{
		user:    user,
		equipes: []domain.Equipe{{ID: other, TenantID: tenant}},
	}
	svc := newUserServiceForEquipe(t, repo)

	ids := []uuid.UUID{other}
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID:    tenant,
		UserID:      user.ID,
		ActorUserID: user.ID,
		EquipeIDs:   &ids,
	})
	if err != nil {
		t.Fatalf("UpdateUser self equipe: %v", err)
	}
	if repo.updatedUser == nil {
		t.Fatal("expected self equipe change to be persisted")
	}
	if len(repo.updatedUser.EquipeIDs) != 1 || repo.updatedUser.EquipeIDs[0] != other {
		t.Fatalf("EquipeIDs = %v, want [%v]", repo.updatedUser.EquipeIDs, other)
	}
}

func TestUpdateUser_allowsSelfProfilesChange(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	user := equipeTestUser(tenant, nil)
	user.Profile = domain.ProfileAdmin
	user.Profiles = []domain.Profile{domain.ProfileAdmin}
	repo := &structureRepo{user: user}
	svc := newUserServiceForEquipe(t, repo)

	profiles := []domain.Profile{domain.ProfileAdmin, domain.ProfileCollaborateur}
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID:    tenant,
		UserID:      user.ID,
		ActorUserID: user.ID,
		Profiles:    &profiles,
	})
	if err != nil {
		t.Fatalf("UpdateUser self profiles: %v", err)
	}
	if repo.updatedUser == nil {
		t.Fatal("expected self profiles change to be persisted")
	}
	if len(repo.updatedUser.Profiles) != 2 {
		t.Fatalf("Profiles = %v, want 2", repo.updatedUser.Profiles)
	}
}

func TestUpdateUser_blocksSelfDeactivate(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	user := equipeTestUser(tenant, nil)
	repo := &structureRepo{user: user}
	svc := newUserServiceForEquipe(t, repo)

	active := false
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID:    tenant,
		UserID:      user.ID,
		ActorUserID: user.ID,
		Active:      &active,
	})
	if !errors.Is(err, domain.ErrCannotModifySelf) {
		t.Fatalf("err = %v, want ErrCannotModifySelf", err)
	}
	if repo.updatedUser != nil {
		t.Fatal("expected no persistence on self deactivate")
	}
}

func TestUpdateUser_blocksSelfDemoteAdmin(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	user := equipeTestUser(tenant, nil)
	user.Profile = domain.ProfileAdmin
	user.Profiles = []domain.Profile{domain.ProfileAdmin}
	repo := &structureRepo{user: user}
	svc := newUserServiceForEquipe(t, repo)

	profiles := []domain.Profile{domain.ProfileCollaborateur}
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID:    tenant,
		UserID:      user.ID,
		ActorUserID: user.ID,
		Profiles:    &profiles,
	})
	if !errors.Is(err, domain.ErrCannotDemoteSelf) {
		t.Fatalf("err = %v, want ErrCannotDemoteSelf", err)
	}
	if repo.updatedUser != nil {
		t.Fatal("expected no persistence on self demotion")
	}
}

func TestUpdateUser_blocksLastAdminDemote(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	admin := equipeTestUser(tenant, nil)
	admin.Profile = domain.ProfileAdmin
	admin.Profiles = []domain.Profile{domain.ProfileAdmin}
	repo := &structureRepo{
		user:        admin,
		listedUsers: []domain.User{admin},
	}
	svc := newUserServiceForEquipe(t, repo)

	profiles := []domain.Profile{domain.ProfileCollaborateur}
	actor := uuid.New()
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID:    tenant,
		UserID:      admin.ID,
		ActorUserID: actor,
		Profiles:    &profiles,
	})
	if !errors.Is(err, domain.ErrLastAdmin) {
		t.Fatalf("err = %v, want ErrLastAdmin", err)
	}
	if repo.updatedUser != nil {
		t.Fatal("expected no persistence on last-admin demotion")
	}
}

func TestUpdateUser_allowsDemoteWhenOtherAdminExists(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	admin := equipeTestUser(tenant, nil)
	admin.Profile = domain.ProfileAdmin
	admin.Profiles = []domain.Profile{domain.ProfileAdmin}
	other := equipeTestUser(tenant, nil)
	other.Profile = domain.ProfileAdmin
	other.Profiles = []domain.Profile{domain.ProfileAdmin}
	repo := &structureRepo{
		user:        admin,
		listedUsers: []domain.User{admin, other},
	}
	svc := newUserServiceForEquipe(t, repo)

	profiles := []domain.Profile{domain.ProfileCollaborateur}
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID:    tenant,
		UserID:      admin.ID,
		ActorUserID: other.ID,
		Profiles:    &profiles,
	})
	if err != nil {
		t.Fatalf("UpdateUser demote with other admin: %v", err)
	}
	if repo.updatedUser == nil || domain.ProfilesContain(repo.updatedUser.Profiles, domain.ProfileAdmin) {
		t.Fatalf("expected demotion persisted without admin, got %+v", repo.updatedUser)
	}
}

func TestUpdateUser_keepsEquipeWhenFieldAbsent(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	current := uuid.New()
	user := equipeTestUser(tenant, &current)
	repo := &structureRepo{user: user}
	svc := newUserServiceForEquipe(t, repo)

	profile := domain.ProfileAdmin
	actor := uuid.New()
	_, err := svc.UpdateUser(context.Background(), ports.UpdateUserCommand{
		TenantID:    tenant,
		UserID:      user.ID,
		ActorUserID: actor,
		Profile:     &profile,
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

func TestCreateApplication_setsActive(t *testing.T) {
	repo := &structureRepo{}
	svc := NewOrganizationService(repo, nil)

	app, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID:   kernel.NewTenantID(uuid.New()),
		ServiceIDs: []uuid.UUID{uuid.New()},
		Libelle:    "CRM",
	})
	if err != nil {
		t.Fatalf("CreateApplication: %v", err)
	}
	if !app.Active {
		t.Fatal("expected new application to be active")
	}
	if app.ModeFacturation != domain.DefaultModeFacturation {
		t.Fatalf("mode = %q, want default", app.ModeFacturation)
	}
	if repo.savedApplication == nil || !repo.savedApplication.Active {
		t.Fatal("expected SaveApplication with Active=true")
	}
}

func TestCreateApplication_persistsRichFields(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	chefID := uuid.New()
	repo := &structureRepo{}
	repo.users = map[uuid.UUID]domain.User{
		chefID: {ID: chefID, TenantID: tenant, Login: "CHEF_user", Active: true},
	}
	svc := NewOrganizationService(repo, nil)

	app, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID:          tenant,
		ServiceIDs:        []uuid.UUID{uuid.New()},
		Libelle:           "Portail",
		Proprietaire:      "Client ACME",
		ModeFacturation:   domain.ModeFacturationForfait,
		UOActivee:         true,
		ChefUtilisateurID: &chefID,
	})
	if err != nil {
		t.Fatalf("CreateApplication: %v", err)
	}
	if app.Proprietaire != "Client ACME" || app.ModeFacturation != domain.ModeFacturationForfait || !app.UOActivee {
		t.Fatalf("got %+v", app)
	}
	if app.ChefUtilisateurID == nil || *app.ChefUtilisateurID != chefID {
		t.Fatalf("chef = %v", app.ChefUtilisateurID)
	}
}

func TestCreateApplication_rejectsBudgetOnCreate(t *testing.T) {
	budgetID := uuid.New()
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID:       kernel.NewTenantID(uuid.New()),
		ServiceIDs:     []uuid.UUID{uuid.New()},
		Libelle:        "X",
		BudgetDefautID: &budgetID,
	})
	if !errors.Is(err, domain.ErrBudgetNotAllowedOnCreate) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateApplication_rejectsEmptyLibelle(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID:   kernel.NewTenantID(uuid.New()),
		ServiceIDs: []uuid.UUID{uuid.New()},
		Libelle:    "   ",
	})
	if !errors.Is(err, domain.ErrInvalidApplicationLibelle) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateApplication_rejectsInvalidMode(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID:        kernel.NewTenantID(uuid.New()),
		ServiceIDs:      []uuid.UUID{uuid.New()},
		Libelle:         "X",
		ModeFacturation: "inconnu",
	})
	if !errors.Is(err, domain.ErrInvalidModeFacturation) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateApplication_rejectsUnknownChef(t *testing.T) {
	chefID := uuid.New()
	svc := NewOrganizationService(&structureRepo{users: map[uuid.UUID]domain.User{}}, nil)
	_, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID:          kernel.NewTenantID(uuid.New()),
		ServiceIDs:        []uuid.UUID{uuid.New()},
		Libelle:           "X",
		ChefUtilisateurID: &chefID,
	})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateApplication_renamesAndDeactivates(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	repo := &structureRepo{applications: map[uuid.UUID]domain.Application{
		appID: {
			ID:         appID,
			TenantID:   tenant,
			ServiceIDs: []uuid.UUID{uuid.New()},
			Libelle:    "Old",
			Active:     true,
		},
	}}
	svc := NewOrganizationService(repo, nil)

	libelle := "New"
	active := false
	got, err := svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		Libelle:       &libelle,
		Active:        &active,
	})
	if err != nil {
		t.Fatalf("UpdateApplication: %v", err)
	}
	if got.Libelle != "New" || got.Active {
		t.Fatalf("got %+v", got)
	}

	_, err = svc.SetApplicationActive(context.Background(), ports.SetApplicationActiveCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		Active:        true,
	})
	if err != nil {
		t.Fatalf("SetApplicationActive: %v", err)
	}
	if !repo.applications[appID].Active {
		t.Fatal("expected application reactivated")
	}
}

func TestUpdateApplication_richFields(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	chefID := uuid.New()
	repo := &structureRepo{
		applications: map[uuid.UUID]domain.Application{
			appID: {
				ID:              appID,
				TenantID:        tenant,
				ServiceIDs:      []uuid.UUID{uuid.New()},
				Libelle:         "App",
				ModeFacturation: domain.ModeFacturationTempsPasse,
				Active:          true,
			},
		},
	}
	repo.users = map[uuid.UUID]domain.User{
		chefID: {ID: chefID, TenantID: tenant, Login: "CHEF_user", Active: true},
	}
	svc := NewOrganizationService(repo, nil)

	prop := "Société X"
	mode := domain.ModeFacturationNon
	uo := true
	chefPtr := &chefID
	got, err := svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:          tenant,
		ApplicationID:     appID,
		Proprietaire:      &prop,
		ModeFacturation:   &mode,
		UOActivee:         &uo,
		ChefUtilisateurID: &chefPtr,
	})
	if err != nil {
		t.Fatalf("UpdateApplication: %v", err)
	}
	if got.Proprietaire != prop || got.ModeFacturation != mode || !got.UOActivee {
		t.Fatalf("got %+v", got)
	}
	if got.ChefUtilisateurID == nil || *got.ChefUtilisateurID != chefID {
		t.Fatalf("chef = %v", got.ChefUtilisateurID)
	}

	var clear *uuid.UUID
	got, err = svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:          tenant,
		ApplicationID:     appID,
		ChefUtilisateurID: &clear,
	})
	if err != nil {
		t.Fatalf("clear chef: %v", err)
	}
	if got.ChefUtilisateurID != nil {
		t.Fatal("expected chef cleared")
	}
}

func TestUpdateApplication_budgetDefaut(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	budgetID := uuid.New()
	repo := &structureRepo{
		applications: map[uuid.UUID]domain.Application{
			appID: {ID: appID, TenantID: tenant, ServiceIDs: []uuid.UUID{uuid.New()}, Libelle: "App", Active: true},
		},
		// Simulates EXISTS (... AND type = 'defaut')
		budgetOK: map[uuid.UUID]bool{budgetID: true},
	}
	svc := NewOrganizationService(repo, nil)
	budgetPtr := &budgetID
	got, err := svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:       tenant,
		ApplicationID:  appID,
		BudgetDefautID: &budgetPtr,
	})
	if err != nil {
		t.Fatalf("UpdateApplication: %v", err)
	}
	if got.BudgetDefautID == nil || *got.BudgetDefautID != budgetID {
		t.Fatalf("budget = %v", got.BudgetDefautID)
	}

	// Budget not marked as default-type for the app → rejected (specifique / other app).
	specific := uuid.New()
	specificPtr := &specific
	_, err = svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:       tenant,
		ApplicationID:  appID,
		BudgetDefautID: &specificPtr,
	})
	if !errors.Is(err, domain.ErrBudgetNotFound) {
		t.Fatalf("err = %v, want ErrBudgetNotFound for non-default budget", err)
	}

	clear := (*uuid.UUID)(nil)
	got, err = svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:       tenant,
		ApplicationID:  appID,
		BudgetDefautID: &clear,
	})
	if err != nil {
		t.Fatalf("clear budget: %v", err)
	}
	if got.BudgetDefautID != nil {
		t.Fatal("expected budget cleared")
	}
}

func TestUpdateApplication_notFound(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	libelle := "x"
	_, err := svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		ApplicationID: uuid.New(),
		Libelle:       &libelle,
	})
	if !errors.Is(err, domain.ErrApplicationNotFound) {
		t.Fatalf("err = %v, want ErrApplicationNotFound", err)
	}
}

func TestCreateApplication_rejectsWithoutShare(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		Libelle:  "Orpheline",
	})
	if !errors.Is(err, domain.ErrApplicationWithoutShare) {
		t.Fatalf("err = %v, want ErrApplicationWithoutShare", err)
	}
}

func TestCreateApplication_multiShares(t *testing.T) {
	svcID1, svcID2 := uuid.New(), uuid.New()
	siteID := uuid.New()
	repo := &structureRepo{}
	svc := NewOrganizationService(repo, nil)
	app, err := svc.CreateApplication(context.Background(), ports.CreateApplicationCommand{
		TenantID:   kernel.NewTenantID(uuid.New()),
		Libelle:    "Shared",
		SiteIDs:    []uuid.UUID{siteID},
		ServiceIDs: []uuid.UUID{svcID1, svcID2, svcID1},
	})
	if err != nil {
		t.Fatalf("CreateApplication: %v", err)
	}
	if len(app.ServiceIDs) != 2 {
		t.Fatalf("serviceIds = %v, want 2 unique", app.ServiceIDs)
	}
	if len(app.SiteIDs) != 1 || app.SiteIDs[0] != siteID {
		t.Fatalf("siteIds = %v", app.SiteIDs)
	}
}

func TestUpdateApplication_replaceShares(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	appID := uuid.New()
	oldSvc, newSvc, siteID, equipeID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo := &structureRepo{applications: map[uuid.UUID]domain.Application{
		appID: {
			ID: appID, TenantID: tenant, Libelle: "App", Active: true,
			ServiceIDs: []uuid.UUID{oldSvc},
			SiteIDs:    []uuid.UUID{siteID},
			EquipeIDs:  []uuid.UUID{equipeID},
		},
	}}
	svc := NewOrganizationService(repo, nil)
	services := []uuid.UUID{newSvc}
	got, err := svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		ServiceIDs:    &services,
	})
	if err != nil {
		t.Fatalf("UpdateApplication: %v", err)
	}
	if len(got.ServiceIDs) != 1 || got.ServiceIDs[0] != newSvc {
		t.Fatalf("serviceIds = %v", got.ServiceIDs)
	}
	// Partial PUT: omitted categories are preserved.
	if len(got.SiteIDs) != 1 || got.SiteIDs[0] != siteID {
		t.Fatalf("siteIds = %v, want preserved", got.SiteIDs)
	}
	if len(got.EquipeIDs) != 1 || got.EquipeIDs[0] != equipeID {
		t.Fatalf("equipeIds = %v, want preserved", got.EquipeIDs)
	}
	emptyServices := []uuid.UUID{}
	emptySites := []uuid.UUID{}
	emptyEquipes := []uuid.UUID{}
	_, err = svc.UpdateApplication(context.Background(), ports.UpdateApplicationCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		ServiceIDs:    &emptyServices,
		SiteIDs:       &emptySites,
		EquipeIDs:     &emptyEquipes,
	})
	if !errors.Is(err, domain.ErrApplicationWithoutShare) {
		t.Fatalf("err = %v, want ErrApplicationWithoutShare", err)
	}
}

func TestSetApplicationActive_notFound(t *testing.T) {
	svc := NewOrganizationService(&structureRepo{}, nil)
	_, err := svc.SetApplicationActive(context.Background(), ports.SetApplicationActiveCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		ApplicationID: uuid.New(),
		Active:        false,
	})
	if !errors.Is(err, domain.ErrApplicationNotFound) {
		t.Fatalf("err = %v", err)
	}
}
