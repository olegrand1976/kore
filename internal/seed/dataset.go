package seed

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	budgetdomain "github.com/kore/kore/internal/modules/budget/domain"
	budgetports "github.com/kore/kore/internal/modules/budget/ports"
	congesdomain "github.com/kore/kore/internal/modules/conges/domain"
	congesports "github.com/kore/kore/internal/modules/conges/ports"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

type craWeekSpec struct {
	number     cradomain.WeekNumber
	dayOffsets []int
	submit     bool
}

type tmaAnalysisContent struct {
	functional   string
	technical    string
	risks        string
	testScenario string
}

type tmaDemandSpec struct {
	appID    uuid.UUID
	authorID uuid.UUID
	subject  string
	assign   bool
	assignee uuid.UUID
	analyze  bool
	analysis tmaAnalysisContent
	resolve  bool
}

func (r *Runner) seedCRAData(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	now := time.Now().UTC()
	prevMonth := now.AddDate(0, -1, 0)
	prevMonthKey, err := cradomain.ParseMonth(prevMonth.Format("2006-01"))
	if err != nil {
		return err
	}
	currMonthKey, err := cradomain.ParseMonth(now.Format("2006-01"))
	if err != nil {
		return err
	}

	specs := []struct {
		userID     uuid.UUID
		appID      uuid.UUID
		month      cradomain.Month
		clientName string
		mission    string
		weeks      []craWeekSpec
		finalize   bool
	}{
		{
			userID: oc.collabID, appID: oc.appID, month: prevMonthKey,
			clientName: DemoClientName, mission: DemoAppLabel,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{1, 2, 3}},
				{number: 2, dayOffsets: []int{8, 9}},
			},
			finalize: true,
		},
		{
			userID: oc.collabID, appID: oc.appID, month: currMonthKey,
			clientName: DemoClientName, mission: DemoAppLabel,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{1, 2, 3}, submit: true},
				{number: 2, dayOffsets: []int{8, 9, 10}},
				{number: 3, dayOffsets: []int{15, 16}},
			},
		},
		{
			userID: oc.collab2ID, appID: oc.appID, month: prevMonthKey,
			clientName: DemoClientName, mission: DemoAppLabel,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{1, 2, 3, 4}},
				{number: 2, dayOffsets: []int{8, 9, 10}},
				{number: 3, dayOffsets: []int{15, 16, 17}},
			},
			finalize: true,
		},
		{
			userID: oc.collab2ID, appID: oc.appID, month: currMonthKey,
			clientName: DemoClientName, mission: DemoAppLabel,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{1, 2}, submit: true},
				{number: 2, dayOffsets: []int{8, 9}},
			},
		},
		{
			userID: oc.managerID, appID: oc.appID, month: currMonthKey,
			clientName: DemoClientName, mission: "Pilotage " + DemoAppLabel,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{1, 2}, submit: true},
				{number: 2, dayOffsets: []int{8}},
			},
		},
		{
			userID: oc.managerID, appID: oc.appID, month: prevMonthKey,
			clientName: DemoClientName, mission: "Pilotage " + DemoAppLabel,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{1, 2, 3}},
				{number: 2, dayOffsets: []int{8, 9}},
			},
			finalize: true,
		},
		{
			userID: oc.prestaID, appID: oc.app2ID, month: prevMonthKey,
			clientName: DemoClient2Name, mission: DemoApp2Label,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{1, 2, 3, 4}},
				{number: 2, dayOffsets: []int{8, 9, 10}},
			},
			finalize: true,
		},
		{
			userID: oc.prestaID, appID: oc.app2ID, month: currMonthKey,
			clientName: DemoClient2Name, mission: DemoApp2Label,
			weeks: []craWeekSpec{
				{number: 1, dayOffsets: []int{2, 3, 4}, submit: true},
				{number: 2, dayOffsets: []int{9, 10, 11}},
			},
		},
	}

	for _, spec := range specs {
		if err := r.seedTimesheet(ctx, tenant, spec.userID, spec.appID, spec.month, spec.clientName, spec.mission, spec.weeks, oc.managerID, spec.finalize); err != nil {
			return err
		}
	}

	log.Println("seed: CRA (6 profils actifs, 2 mois, internes + prestataire + manager) alimenté")
	return nil
}

func (r *Runner) seedTimesheet(
	ctx context.Context,
	tenant kernel.TenantID,
	userID, appID uuid.UUID,
	month cradomain.Month,
	clientName, mission string,
	weeks []craWeekSpec,
	managerID uuid.UUID,
	finalize bool,
) error {
	ts, err := r.deps.CRA.GetOrCreate(ctx, tenant, userID, month)
	if err != nil {
		return err
	}
	if finalize && ts.IsFinal() {
		return nil
	}
	if !finalize && len(ts.Weeks) > 0 && len(weeks) > 0 {
		if w, _ := ts.Week(weeks[0].number); w != nil && len(w.Lines) > 0 {
			return nil
		}
	}

	monthTime, _ := time.Parse("2006-01", string(month))
	monthStart := time.Date(monthTime.Year(), monthTime.Month(), 1, 0, 0, 0, 0, time.UTC)
	duration, err := kernel.NewDuration(420)
	if err != nil {
		return err
	}

	for _, spec := range weeks {
		lines := make([]cradomain.TimeLine, 0, len(spec.dayOffsets))
		for _, offset := range spec.dayOffsets {
			lines = append(lines, cradomain.TimeLine{
				Source:   cradomain.SourceRef{Type: "mission", ID: appID.String()},
				Day:      monthStart.AddDate(0, 0, offset),
				Duration: duration,
				Comment:  mission,
			})
		}
		ts, err = r.deps.CRA.SaveWeek(ctx, craports.SaveWeekCommand{
			TenantID:    tenant,
			TimesheetID: ts.ID,
			WeekNumber:  spec.number,
			Lines:       lines,
		})
		if err != nil {
			return err
		}
		if spec.submit {
			if err := r.deps.CRA.SubmitWeek(ctx, craports.SubmitWeekCommand{
				TenantID:    tenant,
				TimesheetID: ts.ID,
				WeekNumber:  spec.number,
				UserID:      userID,
			}); err != nil {
				return err
			}
		}
	}

	if err := r.deps.CRA.CompleteCommercialInfo(ctx, craports.CommercialCommand{
		TenantID:    tenant,
		TimesheetID: ts.ID,
		Info: cradomain.CommercialInfo{
			Client:  clientName,
			Mission: mission,
		},
	}); err != nil {
		return err
	}
	if finalize {
		ts, err = r.deps.CRA.GetByID(ctx, tenant, ts.ID)
		if err != nil {
			return err
		}
		for _, week := range ts.Weeks {
			if week.SubmittedAt != nil || len(week.Lines) == 0 {
				continue
			}
			if err := r.deps.CRA.SubmitWeek(ctx, craports.SubmitWeekCommand{
				TenantID:    tenant,
				TimesheetID: ts.ID,
				WeekNumber:  week.WeekNumber,
				UserID:      userID,
			}); err != nil {
				return err
			}
		}
		_, err := r.deps.CRA.ValidateFinal(ctx, craports.ManagerValidateCommand{
			TenantID:    tenant,
			TimesheetID: ts.ID,
			ManagerID:   managerID,
		})
		return err
	}
	return nil
}

func (r *Runner) seedCongesData(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	balances := []struct {
		userID                     uuid.UUID
		leaveType                  congesdomain.LeaveType
		acquired, taken, remaining float64
	}{
		{oc.adminID, congesdomain.LeaveTypeCongesPayes, 25, 0, 25},
		{oc.managerID, congesdomain.LeaveTypeCongesPayes, 30, 8, 22},
		{oc.managerID, congesdomain.LeaveTypeRTT, 12, 4, 8},
		{oc.commercialID, congesdomain.LeaveTypeCongesPayes, 25, 3, 22},
		{oc.commercialID, congesdomain.LeaveTypeRTT, 8, 0, 8},
		{oc.collabID, congesdomain.LeaveTypeCongesPayes, 25, 5, 20},
		{oc.collabID, congesdomain.LeaveTypeRTT, 10, 2, 8},
		{oc.collab2ID, congesdomain.LeaveTypeCongesPayes, 25, 2, 23},
		{oc.collab2ID, congesdomain.LeaveTypeRTT, 10, 1, 9},
		{oc.collab2ID, congesdomain.LeaveTypeMaladie, 0, 2, 0},
		{oc.prestaID, congesdomain.LeaveTypeCongesPayes, 0, 0, 0},
	}
	for _, b := range balances {
		if err := r.ensureLeaveBalance(ctx, tenant, b.userID, b.leaveType, b.acquired, b.taken, b.remaining); err != nil {
			return err
		}
	}

	scenarios := []struct {
		userID    uuid.UUID
		leaveType congesdomain.LeaveType
		from, to  time.Time
		motif     string
		approve   bool
		reject    bool
	}{
		{
			userID:    oc.collabID,
			leaveType: congesdomain.LeaveTypeCongesPayes,
			from:      time.Now().UTC().AddDate(0, 0, -20).Truncate(24 * time.Hour),
			to:        time.Now().UTC().AddDate(0, 0, -18).Truncate(24 * time.Hour),
			motif:     "Congés validés (demo)",
			approve:   true,
		},
		{
			userID:    oc.collabID,
			leaveType: congesdomain.LeaveTypeCongesPayes,
			from:      nextMonday(time.Now().UTC().AddDate(0, 0, 14)),
			to:        nextMonday(time.Now().UTC().AddDate(0, 0, 14)).AddDate(0, 0, 2),
			motif:     "Congés été (demo)",
		},
		{
			userID:    oc.collabID,
			leaveType: congesdomain.LeaveTypeRTT,
			from:      nextMonday(time.Now().UTC().AddDate(0, 0, 21)),
			to:        nextMonday(time.Now().UTC().AddDate(0, 0, 21)),
			motif:     "RTT pont (demo)",
		},
		{
			userID:    oc.collab2ID,
			leaveType: congesdomain.LeaveTypeRTT,
			from:      nextMonday(time.Now().UTC().AddDate(0, 0, 7)),
			to:        nextMonday(time.Now().UTC().AddDate(0, 0, 7)),
			motif:     "RTT pont (demo)",
		},
		{
			userID:    oc.collab2ID,
			leaveType: congesdomain.LeaveTypeMaladie,
			from:      time.Now().UTC().AddDate(0, 0, -10).Truncate(24 * time.Hour),
			to:        time.Now().UTC().AddDate(0, 0, -9).Truncate(24 * time.Hour),
			motif:     "Arrêt maladie (demo)",
			approve:   true,
		},
		{
			userID:    oc.managerID,
			leaveType: congesdomain.LeaveTypeCongesPayes,
			from:      time.Now().UTC().AddDate(0, 0, -35).Truncate(24 * time.Hour),
			to:        time.Now().UTC().AddDate(0, 0, -33).Truncate(24 * time.Hour),
			motif:     "Formation managériale (demo)",
			approve:   true,
		},
		{
			userID:    oc.managerID,
			leaveType: congesdomain.LeaveTypeRTT,
			from:      nextMonday(time.Now().UTC().AddDate(0, 0, 28)),
			to:        nextMonday(time.Now().UTC().AddDate(0, 0, 28)),
			motif:     "RTT responsable (demo)",
		},
		{
			userID:    oc.commercialID,
			leaveType: congesdomain.LeaveTypeCongesPayes,
			from:      nextMonday(time.Now().UTC().AddDate(0, 0, 10)),
			to:        nextMonday(time.Now().UTC().AddDate(0, 0, 10)).AddDate(0, 0, 1),
			motif:     "Salon Tech Paris (demo)",
		},
		{
			userID:    oc.adminID,
			leaveType: congesdomain.LeaveTypeCongesPayes,
			from:      nextMonday(time.Now().UTC().AddDate(0, 0, 42)),
			to:        nextMonday(time.Now().UTC().AddDate(0, 0, 42)).AddDate(0, 0, 4),
			motif:     "Congés admin (demo)",
		},
	}

	for _, sc := range scenarios {
		exists, err := r.leaveExists(ctx, tenant, sc.userID, sc.motif)
		if err != nil || exists {
			if err != nil {
				return err
			}
			continue
		}
		req, err := r.deps.Leaves.Request(ctx, congesports.RequestLeaveCommand{
			TenantID: tenant,
			UserID:   sc.userID,
			Type:     sc.leaveType,
			From:     sc.from,
			To:       sc.to,
			Motif:    sc.motif,
		})
		if err != nil {
			return err
		}
		switch {
		case sc.approve:
			_, err = r.deps.Pool.Exec(ctx, `
				UPDATE conges.leave_requests
				SET status = 'valide', decided_by = $3, decided_at = NOW()
				WHERE id = $1 AND tenant_id = $2
			`, req.ID, tenant.UUID(), oc.managerID)
			if err != nil {
				return err
			}
		case sc.reject:
			_, err = r.deps.Pool.Exec(ctx, `
				UPDATE conges.leave_requests
				SET status = 'refuse', decided_by = $3, decided_at = NOW()
				WHERE id = $1 AND tenant_id = $2
			`, req.ID, tenant.UUID(), oc.managerID)
			if err != nil {
				return err
			}
		}
	}

	log.Println("seed: congés (soldes 7 profils + demandes validées/en attente/maladie) alimentés")
	return nil
}

func (r *Runner) leaveExists(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, motif string) (bool, error) {
	var count int
	err := r.deps.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM conges.leave_requests
		WHERE tenant_id = $1 AND user_id = $2 AND motif = $3
	`, tenant.UUID(), userID, motif).Scan(&count)
	return count > 0, err
}

func (r *Runner) ensureLeaveBalance(
	ctx context.Context,
	tenant kernel.TenantID,
	userID uuid.UUID,
	leaveType congesdomain.LeaveType,
	acquired, taken, remaining float64,
) error {
	_, err := r.deps.Pool.Exec(ctx, `
		INSERT INTO conges.leave_balances (id, tenant_id, user_id, type, acquired, taken, remaining)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, user_id, type) DO UPDATE
		SET acquired = EXCLUDED.acquired, taken = EXCLUDED.taken, remaining = EXCLUDED.remaining
	`, uuid.New(), tenant.UUID(), userID, string(leaveType), acquired, taken, remaining)
	return err
}

func (r *Runner) seedBudgetData(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	budgets := []struct {
		appID    uuid.UUID
		days, uo float64
		amount   int64
	}{
		{oc.appID, 120, 600, 12000000},
		{oc.app2ID, 80, 400, 8000000},
	}
	var primaryBudgetID uuid.UUID
	var secondaryBudgetID uuid.UUID
	for _, spec := range budgets {
		budgetID, err := r.ensureBudget(ctx, tenant, spec.appID, spec.days, spec.uo, spec.amount)
		if err != nil {
			return err
		}
		if spec.appID == oc.appID {
			primaryBudgetID = budgetID
		}
		if spec.appID == oc.app2ID {
			secondaryBudgetID = budgetID
		}
	}

	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1)
	period, err := kernel.NewPeriod(periodStart, periodEnd)
	if err != nil {
		return err
	}
	for _, budgetID := range []uuid.UUID{primaryBudgetID, secondaryBudgetID} {
		if budgetID == uuid.Nil {
			continue
		}
		if _, err := r.deps.Budget.RecomputeConsumption(ctx, tenant, budgetID, period); err != nil {
			return err
		}
	}

	log.Println("seed: budgets (2 missions, consommation recalculée) alimentés")
	return nil
}

func (r *Runner) enrichBudgetEstimates(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	budgets, err := r.deps.Budget.List(ctx, tenant)
	if err != nil {
		return err
	}
	var primaryBudgetID uuid.UUID
	for _, b := range budgets {
		if b.ApplicationID == oc.appID && b.Type == budgetdomain.BudgetTypeDefault {
			primaryBudgetID = b.ID
			break
		}
	}
	if primaryBudgetID == uuid.Nil {
		return nil
	}

	demands, err := r.deps.TMA.List(ctx, tenant, tmaports.ExportFilter{})
	if err != nil {
		return err
	}

	estimateTargets := map[string]struct {
		effortUO   float64
		effortDays float64
	}{
		"Lenteur écran de saisie CRA":          {3, 1.5},
		"Régression export Excel budget":       {5, 2.5},
		"Évolution reporting financier Globex": {8, 4},
	}

	for subject, target := range estimateTargets {
		var demandID uuid.UUID
		for _, d := range demands {
			if d.Subject == subject {
				demandID = d.ID
				break
			}
		}
		if demandID == uuid.Nil {
			continue
		}
		var count int
		if err := r.deps.Pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM budget.estimates WHERE tenant_id = $1 AND demand_id = $2
		`, tenant.UUID(), demandID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if _, err := r.deps.Budget.AddEstimate(ctx, budgetports.EstimateCommand{
			TenantID:   tenant,
			BudgetID:   primaryBudgetID,
			DemandID:   demandID,
			EffortUO:   target.effortUO,
			EffortDays: target.effortDays,
		}); err != nil {
			return err
		}
	}

	log.Println("seed: estimations budget liées aux demandes TMA")
	return nil
}

func (r *Runner) ensureBudget(
	ctx context.Context,
	tenant kernel.TenantID,
	appID uuid.UUID,
	days, uo float64,
	amount int64,
) (uuid.UUID, error) {
	budgets, err := r.deps.Budget.List(ctx, tenant)
	if err != nil {
		return uuid.Nil, err
	}
	for _, b := range budgets {
		if b.ApplicationID == appID && b.Type == budgetdomain.BudgetTypeDefault {
			return b.ID, nil
		}
	}
	created, err := r.deps.Budget.CreateBudget(ctx, budgetports.CreateBudgetCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		Type:          budgetdomain.BudgetTypeDefault,
		PlannedDays:   days,
		PlannedUO:     uo,
		PlannedAmount: amount,
		Currency:      "EUR",
	})
	if err != nil {
		return uuid.Nil, err
	}
	return created.ID, nil
}

func (r *Runner) seedTMAData(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	defaultAnalysis := tmaAnalysisContent{
		functional:   "Reproduction confirmée en recette sur parcours nominal.",
		technical:    "Correctif prévu sur le prochain sprint — impact limité au module concerné.",
		risks:        "Impact modéré, pas de régression attendue sur les flux adjacents.",
		testScenario: "Parcours nominal + cas limite + non-régression export.",
	}

	specs := []tmaDemandSpec{
		{oc.appID, oc.collabID, "Erreur export PDF factures", false, uuid.Nil, false, tmaAnalysisContent{}, false},
		{
			oc.appID, oc.collabID, "Lenteur écran de saisie CRA", true, oc.collabID, true,
			tmaAnalysisContent{
				functional:   "Temps de rendu > 3s lors de la saisie hebdomadaire avec plus de 20 lignes.",
				technical:    "Requête N+1 sur les missions — index composite à ajouter + cache court.",
				risks:        "Charge DB en pic de fin de mois.",
				testScenario: "Saisie 30 lignes, navigation semaine suivante, export PDF.",
			},
			false,
		},
		{
			oc.appID, oc.collab2ID, "Badge SSO mobile — accès refusé", true, oc.collab2ID, true,
			tmaAnalysisContent{
				functional:   "Connexion SSO OK desktop, refus token sur app mobile iOS 17.",
				technical:    "Durée de vie cookie incompatible SameSite=None — aligner config BFF.",
				risks:        "Régression auth desktop si mal configuré.",
				testScenario: "Login iOS Safari + Android Chrome + desktop Firefox.",
			},
			true,
		},
		{
			oc.app2ID, oc.prestaID, "Évolution reporting financier Globex", true, oc.prestaID, false,
			tmaAnalysisContent{}, false,
		},
		{oc.appID, oc.adminID, "Paramétrage notifications globales", false, uuid.Nil, false, tmaAnalysisContent{}, false},
		{
			oc.appID, oc.managerID, "Validation workflow congés — délai", true, oc.collabID, true,
			tmaAnalysisContent{
				functional:   "Les managers ne reçoivent pas de rappel J+2 sur demandes en attente.",
				technical:    "Règle notification leave-requested mal paramétrée — destinataires vides.",
				risks:        "Spam si fréquence mal calibrée.",
				testScenario: "Créer demande, vérifier notification J+0 et J+2.",
			},
			false,
		},
		{
			oc.appID, oc.clientUserID, "Accès rapport mensuel ACME", true, oc.collabID, true,
			tmaAnalysisContent{
				functional:   "Le contact client ne voit pas le rapport PDF du mois précédent.",
				technical:    "Filtrage RBAC côté BFF exclut le profil Client sur endpoint export.",
				risks:        "Fuite de données si filtre trop permissif.",
				testScenario: "Login CLI_contact, téléchargement rapport M-1.",
			},
			false,
		},
		{oc.app2ID, oc.commercialID, "Demande démo personnalisée Globex", false, uuid.Nil, false, tmaAnalysisContent{}, false},
		{
			oc.appID, oc.collab2ID, "Régression export Excel budget", true, oc.collab2ID, true,
			tmaAnalysisContent{
				functional:   "Colonnes UO manquantes depuis la dernière release.",
				technical:    "Template XLSX obsolète — regénérer à partir du modèle v2.",
				risks:        "Livraison client bloquée fin de mois.",
				testScenario: "Export budget mission ACME, comparer avec référence v1.",
			},
			false,
		},
		{oc.appID, oc.clientUserID, "Timeout session portail client", false, uuid.Nil, false, tmaAnalysisContent{}, false},
		{
			oc.app2ID, oc.managerID, "Incident batch nocturne Globex", true, oc.prestaID, true,
			tmaAnalysisContent{
				functional:   "Batch consolidation échoue à 02:00 avec erreur timeout.",
				technical:    "Augmenter timeout PostgreSQL + partitionner le traitement.",
				risks:        "Données financières incomplètes le matin.",
				testScenario: "Relance batch staging, vérifier logs et KPI dashboard.",
			},
			false,
		},
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
		demand, err := r.deps.TMA.CreateDemand(ctx, tmaports.CreateDemandCommand{
			TenantID:         tenant,
			ApplicationID:    spec.appID,
			AuthorID:         spec.authorID,
			Subject:          spec.subject,
			RequiresChefGate: false,
		})
		if err != nil {
			return err
		}
		assignee := assigneeForApp(oc, spec.appID)
		if spec.assignee != uuid.Nil {
			assignee = spec.assignee
		}
		if spec.assign {
			if err := r.deps.TMA.Assign(ctx, tmaports.AssignCommand{
				TenantID:   tenant,
				ID:         demand.ID,
				AssigneeID: assignee,
				ActorID:    oc.managerID,
			}); err != nil {
				return err
			}
		}
		analysis := spec.analysis
		if spec.analyze && analysis == (tmaAnalysisContent{}) {
			analysis = defaultAnalysis
		}
		if spec.analyze {
			if err := r.deps.TMA.AddAnalysis(ctx, tmaports.AnalysisCommand{
				TenantID:     tenant,
				DemandID:     demand.ID,
				Functional:   analysis.functional,
				Technical:    analysis.technical,
				Risks:        analysis.risks,
				TestScenario: analysis.testScenario,
			}); err != nil {
				return err
			}
		}
		if spec.resolve {
			if err := r.deps.TMA.Resolve(ctx, tenant, demand.ID, assignee); err != nil {
				return err
			}
		}
	}

	log.Println("seed: demandes TMA (12 sujets, tous profils auteurs) créées")
	return nil
}

func logDemoAccounts() {
	log.Println("seed: comptes demo disponibles (19 utilisateurs, mot de passe groupe : Collab123! / Presta123! / Client123! / Commercial123!)")
	log.Println("  — Direction & management —")
	log.Printf("    admin     %s / %s", AdminLogin, AdminPassword)
	log.Printf("    manager   %s / %s", ManagerLogin, ManagerPassword)
	log.Printf("    chef équipe %s / %s", ChefDevLogin, ChefDevPassword)
	log.Println("  — Équipe Dev & QA —")
	log.Printf("    %s / %s", CollabLogin, CollabPassword)
	log.Printf("    %s / %s", Collab2Login, Collab2Password)
	log.Printf("    %s / %s", Collab3Login, CollabPassword)
	log.Printf("    %s / %s", CollabQALogin, CollabPassword)
	log.Println("  — Équipe Data & Intégration —")
	log.Printf("    %s / %s", CollabIntegLogin, CollabPassword)
	log.Println("  — Prestataires ETT —")
	log.Printf("    %s / %s", PrestaLogin, PrestaPassword)
	log.Printf("    %s / %s", Presta2Login, PrestaPassword)
	log.Printf("    %s / %s", Presta3Login, PrestaPassword)
	log.Printf("    %s / %s", PrestaIntegLogin, PrestaPassword)
	log.Println("  — Contacts clients —")
	log.Printf("    %s / %s (ACME)", ClientUserLogin, ClientUserPass)
	log.Printf("    %s / %s (ACME MOA)", ClientMOALogin, ClientUserPass)
	log.Printf("    %s / %s (Globex DSI)", ClientDSILogin, ClientUserPass)
	log.Printf("    %s / %s (Globex PMO)", ClientPMOLogin, ClientUserPass)
	log.Println("  — Commerciaux —")
	log.Printf("    %s / %s", CommercialLogin, CommercialPass)
	log.Printf("    %s / %s", Commercial2Login, CommercialPass)
	log.Printf("    %s / %s", Commercial3Login, CommercialPass)
}
