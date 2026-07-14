package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
)

func (s *Service) SuggestAssignee(ctx context.Context, cmd ports.SuggestAssigneeCommand) (ports.SuggestAssigneeResult, error) {
	const capCode = "tma.suggest_assignee"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.SuggestAssigneeResult{}, err
	}
	demands, err := s.tma.ListDemands(ctx, cmd.TenantID, true)
	if err != nil {
		return ports.SuggestAssigneeResult{}, err
	}
	var candidate uuid.UUID
	for _, d := range demands {
		if d.AssigneeID != nil && d.Status != tmadomain.DemandStatusResolved {
			candidate = *d.AssigneeID
			break
		}
	}
	rationale := "Suggestion basée sur la charge courante (stub)."
	if candidate == uuid.Nil {
		rationale = "Aucun assigné actif détecté — sélection manuelle requise."
	}
	result := ports.SuggestAssigneeResult{
		SuggestedUserID: candidate,
		Rationale:       rationale,
	}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "tma_demand", EntityID: ptrUUID(cmd.DemandID),
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.SuggestAssigneeResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) ExecutiveSummary(ctx context.Context, cmd ports.ExecutiveSummaryCommand) (ports.ExecutiveSummaryResult, error) {
	const capCode = "tma.executive_summary"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.ExecutiveSummaryResult{}, err
	}
	demands, err := s.tma.ListDemands(ctx, cmd.TenantID, true)
	if err != nil {
		return ports.ExecutiveSummaryResult{}, err
	}
	open, resolved := 0, 0
	for _, d := range demands {
		if cmd.ApplicationID != nil && d.ApplicationID != *cmd.ApplicationID {
			continue
		}
		if d.Status == tmadomain.DemandStatusResolved {
			resolved++
		} else {
			open++
		}
	}
	summary := fmt.Sprintf("Portefeuille TMA : %d ouvert(s), %d résolu(s). Synthèse indicative pour reporting.", open, resolved)
	highlights := []string{
		fmt.Sprintf("%d demandes ouvertes", open),
		fmt.Sprintf("%d demandes résolues", resolved),
	}
	result := ports.ExecutiveSummaryResult{Summary: summary, Highlights: highlights}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.ExecutiveSummaryResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) SummarizeCraComments(ctx context.Context, cmd ports.CommentSummaryCommand) (ports.CommentSummaryResult, error) {
	const capCode = "cra.comment_summary"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.CommentSummaryResult{}, err
	}
	ts, err := s.cra.GetTimesheetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return ports.CommentSummaryResult{}, err
	}
	var comments []string
	for _, week := range ts.Weeks {
		if cmd.WeekNumber > 0 && int(week.WeekNumber) != cmd.WeekNumber {
			continue
		}
		for _, line := range week.Lines {
			if c := strings.TrimSpace(line.Comment); c != "" {
				comments = append(comments, c)
			}
		}
	}
	summary := "Aucun commentaire saisi sur la période."
	if len(comments) > 0 {
		summary = fmt.Sprintf("Résumé (%d commentaire(s)) : %s", len(comments), strings.Join(comments, " ; "))
		if len(summary) > 500 {
			summary = summary[:497] + "..."
		}
	}
	result := ports.CommentSummaryResult{Summary: summary}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "timesheet", EntityID: ptrUUID(cmd.TimesheetID),
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.CommentSummaryResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) ForecastBudgetOverrun(ctx context.Context, cmd ports.OverrunForecastCommand) (ports.OverrunForecastResult, error) {
	const capCode = "budget.overrun_forecast"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.OverrunForecastResult{}, err
	}
	result := ports.OverrunForecastResult{
		ForecastDays: 12,
		ForecastUO:   24,
		OverrunRisk:  "medium",
		Narrative:    "Estimation indicative : tendance linéaire sur consommation courante (stub).",
	}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "budget", EntityID: ptrUUID(cmd.BudgetID),
		InputHash: hashInput(cmd), OutputJSON: out, Model: "rules-v1",
	})
	if err != nil {
		return ports.OverrunForecastResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) SuggestLeaveDates(ctx context.Context, cmd ports.DateSuggestCommand) (ports.DateSuggestResult, error) {
	const capCode = "conges.date_suggest"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.DateSuggestResult{}, err
	}
	days := cmd.DurationDays
	if days <= 0 {
		days = 5
	}
	start := s.clock.Now().AddDate(0, 1, 0)
	end := start.AddDate(0, 0, days-1)
	suggestions := []ports.DateSuggestion{{
		From:      start.Format("2006-01-02"),
		To:        end.Format("2006-01-02"),
		Rationale: "Créneau proposé hors période immédiate (stub).",
	}}
	result := ports.DateSuggestResult{Suggestions: suggestions}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.DateSuggestResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) ScoreLead(ctx context.Context, cmd ports.LeadScoringCommand) (ports.LeadScoringResult, error) {
	score := 30
	factors := []string{"formulaire incomplet"}
	if strings.TrimSpace(cmd.Email) != "" {
		score += 20
		factors = append(factors, "email renseigné")
	}
	if len(cmd.Modules) > 0 {
		score += 15 * len(cmd.Modules)
		if score > 90 {
			score = 90
		}
		factors = append(factors, "modules cités")
	}
	if strings.Contains(strings.ToLower(cmd.CompanySize), "esn") || strings.Contains(strings.ToLower(cmd.CompanySize), "100") {
		score += 15
		factors = append(factors, "taille ESN")
	}
	tier := "cold"
	switch {
	case score >= 70:
		tier = "hot"
	case score >= 45:
		tier = "warm"
	}
	result := ports.LeadScoringResult{Score: score, Tier: tier, Factors: factors}
	result.RequestID = uuid.New()
	return result, nil
}

func (s *Service) DigestNotifications(ctx context.Context, cmd ports.NotificationsDigestCommand) (ports.NotificationsDigestResult, error) {
	const capCode = "notifications.digest"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.NotificationsDigestResult{}, err
	}
	count := cmd.UnreadTma + cmd.UnreadLeave + cmd.UnreadCra + cmd.UnreadWorkflow
	period := cmd.Period
	if period == "" {
		period = "daily"
	}
	digest := fmt.Sprintf("Digest %s : %d notification(s) non lue(s) (TMA %d, congés %d, CRA %d, workflow %d).",
		period, count, cmd.UnreadTma, cmd.UnreadLeave, cmd.UnreadCra, cmd.UnreadWorkflow)
	links := []ports.DigestLink{
		{Label: "TMA", Href: "/tma"},
		{Label: "Congés", Href: "/conges"},
	}
	result := ports.NotificationsDigestResult{Digest: digest, ItemCount: count, Links: links}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		InputHash: hashInput(cmd), OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.NotificationsDigestResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}

func (s *Service) TranscribeVoiceCra(ctx context.Context, cmd ports.VoiceCraCommand) (ports.VoiceCraResult, error) {
	const capCode = "mobile.voice_cra"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.VoiceCraResult{}, err
	}
	transcript := strings.TrimSpace(cmd.Transcript)
	if transcript == "" {
		return ports.VoiceCraResult{}, fmt.Errorf("transcript required")
	}
	result := ports.VoiceCraResult{Lines: []ports.VoiceCraLine{{
		Duration: 1,
		Comment:  transcript,
	}}}
	out, _ := json.Marshal(result)
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID: cmd.TenantID, UserID: cmd.UserID, CapabilityCode: capCode,
		EntityType: "timesheet", EntityID: ptrUUID(cmd.TimesheetID),
		InputHash: hashInput(map[string]any{"week": cmd.WeekNumber, "len": len(transcript)}),
		OutputJSON: out, Model: stub.ModelName,
	})
	if err != nil {
		return ports.VoiceCraResult{}, err
	}
	result.RequestID = reqID
	return result, nil
}
