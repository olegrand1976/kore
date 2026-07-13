package seed

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	congesdomain "github.com/kore/kore/internal/modules/conges/domain"
	congesports "github.com/kore/kore/internal/modules/conges/ports"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

const (
	ChefDevLogin     = "CHE_chefdev"
	ChefDevPassword  = "Chef123!"
	Collab3Login     = "COL_dev3"
	CollabQALogin    = "COL_qa1"
	CollabIntegLogin = "COL_integ1"
	Presta2Login     = "PRE_presta2"
	Presta3Login     = "PRE_data"
	PrestaIntegLogin = "PRE_integ"
	ClientMOALogin   = "CLI_moa1"
	ClientDSILogin   = "CLI_dsi_gx"
	ClientPMOLogin   = "CLI_pmo_gx"
	Commercial2Login = "COM_sales2"
	Commercial3Login = "COM_sales3"

	DemoApp3Label    = "Data Hub Initech"
	DemoEquipe3Label = "Équipe Data & Intégration"
	DemoEquipeDevExt = "Équipe Dev — renfort"
	DemoClient3Name  = "Initech SA"
	DemoClient3TVA   = "FR11223344556"
)

var (
	DemoApp3ID    = uuid.MustParse("00000000-0000-4000-8000-000000000017")
	DemoEquipe3ID = uuid.MustParse("00000000-0000-4000-8000-000000000018")
	DemoEquipe4ID = uuid.MustParse("00000000-0000-4000-8000-000000000019")
)

type extendedUserDef struct {
	login, password string
	profile         orgdomain.Profile
	equipeID        *uuid.UUID
	typeCompte      string
	craRequis       bool
	salarieETT      bool
}

func extendedUserDefs() []extendedUserDef {
	eData := DemoEquipe3ID
	eDevExt := DemoEquipe4ID
	eTMA := DemoEquipe2ID
	return []extendedUserDef{
		{ChefDevLogin, ChefDevPassword, orgdomain.Profile("Chef d'équipe"), &DemoEquipeID, "Interne", true, false},
		{Collab3Login, CollabPassword, orgdomain.ProfileCollaborateur, &eDevExt, "Interne", true, false},
		{CollabQALogin, CollabPassword, orgdomain.ProfileCollaborateur, &eDevExt, "Interne", true, false},
		{CollabIntegLogin, CollabPassword, orgdomain.ProfileCollaborateur, &eData, "Interne", true, false},
		{Presta2Login, PrestaPassword, orgdomain.ProfileCollaborateur, &eTMA, "Prestataire", true, true},
		{Presta3Login, PrestaPassword, orgdomain.ProfileCollaborateur, &eData, "Prestataire", true, true},
		{PrestaIntegLogin, PrestaPassword, orgdomain.ProfileCollaborateur, &eData, "Prestataire", true, true},
		{ClientMOALogin, ClientUserPass, orgdomain.ProfileCollaborateur, nil, "Client", false, false},
		{ClientDSILogin, ClientUserPass, orgdomain.ProfileCollaborateur, nil, "Client", false, false},
		{ClientPMOLogin, ClientUserPass, orgdomain.ProfileCollaborateur, nil, "Client", false, false},
		{Commercial2Login, CommercialPass, orgdomain.ProfileCollaborateur, nil, "Interne", false, false},
		{Commercial3Login, CommercialPass, orgdomain.ProfileCollaborateur, nil, "Interne", false, false},
	}
}

func (r *Runner) seedExtendedOrg(ctx context.Context, tenant kernel.TenantID, oc *orgContext) error {
	if oc.usersByLogin == nil {
		oc.usersByLogin = make(map[string]uuid.UUID)
	}
	oc.registerUser(AdminLogin, oc.adminID)
	oc.registerUser(ManagerLogin, oc.managerID)
	oc.registerUser(CommercialLogin, oc.commercialID)
	oc.registerUser(CollabLogin, oc.collabID)
	oc.registerUser(Collab2Login, oc.collab2ID)
	oc.registerUser(PrestaLogin, oc.prestaID)
	oc.registerUser(ClientUserLogin, oc.clientUserID)

	if err := r.ensureApplication3(ctx, tenant); err != nil {
		return err
	}
	if err := r.ensureEquipeNamed(ctx, tenant, DemoEquipe3ID, DemoApp3ID, DemoEquipe3Label, oc.managerID); err != nil {
		return err
	}
	if err := r.ensureEquipeNamed(ctx, tenant, DemoEquipe4ID, DemoAppID, DemoEquipeDevExt, oc.managerID); err != nil {
		return err
	}
	oc.app3ID = DemoApp3ID
	oc.equipeDataID = DemoEquipe3ID
	oc.equipeDevExtID = DemoEquipe4ID

	if err := r.ensureClientNamed(ctx, tenant, DemoClient3Name, DemoClient3TVA); err != nil {
		return err
	}

	for _, def := range extendedUserDefs() {
		user, err := r.ensureUser(ctx, tenant, def.login, def.password, def.profile, def.equipeID)
		if err != nil {
			return err
		}
		oc.registerUser(def.login, user.ID)
		if err := r.patchUserMeta(ctx, tenant, def.login, def.typeCompte, "fr", def.craRequis, def.salarieETT); err != nil {
			return err
		}
		switch def.typeCompte {
		case "Prestataire":
			oc.prestaIDs = append(oc.prestaIDs, user.ID)
		case "Client":
			oc.clientUserIDs = append(oc.clientUserIDs, user.ID)
		case "Interne":
			if def.craRequis {
				oc.collabIDs = append(oc.collabIDs, user.ID)
			} else {
				oc.commercialIDs = append(oc.commercialIDs, user.ID)
			}
		}
	}

	if chefID, ok := oc.usersByLogin[ChefDevLogin]; ok {
		oc.chefEquipeID = chefID
		oc.collabIDs = append(oc.collabIDs, chefID)
	}

	if err := r.ensureClientContacts(ctx, tenant, DemoClient3Name, clientContactsInitech); err != nil {
		return err
	}

	log.Printf("seed: équipe élargie — %d utilisateurs, 4 équipes, 3 missions", len(oc.usersByLogin))
	return nil
}

const clientContactsInitech = `[{"nom":"Lumbergh","prenom":"Bill","email":"bill.lumbergh@initech.test","role":"VP Engineering","telephone":"+33 1 55 00 00 01"},{"nom":"Smykowski","prenom":"Tom","email":"tom.smykowski@initech.test","role":"Consultant SI","telephone":"+33 1 55 00 00 02"}]`

func (oc *orgContext) registerUser(login string, id uuid.UUID) {
	if oc.usersByLogin == nil {
		oc.usersByLogin = make(map[string]uuid.UUID)
	}
	oc.usersByLogin[login] = id
}

func (oc *orgContext) userID(login string) uuid.UUID {
	if oc.usersByLogin == nil {
		return uuid.Nil
	}
	return oc.usersByLogin[login]
}

func (r *Runner) ensureApplication3(ctx context.Context, tenant kernel.TenantID) error {
	exists, err := r.rowExists(ctx, `SELECT EXISTS(SELECT 1 FROM org.applications WHERE id = $1)`, DemoApp3ID)
	if err != nil || exists {
		return err
	}
	return r.deps.OrgRepo.SaveApplication(ctx, orgdomain.Application{
		ID:        DemoApp3ID,
		TenantID:  tenant,
		ServiceID: DemoServiceID,
		Libelle:   DemoApp3Label,
	})
}

func (r *Runner) ensureEquipeNamed(ctx context.Context, tenant kernel.TenantID, equipeID, appID uuid.UUID, label string, managerID uuid.UUID) error {
	exists, err := r.rowExists(ctx, `SELECT EXISTS(SELECT 1 FROM org.equipes WHERE id = $1)`, equipeID)
	if err != nil || exists {
		return err
	}
	_, err = r.deps.Pool.Exec(ctx, `
		INSERT INTO org.equipes (id, tenant_id, application_id, libelle, responsable_id)
		VALUES ($1, $2, $3, $4, $5)
	`, equipeID, tenant.UUID(), appID, label, managerID)
	return err
}

func (r *Runner) seedExtendedTeamData(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	if err := r.seedExtendedCRA(ctx, tenant, oc); err != nil {
		return err
	}
	if err := r.seedExtendedConges(ctx, tenant, oc); err != nil {
		return err
	}
	if err := r.seedExtendedBudget(ctx, tenant, oc); err != nil {
		return err
	}
	if err := r.seedExtendedTMA(ctx, tenant, oc); err != nil {
		return err
	}
	if err := r.seedExtendedPublicsite(ctx, oc); err != nil {
		return err
	}
	return nil
}

func (r *Runner) seedExtendedCRA(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	now := time.Now().UTC()
	prevMonthKey, err := cradomain.ParseMonth(now.AddDate(0, -1, 0).Format("2006-01"))
	if err != nil {
		return err
	}
	currMonthKey, err := cradomain.ParseMonth(now.Format("2006-01"))
	if err != nil {
		return err
	}

	type craMemberSpec struct {
		login      string
		appID      uuid.UUID
		clientName string
		mission    string
		weeksCurr  []craWeekSpec
		weeksPrev  []craWeekSpec
		finalize   bool
	}

	specs := []craMemberSpec{
		{
			login: ChefDevLogin, appID: oc.appID, clientName: DemoClientName, mission: DemoAppLabel,
			weeksCurr: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2, 3}, submit: true}, {number: 2, dayOffsets: []int{8, 9}}},
			weeksPrev: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2}}, {number: 2, dayOffsets: []int{8, 9, 10}}},
			finalize: true,
		},
		{
			login: Collab3Login, appID: oc.appID, clientName: DemoClientName, mission: DemoAppLabel,
			weeksCurr: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2, 3, 4}, submit: true}, {number: 2, dayOffsets: []int{8, 9, 10}}},
			weeksPrev: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2, 3}}, {number: 2, dayOffsets: []int{8, 9}}},
			finalize: true,
		},
		{
			login: CollabQALogin, appID: oc.appID, clientName: DemoClientName, mission: "Recette " + DemoAppLabel,
			weeksCurr: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2}, submit: true}, {number: 2, dayOffsets: []int{8, 9}}},
		},
		{
			login: CollabIntegLogin, appID: oc.app3ID, clientName: DemoClient3Name, mission: DemoApp3Label,
			weeksCurr: []craWeekSpec{{number: 1, dayOffsets: []int{2, 3, 4}, submit: true}, {number: 2, dayOffsets: []int{9, 10}}},
			weeksPrev: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2, 3, 4}}, {number: 2, dayOffsets: []int{8, 9, 10}}},
			finalize: true,
		},
		{
			login: Presta2Login, appID: oc.app2ID, clientName: DemoClient2Name, mission: DemoApp2Label,
			weeksCurr: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2, 3}, submit: true}, {number: 2, dayOffsets: []int{8, 9, 10}}},
			weeksPrev: []craWeekSpec{{number: 1, dayOffsets: []int{2, 3}}, {number: 2, dayOffsets: []int{9, 10}}},
			finalize: true,
		},
		{
			login: Presta3Login, appID: oc.app3ID, clientName: DemoClient3Name, mission: DemoApp3Label,
			weeksCurr: []craWeekSpec{{number: 1, dayOffsets: []int{1, 2, 3, 4}, submit: true}, {number: 2, dayOffsets: []int{8, 9}}},
		},
		{
			login: PrestaIntegLogin, appID: oc.app3ID, clientName: DemoClient3Name, mission: "ETL " + DemoApp3Label,
			weeksCurr: []craWeekSpec{{number: 1, dayOffsets: []int{3, 4, 5}, submit: true}, {number: 2, dayOffsets: []int{10, 11}}},
		},
	}

	for _, spec := range specs {
		userID := oc.userID(spec.login)
		if userID == uuid.Nil {
			continue
		}
		if len(spec.weeksPrev) > 0 {
			if err := r.seedTimesheet(ctx, tenant, userID, spec.appID, prevMonthKey, spec.clientName, spec.mission, spec.weeksPrev, oc.managerID, spec.finalize); err != nil {
				return err
			}
		}
		if err := r.seedTimesheet(ctx, tenant, userID, spec.appID, currMonthKey, spec.clientName, spec.mission, spec.weeksCurr, oc.managerID, false); err != nil {
			return err
		}
	}

	log.Println("seed: CRA équipe élargie (7 collaborateurs + prestataires) alimenté")
	return nil
}

func (r *Runner) seedExtendedConges(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	type balanceSpec struct {
		login                    string
		leaveType                congesdomain.LeaveType
		acquired, taken, remaining float64
	}
	balances := []balanceSpec{
		{ChefDevLogin, congesdomain.LeaveTypeCongesPayes, 25, 4, 21},
		{ChefDevLogin, congesdomain.LeaveTypeRTT, 10, 2, 8},
		{Collab3Login, congesdomain.LeaveTypeCongesPayes, 25, 1, 24},
		{CollabQALogin, congesdomain.LeaveTypeCongesPayes, 25, 3, 22},
		{CollabQALogin, congesdomain.LeaveTypeRTT, 8, 0, 8},
		{CollabIntegLogin, congesdomain.LeaveTypeCongesPayes, 25, 6, 19},
		{Presta2Login, congesdomain.LeaveTypeCongesPayes, 0, 0, 0},
		{Presta3Login, congesdomain.LeaveTypeCongesPayes, 0, 0, 0},
		{Commercial2Login, congesdomain.LeaveTypeCongesPayes, 25, 1, 24},
		{Commercial3Login, congesdomain.LeaveTypeCongesPayes, 25, 0, 25},
	}

	for _, b := range balances {
		userID := oc.userID(b.login)
		if userID == uuid.Nil {
			continue
		}
		if err := r.ensureLeaveBalance(ctx, tenant, userID, b.leaveType, b.acquired, b.taken, b.remaining); err != nil {
			return err
		}
	}

	type leaveSpec struct {
		login   string
		leaveType congesdomain.LeaveType
		from, to time.Time
		motif   string
		approve bool
	}
	now := time.Now().UTC()
	leaves := []leaveSpec{
		{Collab3Login, congesdomain.LeaveTypeCongesPayes, nextMonday(now.AddDate(0, 0, 5)), nextMonday(now.AddDate(0, 0, 5)).AddDate(0, 0, 2), "Onboarding client Initech (demo)", false},
		{CollabQALogin, congesdomain.LeaveTypeRTT, nextMonday(now.AddDate(0, 0, 12)), nextMonday(now.AddDate(0, 0, 12)), "RTT QA recette (demo)", false},
		{CollabIntegLogin, congesdomain.LeaveTypeCongesPayes, now.AddDate(0, 0, -14).Truncate(24 * time.Hour), now.AddDate(0, 0, -12).Truncate(24 * time.Hour), "Mission Initech onsite (demo)", true},
		{ChefDevLogin, congesdomain.LeaveTypeCongesPayes, nextMonday(now.AddDate(0, 0, 35)), nextMonday(now.AddDate(0, 0, 35)).AddDate(0, 0, 1), "Congés chef équipe (demo)", false},
		{Commercial2Login, congesdomain.LeaveTypeCongesPayes, nextMonday(now.AddDate(0, 0, 18)), nextMonday(now.AddDate(0, 0, 18)), "RDV client Lyon (demo)", false},
		{Commercial3Login, congesdomain.LeaveTypeCongesPayes, nextMonday(now.AddDate(0, 0, 25)), nextMonday(now.AddDate(0, 0, 25)).AddDate(0, 0, 2), "Salon ESN Bordeaux (demo)", false},
	}

	for _, sc := range leaves {
		userID := oc.userID(sc.login)
		if userID == uuid.Nil {
			continue
		}
		exists, err := r.leaveExists(ctx, tenant, userID, sc.motif)
		if err != nil || exists {
			if err != nil {
				return err
			}
			continue
		}
		req, err := r.deps.Leaves.Request(ctx, congesports.RequestLeaveCommand{
			TenantID: tenant, UserID: userID, Type: sc.leaveType,
			From: sc.from, To: sc.to, Motif: sc.motif,
		})
		if err != nil {
			return err
		}
		if sc.approve {
			_, err = r.deps.Pool.Exec(ctx, `
				UPDATE conges.leave_requests SET status = 'valide', decided_by = $3, decided_at = NOW()
				WHERE id = $1 AND tenant_id = $2
			`, req.ID, tenant.UUID(), oc.managerID)
			if err != nil {
				return err
			}
		}
	}

	log.Println("seed: congés équipe élargie alimentés")
	return nil
}

func (r *Runner) seedExtendedTMA(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	specs := []struct {
		appID    uuid.UUID
		login    string
		subject  string
		assignTo string
		assign   bool
		analyze  bool
		resolve  bool
	}{
		{oc.appID, ClientMOALogin, "Évolution tableau de bord MOA ACME", Collab3Login, true, true, false},
		{oc.appID, ClientDSILogin, "Export audit logs Globex — accès refusé", CollabIntegLogin, true, true, false},
		{oc.app2ID, ClientPMOLogin, "Planning release Q3 Globex", Presta2Login, true, true, false},
		{oc.app3ID, ClientMOALogin, "Connecteur SAP Initech — spec manquante", "", false, false, false},
		{oc.app3ID, CollabIntegLogin, "Pipeline ETL — échec chargement nocturne", Presta3Login, true, true, true},
		{oc.appID, CollabQALogin, "Non-régression SSO après patch", CollabQALogin, true, true, false},
		{oc.app2ID, Presta2Login, "API REST Globex — pagination incorrecte", Presta2Login, true, true, false},
		{oc.app3ID, ChefDevLogin, "Revue architecture Data Hub", CollabIntegLogin, true, true, false},
		{oc.appID, Commercial2Login, "Demande POC ACME — module budget", "", false, false, false},
		{oc.app3ID, Commercial3Login, "Propale Initech — atelier cadrage", "", false, false, false},
		{oc.app2ID, PrestaIntegLogin, "Mapping données legacy Globex", PrestaIntegLogin, true, true, false},
		{oc.appID, ClientDSILogin, "Notification SLA portail — délai dépassé", Collab2Login, true, true, true},
	}

	demands, err := r.deps.TMA.List(ctx, tenant, tmaports.ExportFilter{})
	if err != nil {
		return err
	}
	existing := make(map[string]bool, len(demands))
	for _, d := range demands {
		existing[d.Subject] = true
	}

	for _, spec := range specs {
		if existing[spec.subject] {
			continue
		}
		authorID := oc.userID(spec.login)
		if authorID == uuid.Nil {
			continue
		}
		demand, err := r.deps.TMA.CreateDemand(ctx, tmaports.CreateDemandCommand{
			TenantID: tenant, ApplicationID: spec.appID, AuthorID: authorID, Subject: spec.subject,
		})
		if err != nil {
			return err
		}
		assigneeID := assigneeForApp(oc, spec.appID)
		if spec.assignTo != "" {
			assigneeID = oc.userID(spec.assignTo)
		}
		if spec.assign && assigneeID != uuid.Nil {
			if err := r.deps.TMA.Assign(ctx, tmaports.AssignCommand{
				TenantID: tenant, ID: demand.ID, AssigneeID: assigneeID, ActorID: oc.managerID,
			}); err != nil {
				return err
			}
		}
		if spec.analyze {
			if err := r.deps.TMA.AddAnalysis(ctx, tmaports.AnalysisCommand{
				TenantID: tenant, DemandID: demand.ID,
				Functional:   "Analyse fonctionnelle demo — reproduction validée.",
				Technical:    "Correctif ou évolution estimée sur sprint en cours.",
				Risks:        "Impact modéré sur le périmètre mission.",
				TestScenario: "Parcours nominal + cas limites métier.",
			}); err != nil {
				return err
			}
		}
		if spec.resolve {
			if err := r.deps.TMA.Resolve(ctx, tenant, demand.ID, assigneeID); err != nil {
				return err
			}
		}
	}

	log.Println("seed: TMA équipe élargie (12 demandes clients/prestataires/commerciaux) créées")
	return nil
}

func (r *Runner) seedExtendedBudget(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	if oc.app3ID == uuid.Nil {
		return nil
	}
	budgetID, err := r.ensureBudget(ctx, tenant, oc.app3ID, 60, 300, 6000000)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewPeriod(periodStart, periodStart.AddDate(0, 1, -1))
	if err != nil {
		return err
	}
	if _, err := r.deps.Budget.RecomputeConsumption(ctx, tenant, budgetID, period); err != nil {
		return err
	}

	if _, err := r.deps.Pool.Exec(ctx, `
		UPDATE org.applications
		SET proprietaire = $3, chef_utilisateur_id = $4, uo_activee = TRUE, mode_facturation = 'temps_passe'
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), oc.app3ID, DemoClient3Name, oc.userID(CollabIntegLogin)); err != nil {
		return err
	}

	log.Println("seed: budget mission Initech créé")
	return nil
}

func (r *Runner) seedExtendedPublicsite(ctx context.Context, oc orgContext) error {
	if r.deps.PublicSlots == nil {
		return nil
	}
	start := nextWeekdayAt(time.Now().UTC(), time.Wednesday, 14, 0)
	for i, login := range []string{Commercial2Login, Commercial3Login} {
		commercialID := oc.userID(login)
		if commercialID == uuid.Nil {
			continue
		}
		slotStart := start.Add(time.Duration(i*24) * time.Hour)
		if err := r.deps.PublicSlots.SeedSlot(ctx, commercialID, slotStart, slotStart.Add(30*time.Minute)); err != nil {
			return err
		}
	}
	log.Println("seed: créneaux publicsite commerciaux supplémentaires")
	return nil
}
