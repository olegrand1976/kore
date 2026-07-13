package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	congesdomain "github.com/kore/kore/internal/modules/conges/domain"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
	wfdomain "github.com/kore/kore/internal/modules/workflow/domain"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/pkg/kernel"
)

type Service struct {
	repo     ports.Repository
	llm      ports.LLMProvider
	tma      ports.TMAReader
	cra      ports.CRAReader
	leaves   ports.LeaveReader
	workflow ports.WorkflowReader
	clock    ports.Clock
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

func NewService(
	repo ports.Repository,
	llm ports.LLMProvider,
	tma ports.TMAReader,
	cra ports.CRAReader,
	leaves ports.LeaveReader,
	workflow ports.WorkflowReader,
) *Service {
	return &Service{
		repo:     repo,
		llm:      llm,
		tma:      tma,
		cra:      cra,
		leaves:   leaves,
		workflow: workflow,
		clock:    systemClock{},
	}
}

func (s *Service) ensureAI(ctx context.Context, tenant kernel.TenantID, capability string) error {
	settings, err := s.repo.GetTenantSettings(ctx, tenant)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return domain.ErrAIDisabled
	}
	ok, err := s.repo.IsCapabilityEnabled(ctx, capability)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrCapabilityOff
	}
	return nil
}

func hashInput(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (s *Service) logRequest(ctx context.Context, log domain.RequestLog) (uuid.UUID, error) {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = s.clock.Now()
	}
	if err := s.repo.InsertRequestLog(ctx, log); err != nil {
		return uuid.Nil, err
	}
	return log.ID, nil
}

func (s *Service) SuggestAnalysisDraft(ctx context.Context, cmd ports.AnalysisDraftCommand) (ports.AnalysisDraftResult, error) {
	const capCode = "tma.analysis_draft"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.AnalysisDraftResult{}, err
	}
	subject := cmd.Subject
	if subject == "" && cmd.DemandID != uuid.Nil {
		d, err := s.tma.GetDemand(ctx, cmd.TenantID, cmd.DemandID)
		if err == nil {
			subject = d.Subject
		}
	}
	draft := buildAnalysisDraft(subject)
	out, _ := json.Marshal(draft)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID:       cmd.TenantID,
		UserID:         cmd.UserID,
		CapabilityCode: capCode,
		EntityType:     "tma_demand",
		EntityID:       ptrUUID(cmd.DemandID),
		InputHash:      hashInput(cmd),
		OutputJSON:     out,
		Model:          stub.ModelName,
		ExplainContext: map[string]any{"subject": subject, "capability": capCode},
	})
	if err != nil {
		return ports.AnalysisDraftResult{}, err
	}
	return ports.AnalysisDraftResult{Draft: draft, RequestID: reqID}, nil
}

func buildAnalysisDraft(subject string) domain.AnalysisDraft {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = "demande TMA"
	}
	return domain.AnalysisDraft{
		Functional:   fmt.Sprintf("Comportement attendu pour « %s » : décrire le parcours utilisateur impacté et la reproduction.", subject),
		Technical:    fmt.Sprintf("Investigation technique sur « %s » : identifier composant, logs et hypothèse de cause.", subject),
		Risks:        "Risque de régression sur modules adjacents ; valider jeux de données et déploiements récents.",
		TestScenario: "1. Reproduire le cas. 2. Valider le correctif. 3. Smoke test non-régression.",
	}
}

func (s *Service) ClassifyDemand(ctx context.Context, cmd ports.ClassifyDemandCommand) (ports.ClassifyResult, error) {
	const capCode = "tma.classify"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.ClassifyResult{}, err
	}
	category, confidence := stub.ClassifySubject(cmd.Subject)
	result := ports.ClassifyResult{Category: category, Confidence: confidence}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
		ExplainContext: map[string]any{"subject": cmd.Subject, "category": category},
	})
	if err != nil {
		return ports.ClassifyResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) FindSimilarDemands(ctx context.Context, cmd ports.SimilarDemandsCommand) ([]ports.SimilarDemand, error) {
	const capCode = "tma.similar"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return nil, err
	}
	limit := cmd.Limit
	if limit <= 0 {
		limit = 5
	}
	demands, err := s.tma.ListDemands(ctx, cmd.TenantID, true)
	if err != nil {
		return nil, err
	}
	subjectLower := strings.ToLower(strings.TrimSpace(cmd.Subject))
	var out []ports.SimilarDemand
	for _, d := range demands {
		if cmd.ApplicationID != nil && d.ApplicationID != *cmd.ApplicationID {
			continue
		}
		if d.Status != tmadomain.DemandStatusResolved {
			continue
		}
		score := similarityScore(subjectLower, strings.ToLower(d.Subject))
		if score < 0.2 {
			continue
		}
		out = append(out, ports.SimilarDemand{DemandID: d.ID, Subject: d.Subject, Score: score})
	}
	if len(out) > limit {
		out = out[:limit]
	}
	_, _ = s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		InputHash: hashInput(cmd), OutputJSON: mustJSON(out), Model: stub.ModelName,
	})
	return out, nil
}

func similarityScore(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1
	}
	if strings.Contains(b, a) || strings.Contains(a, b) {
		return 0.75
	}
	wordsA := strings.Fields(a)
	matches := 0
	for _, w := range wordsA {
		if len(w) > 3 && strings.Contains(b, w) {
			matches++
		}
	}
	if len(wordsA) == 0 {
		return 0
	}
	return float64(matches) / float64(len(wordsA))
}

func (s *Service) SuggestCraPrefill(ctx context.Context, cmd ports.CraPrefillCommand) (ports.CraPrefillResult, error) {
	const capCode = "cra.prefill"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.CraPrefillResult{}, err
	}
	ts, err := s.cra.GetTimesheetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return ports.CraPrefillResult{}, err
	}
	recent, _ := s.cra.ListRecentTimesheets(ctx, cmd.TenantID, cmd.UserID, 2)
	var lines []ports.PrefillLine
	for _, week := range ts.Weeks {
		if cmd.WeekNumber > 0 && int(week.WeekNumber) != cmd.WeekNumber {
			continue
		}
		for _, line := range week.Lines {
			if line.Duration.Minutes > 0 {
				lines = append(lines, ports.PrefillLine{
					Day:      line.Day.Format("2006-01-02"),
					Duration: float64(line.Duration.Minutes) / 60,
					Comment:  line.Comment,
				})
			}
		}
	}
	if len(lines) == 0 && len(recent) > 1 && len(recent[1].Weeks) > 0 {
		for _, line := range recent[1].Weeks[0].Lines {
			if line.Duration.Minutes > 0 {
				lines = append(lines, ports.PrefillLine{
					Day:      line.Day.Format("2006-01-02"),
					Duration: float64(line.Duration.Minutes) / 60,
					Comment:  "Suggestion basée sur période précédente",
				})
			}
		}
	}
	result := ports.CraPrefillResult{Lines: lines}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "timesheet", EntityID: ptrUUID(cmd.TimesheetID),
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.CraPrefillResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) ListCraAnomalies(ctx context.Context, cmd ports.CraAnomaliesCommand) ([]ports.CraAnomaly, error) {
	const capCode = "cra.anomalies"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return nil, err
	}
	ts, err := s.cra.GetTimesheetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return nil, err
	}
	var anomalies []ports.CraAnomaly
	for _, week := range ts.Weeks {
		dayTotals := map[string]float64{}
		daySet := map[string]bool{}
		for _, line := range week.Lines {
			key := line.Day.Format("2006-01-02")
			daySet[key] = true
			dayTotals[key] += float64(line.Duration.Minutes) / 60
		}
		for day, hours := range dayTotals {
			if hours > 8 {
				anomalies = append(anomalies, ports.CraAnomaly{
					Code: "DAY_CAPACITY", Message: "Heures supérieures à 8h sur la journée", Day: day,
				})
			}
		}
		if len(daySet) == 0 && ts.CanEdit() {
			anomalies = append(anomalies, ports.CraAnomaly{
				Code: "EMPTY_WEEK", Message: fmt.Sprintf("Semaine %d sans saisie", week.WeekNumber),
			})
		}
	}
	_, _ = s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "timesheet", EntityID: ptrUUID(cmd.TimesheetID),
		InputHash: hashInput(cmd), OutputJSON: mustJSON(anomalies), Model: "rules-v1",
	})
	return anomalies, nil
}

func (s *Service) EstimateBudgetEffort(ctx context.Context, cmd ports.BudgetEstimateCommand) (ports.BudgetEstimateResult, error) {
	const capCode = "budget.estimate"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.BudgetEstimateResult{}, err
	}
	d, err := s.tma.GetDemand(ctx, cmd.TenantID, cmd.DemandID)
	if err != nil {
		return ports.BudgetEstimateResult{}, err
	}
	days := 2.0
	uo := 4.0
	rationale := fmt.Sprintf("Estimation indicative pour « %s » (incident standard).", d.Subject)
	if analysis, err := s.tma.GetAnalysis(ctx, cmd.TenantID, cmd.DemandID); err == nil {
		if strings.Contains(strings.ToLower(analysis.Technical), "complex") {
			days = 5
			uo = 10
			rationale = "Analyse technique dense — effort majoré."
		}
	}
	result := ports.BudgetEstimateResult{EffortDays: days, EffortUO: uo, Rationale: rationale}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "tma_demand", EntityID: ptrUUID(cmd.DemandID),
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.BudgetEstimateResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) SuggestBudgetDemands(ctx context.Context, cmd ports.BudgetDemandSuggestCommand) ([]ports.DemandSuggestion, error) {
	const capCode = "budget.demand_suggest"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return nil, err
	}
	limit := cmd.Limit
	if limit <= 0 {
		limit = 10
	}
	demands, err := s.tma.ListDemands(ctx, cmd.TenantID, true)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(strings.TrimSpace(cmd.Query))
	var out []ports.DemandSuggestion
	for _, d := range demands {
		if d.Status == tmadomain.DemandStatusResolved {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(d.Subject), q) && !strings.Contains(d.ID.String(), q) {
			continue
		}
		out = append(out, ports.DemandSuggestion{
			DemandID: d.ID, Subject: d.Subject, Status: string(d.Status),
		})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *Service) DashboardBriefing(ctx context.Context, cmd ports.DashboardBriefingCommand) (ports.BriefingResult, error) {
	const capCode = "dashboard.briefing"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.BriefingResult{}, err
	}
	var parts []string
	if cmd.CraStatus != "" {
		parts = append(parts, fmt.Sprintf("CRA du mois : %s.", cmd.CraStatus))
	}
	if cmd.LeavePending > 0 {
		parts = append(parts, fmt.Sprintf("%d demande(s) de congés en attente.", cmd.LeavePending))
	}
	if cmd.PendingValidations > 0 {
		parts = append(parts, fmt.Sprintf("%d validation(s) manager à traiter.", cmd.PendingValidations))
	}
	if cmd.TmaOpen > 0 {
		parts = append(parts, fmt.Sprintf("%d demande(s) TMA ouverte(s).", cmd.TmaOpen))
	}
	if cmd.BudgetOverrun > 0 {
		parts = append(parts, fmt.Sprintf("%d budget(s) en dépassement.", cmd.BudgetOverrun))
	} else if cmd.BudgetConsumption > 0 {
		parts = append(parts, fmt.Sprintf("Consommation budget moyenne : %.0f%%.", cmd.BudgetConsumption))
	}
	text := "Aucune action urgente détectée."
	if len(parts) > 0 {
		text = strings.Join(parts, " ")
	}
	result := ports.BriefingResult{Text: text}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.BriefingResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) CongesManagerContext(ctx context.Context, cmd ports.CongesManagerCommand) (ports.ManagerContextResult, error) {
	const capCode = "conges.manager_assist"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.ManagerContextResult{}, err
	}
	leave, err := s.leaves.GetLeave(ctx, cmd.TenantID, cmd.LeaveRequestID)
	if err != nil {
		return ports.ManagerContextResult{}, err
	}
	balances, _ := s.leaves.ListBalances(ctx, cmd.TenantID, leave.UserID)
	pending, _ := s.leaves.ListLeaves(ctx, cmd.TenantID, ptrStatus(congesdomain.LeaveStatusPending))
	var balanceText string
	for _, b := range balances {
		balanceText += fmt.Sprintf("%s: %.1f restant; ", b.Type, b.Remaining)
	}
	contextText := fmt.Sprintf(
		"Demande du %s au %s. Motif : %s. Soldes demandeur : %s Autres demandes en attente équipe : %d. Contexte factuel — décision manager requise.",
		leave.Period.From.Format("02/01/2006"), leave.Period.To.Format("02/01/2006"), leave.Motif, balanceText, len(pending),
	)
	result := ports.ManagerContextResult{Context: contextText}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "leave_request", EntityID: ptrUUID(cmd.LeaveRequestID),
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
		ExplainContext: map[string]any{"leaveId": cmd.LeaveRequestID.String()},
	})
	if err != nil {
		return ports.ManagerContextResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) ExplainWorkflow(ctx context.Context, cmd ports.WorkflowExplainCommand) (domain.ExplainResult, error) {
	const capCode = "workflow.explain"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return domain.ExplainResult{}, err
	}
	inst, err := s.workflow.GetInstance(ctx, cmd.TenantID, wfdomain.InstanceID(cmd.InstanceID))
	if err != nil {
		return domain.ExplainResult{}, err
	}
	identity := authx.Identity{TenantID: cmd.TenantID, UserID: cmd.UserID}
	actions, _ := s.workflow.AvailableActions(ctx, cmd.TenantID, wfdomain.InstanceID(cmd.InstanceID), identity)
	summary := fmt.Sprintf("État courant : %s.", string(inst.CurrentState))
	if len(actions) > 0 {
		summary += fmt.Sprintf(" Actions possibles pour vous : %s.", strings.Join(actionCodes(actions), ", "))
	}
	result := domain.ExplainResult{
		Capability: capCode,
		Summary:    summary,
		Factors: []domain.ExplainFactor{
			{Label: "État", Value: string(inst.CurrentState)},
			{Label: "Actions", Value: strings.Join(actionCodes(actions), ", ")},
		},
		Disclaimer: "Information indicative — les transitions restent soumises au moteur workflow.",
	}
	out, _ := json.Marshal(result)
	reqID, _ := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "workflow_instance", EntityID: ptrUUID(cmd.InstanceID),
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	result.RequestID = reqID
	return result, nil
}

func actionCodes(actions []wfdomain.ActionCode) []string {
	out := make([]string, 0, len(actions))
	for _, a := range actions {
		out = append(out, string(a))
	}
	return out
}

func (s *Service) PublicChat(_ context.Context, cmd ports.PublicChatCommand) (ports.ChatResult, error) {
	reply := stub.PublicChatReply(cmd.Message)
	result := ports.ChatResult{Reply: reply}
	reqID := uuid.New()
	result.RequestID = reqID
	return result, nil
}

func (s *Service) ExplainRequest(ctx context.Context, tenant kernel.TenantID, requestID uuid.UUID) (domain.ExplainResult, error) {
	log, err := s.repo.GetRequestLog(ctx, tenant, requestID)
	if err != nil {
		return domain.ExplainResult{}, err
	}
	subject, _ := log.ExplainContext["subject"].(string)
	summary := fmt.Sprintf("Suggestion %s générée le %s.", log.CapabilityCode, log.CreatedAt.Format(time.RFC3339))
	factors := []domain.ExplainFactor{
		{Label: "Capability", Value: log.CapabilityCode},
		{Label: "Modèle", Value: log.Model},
	}
	if subject != "" {
		factors = append(factors, domain.ExplainFactor{Label: "Sujet", Value: subject})
	}
	return domain.ExplainResult{
		RequestID:  requestID,
		Capability: log.CapabilityCode,
		Summary:    summary,
		Factors:    factors,
		Disclaimer: "Suggestion non décisionnelle — validation humaine requise.",
	}, nil
}

func (s *Service) GetTenantSettings(ctx context.Context, tenant kernel.TenantID) (domain.TenantSettings, error) {
	return s.repo.GetTenantSettings(ctx, tenant)
}

func (s *Service) EnableAI(ctx context.Context, cmd ports.EnableAICommand) error {
	if !cmd.NoticeAccepted || !cmd.WorkersInformed {
		return fmt.Errorf("notice and worker information required")
	}
	now := s.clock.Now()
	return s.repo.UpsertTenantSettings(ctx, domain.TenantSettings{
		TenantID:          cmd.TenantID,
		Enabled:           true,
		NoticeAcceptedAt:  &now,
		NoticeAcceptedBy:  &cmd.UserID,
		WorkersInformedAt: &now,
		LLMProvider:       "stub",
	})
}

func ptrUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func ptrStatus(s congesdomain.LeaveStatus) *congesdomain.LeaveStatus {
	return &s
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

var _ ports.AIService = (*Service)(nil)
