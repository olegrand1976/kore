package seed

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	billingdomain "github.com/kore/kore/internal/modules/billing/domain"
	budgetdomain "github.com/kore/kore/internal/modules/budget/domain"
	budgetports "github.com/kore/kore/internal/modules/budget/ports"
	congesdomain "github.com/kore/kore/internal/modules/conges/domain"
	congesports "github.com/kore/kore/internal/modules/conges/ports"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	notifdomain "github.com/kore/kore/internal/modules/notifications/domain"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	publicports "github.com/kore/kore/internal/modules/publicsite/ports"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	wfports "github.com/kore/kore/internal/modules/workflow/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type TrialEnsurer interface {
	EnsureTrial(ctx context.Context, tenantID kernel.TenantID, seats int, modules []billingdomain.ModuleCode) error
}

type PublicSlotSeeder interface {
	SeedSlot(ctx context.Context, commercialID uuid.UUID, start, end time.Time) error
}

type Dependencies struct {
	Pool         *db.Pool
	OrgRepo      orgports.OrganizationRepository
	Org          orgports.OrganizationService
	Users        orgports.UserService
	Clients      orgports.ClientService
	Billing      TrialEnsurer
	Workflow     wfports.WorkflowService
	CRA          craports.CRAService
	Leaves       congesports.LeaveService
	Budget       budgetports.BudgetService
	TMA          tmaports.TMAService
	Notifications notifports.NotificationService
	Public       publicports.PublicSiteService
	PublicSlots  PublicSlotSeeder
}

type Runner struct {
	deps Dependencies
}

func NewRunner(deps Dependencies) *Runner {
	return &Runner{deps: deps}
}

type orgContext struct {
	tenant      kernel.TenantID
	adminID     uuid.UUID
	managerID   uuid.UUID
	collabID    uuid.UUID
	commercialID uuid.UUID
	appID       uuid.UUID
}

func (r *Runner) Run(ctx context.Context) error {
	tenant := kernel.NewTenantID(DemoTenantID)
	if err := r.ensureTenant(ctx); err != nil {
		return err
	}
	if err := r.ensureTrial(ctx, tenant); err != nil {
		return err
	}
	if err := r.ensureWorkflows(ctx, tenant); err != nil {
		return err
	}

	seeded, err := r.deps.OrgRepo.ExistsLogin(ctx, tenant, MarkerLogin)
	if err != nil {
		return err
	}
	if seeded {
		log.Println("seed: jeu de données demo déjà présent, skip données métier")
		return r.ensureNotificationRules(ctx, tenant)
	}

	oc, err := r.seedOrg(ctx, tenant)
	if err != nil {
		return err
	}
	if err := r.seedBudget(ctx, tenant, oc.appID); err != nil {
		return err
	}
	if err := r.seedCRA(ctx, tenant, oc); err != nil {
		return err
	}
	if err := r.seedConges(ctx, tenant, oc); err != nil {
		return err
	}
	if err := r.seedTMA(ctx, tenant, oc); err != nil {
		return err
	}
	if err := r.seedPublicsite(ctx, oc); err != nil {
		return err
	}
	if err := r.ensureNotificationRules(ctx, tenant); err != nil {
		return err
	}

	log.Println("seed: jeu de données demo complet appliqué")
	return nil
}

func (r *Runner) ensureTenant(ctx context.Context) error {
	return r.deps.OrgRepo.SaveTenant(ctx, orgdomain.Tenant{ID: DemoTenantID, Name: TenantName})
}

func (r *Runner) ensureTrial(ctx context.Context, tenant kernel.TenantID) error {
	if r.deps.Billing == nil {
		return nil
	}
	modules := []billingdomain.ModuleCode{
		billingdomain.ModuleOrg,
		billingdomain.ModuleCRA,
		billingdomain.ModuleConges,
		billingdomain.ModuleBudget,
		billingdomain.ModuleTMA,
		billingdomain.ModuleWorkflow,
		billingdomain.ModuleNotifications,
		billingdomain.ModuleBilling,
	}
	return r.deps.Billing.EnsureTrial(ctx, tenant, TrialSeats, modules)
}

func (r *Runner) seedOrg(ctx context.Context, tenant kernel.TenantID) (orgContext, error) {
	oc := orgContext{tenant: tenant, appID: DemoAppID}

	admin, err := r.ensureUser(ctx, tenant, AdminLogin, AdminPassword, orgdomain.ProfileAdmin, nil)
	if err != nil {
		return oc, err
	}
	oc.adminID = admin.ID

	manager, err := r.ensureUser(ctx, tenant, ManagerLogin, ManagerPassword, orgdomain.Profile("Chef d'équipe"), nil)
	if err != nil {
		return oc, err
	}
	oc.managerID = manager.ID

	commercial, err := r.ensureUser(ctx, tenant, CommercialLogin, CommercialPass, orgdomain.ProfileCollaborateur, nil)
	if err != nil {
		return oc, err
	}
	oc.commercialID = commercial.ID

	if err := r.ensureSociete(ctx, tenant); err != nil {
		return oc, err
	}
	if err := r.ensureSite(ctx, tenant); err != nil {
		return oc, err
	}
	if err := r.ensureService(ctx, tenant, oc.managerID); err != nil {
		return oc, err
	}
	if err := r.ensureApplication(ctx, tenant); err != nil {
		return oc, err
	}
	if err := r.ensureEquipe(ctx, tenant, oc.managerID); err != nil {
		return oc, err
	}

	collab, err := r.ensureUser(ctx, tenant, CollabLogin, CollabPassword, orgdomain.ProfileCollaborateur, &DemoEquipeID)
	if err != nil {
		return oc, err
	}
	oc.collabID = collab.ID

	if err := r.ensureClient(ctx, tenant); err != nil {
		return oc, err
	}

	log.Println("seed: organisation demo créée")
	return oc, nil
}

func (r *Runner) ensureUser(
	ctx context.Context,
	tenant kernel.TenantID,
	login, password string,
	profile orgdomain.Profile,
	equipeID *uuid.UUID,
) (orgdomain.User, error) {
	exists, err := r.deps.OrgRepo.ExistsLogin(ctx, tenant, login)
	if err != nil {
		return orgdomain.User{}, err
	}
	if exists {
		return r.deps.OrgRepo.FindUserByLogin(ctx, tenant, login)
	}
	return r.deps.Users.CreateUser(ctx, orgports.CreateUserCommand{
		TenantID: tenant,
		Login:    login,
		Password: password,
		Profile:  profile,
		EquipeID: equipeID,
	})
}

func (r *Runner) ensureSociete(ctx context.Context, tenant kernel.TenantID) error {
	exists, err := r.rowExists(ctx, `SELECT EXISTS(SELECT 1 FROM org.societes WHERE id = $1)`, DemoSocieteID)
	if err != nil || exists {
		return err
	}
	return r.deps.OrgRepo.SaveSociete(ctx, orgdomain.Societe{
		ID:            DemoSocieteID,
		TenantID:      tenant,
		RaisonSociale: DemoSocieteName,
		Devise:        "EUR",
		Adresse:       "1 rue de la Démo, 75001 Paris",
		Siret:         "12345678901234",
		URLTenant:     "demo.kore.local",
	})
}

func (r *Runner) ensureSite(ctx context.Context, tenant kernel.TenantID) error {
	exists, err := r.rowExists(ctx, `SELECT EXISTS(SELECT 1 FROM org.sites WHERE id = $1)`, DemoSiteID)
	if err != nil || exists {
		return err
	}
	return r.deps.OrgRepo.SaveSite(ctx, orgdomain.Site{
		ID:        DemoSiteID,
		TenantID:  tenant,
		SocieteID: DemoSocieteID,
		Libelle:   DemoSiteLabel,
	})
}

func (r *Runner) ensureService(ctx context.Context, tenant kernel.TenantID, managerID uuid.UUID) error {
	exists, err := r.rowExists(ctx, `SELECT EXISTS(SELECT 1 FROM org.services WHERE id = $1)`, DemoServiceID)
	if err != nil || exists {
		return err
	}
	return r.deps.OrgRepo.SaveService(ctx, orgdomain.Service{
		ID:            DemoServiceID,
		TenantID:      tenant,
		SiteID:        DemoSiteID,
		ResponsableID: &managerID,
	})
}

func (r *Runner) ensureApplication(ctx context.Context, tenant kernel.TenantID) error {
	exists, err := r.rowExists(ctx, `SELECT EXISTS(SELECT 1 FROM org.applications WHERE id = $1)`, DemoAppID)
	if err != nil || exists {
		return err
	}
	return r.deps.OrgRepo.SaveApplication(ctx, orgdomain.Application{
		ID:        DemoAppID,
		TenantID:  tenant,
		ServiceID: DemoServiceID,
		Libelle:   DemoAppLabel,
	})
}

func (r *Runner) ensureEquipe(ctx context.Context, tenant kernel.TenantID, managerID uuid.UUID) error {
	exists, err := r.rowExists(ctx, `SELECT EXISTS(SELECT 1 FROM org.equipes WHERE id = $1)`, DemoEquipeID)
	if err != nil || exists {
		return err
	}
	_, err = r.deps.Pool.Exec(ctx, `
		INSERT INTO org.equipes (id, tenant_id, application_id, libelle, responsable_id)
		VALUES ($1, $2, $3, $4, $5)
	`, DemoEquipeID, tenant.UUID(), DemoAppID, DemoEquipeLabel, managerID)
	return err
}

func (r *Runner) ensureClient(ctx context.Context, tenant kernel.TenantID) error {
	clients, err := r.deps.Clients.ListClients(ctx, tenant)
	if err != nil {
		return err
	}
	for _, c := range clients {
		if c.RaisonSociale == DemoClientName {
			return nil
		}
	}
	_, err = r.deps.Clients.CreateClient(ctx, orgports.CreateClientCommand{
		TenantID:      tenant,
		RaisonSociale: DemoClientName,
		TVA:           DemoClientTVA,
	})
	return err
}

func (r *Runner) seedBudget(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID) error {
	budgets, err := r.deps.Budget.List(ctx, tenant)
	if err != nil {
		return err
	}
	for _, b := range budgets {
		if b.ApplicationID == appID && b.Type == budgetdomain.BudgetTypeDefault {
			return nil
		}
	}
	_, err = r.deps.Budget.CreateBudget(ctx, budgetports.CreateBudgetCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		Type:          budgetdomain.BudgetTypeDefault,
		PlannedDays:   120,
		PlannedUO:     600,
		PlannedAmount: 12000000,
		Currency:      "EUR",
	})
	if err != nil {
		return err
	}
	log.Println("seed: budget défaut créé")
	return nil
}

func (r *Runner) seedCRA(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	now := time.Now().UTC()
	month, err := cradomain.ParseMonth(now.Format("2006-01"))
	if err != nil {
		return err
	}
	ts, err := r.deps.CRA.GetOrCreate(ctx, tenant, oc.collabID, month)
	if err != nil {
		return err
	}
	if len(ts.Weeks) > 0 && len(ts.Weeks[0].Lines) > 0 {
		return nil
	}

	day := time.Date(now.Year(), now.Month(), 3, 0, 0, 0, 0, time.UTC)
	duration, err := kernel.NewDuration(420)
	if err != nil {
		return err
	}
	_, err = r.deps.CRA.SaveWeek(ctx, craports.SaveWeekCommand{
		TenantID:    tenant,
		TimesheetID: ts.ID,
		WeekNumber:  1,
		Lines: []cradomain.TimeLine{
			{
				Source:   cradomain.SourceRef{Type: "mission", ID: oc.appID.String()},
				Day:      day,
				Duration: duration,
				Comment:  "Développement portail client",
			},
			{
				Source:   cradomain.SourceRef{Type: "mission", ID: oc.appID.String()},
				Day:      day.AddDate(0, 0, 1),
				Duration: duration,
				Comment:  "Atelier fonctionnel",
			},
		},
	})
	if err != nil {
		return err
	}
	if err := r.deps.CRA.CompleteCommercialInfo(ctx, craports.CommercialCommand{
		TenantID:    tenant,
		TimesheetID: ts.ID,
		Info: cradomain.CommercialInfo{
			Client:  DemoClientName,
			Mission: DemoAppLabel,
		},
	}); err != nil {
		return err
	}
	log.Println("seed: CRA collaborateur alimenté")
	return nil
}

func (r *Runner) seedConges(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	if err := r.ensureLeaveBalance(ctx, tenant, oc.collabID); err != nil {
		return err
	}

	requests, err := r.deps.Leaves.List(ctx, tenant, &oc.collabID, nil)
	if err != nil {
		return err
	}
	if len(requests) > 0 {
		return nil
	}

	start := nextMonday(time.Now().UTC().AddDate(0, 0, 14))
	end := start.AddDate(0, 0, 2)
	_, err = r.deps.Leaves.Request(ctx, congesports.RequestLeaveCommand{
		TenantID: tenant,
		UserID:   oc.collabID,
		Type:     congesdomain.LeaveTypeCongesPayes,
		From:     start,
		To:       end,
		Motif:    "Congés été (demo)",
	})
	if err != nil {
		return err
	}
	log.Println("seed: demande de congé demo créée")
	return nil
}

func (r *Runner) ensureLeaveBalance(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) error {
	_, err := r.deps.Pool.Exec(ctx, `
		INSERT INTO conges.leave_balances (id, tenant_id, user_id, type, acquired, taken, remaining)
		VALUES ($1, $2, $3, $4, 25, 3, 22)
		ON CONFLICT (tenant_id, user_id, type) DO NOTHING
	`, uuid.New(), tenant.UUID(), userID, string(congesdomain.LeaveTypeCongesPayes))
	return err
}

func (r *Runner) seedTMA(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	demands, err := r.deps.TMA.List(ctx, tenant, tmaports.ExportFilter{})
	if err != nil {
		return err
	}
	if len(demands) > 0 {
		return nil
	}

	open, err := r.deps.TMA.CreateDemand(ctx, tmaports.CreateDemandCommand{
		TenantID:         tenant,
		ApplicationID:    oc.appID,
		AuthorID:         oc.collabID,
		Subject:          "Erreur export PDF factures",
		RequiresChefGate: false,
	})
	if err != nil {
		return err
	}

	assigned, err := r.deps.TMA.CreateDemand(ctx, tmaports.CreateDemandCommand{
		TenantID:         tenant,
		ApplicationID:    oc.appID,
		AuthorID:         oc.collabID,
		Subject:          "Lenteur écran de saisie CRA",
		RequiresChefGate: false,
	})
	if err != nil {
		return err
	}
	if err := r.deps.TMA.Assign(ctx, tmaports.AssignCommand{
		TenantID:   tenant,
		ID:         assigned.ID,
		AssigneeID: oc.collabID,
		ActorID:    oc.managerID,
	}); err != nil {
		return err
	}
	if err := r.deps.TMA.AddAnalysis(ctx, tmaports.AnalysisCommand{
		TenantID:     tenant,
		DemandID:     assigned.ID,
		Functional:   "Temps de rendu > 3s sur mobile",
		Technical:    "Requêtes N+1 sur agrégation semaines",
		Risks:        "Impact faible",
		TestScenario: "Saisie 5 jours sur iPhone SE",
	}); err != nil {
		return err
	}
	_ = open
	log.Println("seed: demandes TMA demo créées")
	return nil
}

func (r *Runner) seedPublicsite(ctx context.Context, oc orgContext) error {
	if r.deps.PublicSlots != nil {
		start := nextWeekdayAt(time.Now().UTC(), time.Tuesday, 10, 0)
		end := start.Add(30 * time.Minute)
		if err := r.deps.PublicSlots.SeedSlot(ctx, oc.commercialID, start, end); err != nil {
			return err
		}
		start2 := start.Add(24 * time.Hour)
		if err := r.deps.PublicSlots.SeedSlot(ctx, oc.commercialID, start2, start2.Add(30*time.Minute)); err != nil {
			return err
		}
	}

	if r.deps.Public != nil {
		var count int
		err := r.deps.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM publicsite.leads WHERE email = $1`, "demo@acme.test").Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err = r.deps.Public.CaptureLead(ctx, publicports.CaptureLeadCommand{
				Email:     "demo@acme.test",
				Company:   DemoClientName,
				Size:      "50-200",
				Need:      "PSA modulaire",
				UTMSource: "seed",
				Consent:   true,
			})
			if err != nil {
				return err
			}
		}
	}

	log.Println("seed: publicsite (créneaux + lead) alimenté")
	return nil
}

func (r *Runner) ensureNotificationRules(ctx context.Context, tenant kernel.TenantID) error {
	if r.deps.Notifications == nil {
		return nil
	}
	rules := []notifdomain.NotificationRule{
		{
			TenantID:   tenant,
			Code:       "leave-requested",
			Trigger:    "leave.requested",
			Frequency:  notifdomain.FrequencyImmediate,
			Template:   "Nouvelle demande de congé de {{user}}.",
			AttachPDF:  false,
			RecipientsPolicy: notifdomain.RecipientPolicy{
				UserIDs: []uuid.UUID{},
			},
		},
		{
			TenantID:  tenant,
			Code:      "tma-demand-created",
			Trigger:   "tma.demand.created",
			Frequency: notifdomain.FrequencyMorning,
			Template:  "Nouvelle demande TMA : {{subject}}.",
		},
	}
	for _, rule := range rules {
		existing, err := r.deps.Notifications.ListRules(ctx, tenant)
		if err != nil {
			return err
		}
		found := false
		for _, e := range existing {
			if e.Trigger == rule.Trigger {
				found = true
				break
			}
		}
		if found {
			continue
		}
		if err := r.deps.Notifications.DefineRule(ctx, rule); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) rowExists(ctx context.Context, query string, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.deps.Pool.QueryRow(ctx, query, id).Scan(&exists)
	return exists, err
}

func nextMonday(from time.Time) time.Time {
	day := from.UTC()
	for day.Weekday() != time.Monday {
		day = day.AddDate(0, 0, 1)
	}
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}

func nextWeekdayAt(from time.Time, weekday time.Weekday, hour, minute int) time.Time {
	day := from.UTC()
	for i := 0; i < 14; i++ {
		if day.Weekday() == weekday && day.After(from) {
			return time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, time.UTC)
		}
		day = day.AddDate(0, 0, 1)
	}
	return from.Add(48 * time.Hour)
}
