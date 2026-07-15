package app

import (
	"fmt"
	"strings"

	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
)

func parseAnalysisDraftLLM(text string) (domain.AnalysisDraft, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return domain.AnalysisDraft{}, false
	}

	draft := domain.AnalysisDraft{}
	known := map[string]*string{
		"FUNCTIONAL": &draft.Functional,
		"TECHNICAL":  &draft.Technical,
		"RISKS":      &draft.Risks,
		"TESTS":      &draft.TestScenario,
	}

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for prefix, target := range known {
			tag := prefix + "|"
			if strings.HasPrefix(strings.ToUpper(line), tag) {
				*target = strings.TrimSpace(line[len(tag):])
				break
			}
		}
	}

	if draft.Functional == "" && draft.Technical == "" {
		return domain.AnalysisDraft{}, false
	}
	if draft.Functional == "" {
		draft.Functional = text
	}
	if draft.Technical == "" {
		draft.Technical = "À compléter après investigation technique."
	}
	if draft.Risks == "" {
		draft.Risks = "Évaluer les risques de régression sur le périmètre impacté."
	}
	if draft.TestScenario == "" {
		draft.TestScenario = "1. Reproduire le cas. 2. Valider le correctif. 3. Smoke test non-régression."
	}
	return draft, true
}

func parseClassifyLLM(text string) (category string, confidence float64, ok bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", 0, false
	}
	category = ""
	confidence = 0
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "CATEGORY|"):
			category = strings.ToLower(strings.TrimSpace(line[len("CATEGORY|"):]))
		case strings.HasPrefix(upper, "CONFIDENCE|"):
			if v, err := parseConfidence(strings.TrimSpace(line[len("CONFIDENCE|"):])); err == nil {
				confidence = v
			}
		}
	}
	if category == "" {
		return "", 0, false
	}
	if confidence <= 0 {
		confidence = 0.7
	}
	return category, confidence, true
}

func parseConfidence(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, "%")
	var v float64
	_, err := fmt.Sscanf(raw, "%f", &v)
	if err != nil {
		return 0, err
	}
	if v > 1 {
		v = v / 100
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return v, nil
}

func parseExecutiveSummaryLLM(text string) (summary string, highlights []string, ok bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil, false
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "SUMMARY|"):
			summary = strings.TrimSpace(line[len("SUMMARY|"):])
		case strings.HasPrefix(upper, "HIGHLIGHT|"):
			if h := strings.TrimSpace(line[len("HIGHLIGHT|"):]); h != "" {
				highlights = append(highlights, h)
			}
		}
	}
	if summary == "" {
		return "", nil, false
	}
	return summary, highlights, true
}

func briefingContextFields(cmd ports.DashboardBriefingCommand) map[string]string {
	fields := map[string]string{
		"profile": cmd.Profile,
	}
	if cmd.CraStatus != "" {
		fields["cra_status"] = cmd.CraStatus
	}
	if cmd.LeavePending > 0 {
		fields["leave_pending"] = fmt.Sprintf("%d", cmd.LeavePending)
	}
	if cmd.PendingValidations > 0 {
		fields["pending_validations"] = fmt.Sprintf("%d", cmd.PendingValidations)
	}
	if cmd.TmaOpen > 0 {
		fields["tma_open"] = fmt.Sprintf("%d", cmd.TmaOpen)
	}
	if cmd.BudgetOverrun > 0 {
		fields["budget_overrun"] = fmt.Sprintf("%d", cmd.BudgetOverrun)
	}
	if cmd.BudgetConsumption > 0 {
		fields["budget_consumption_pct"] = fmt.Sprintf("%.0f", cmd.BudgetConsumption)
	}
	return fields
}

func buildBriefingFallback(cmd ports.DashboardBriefingCommand) string {
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
	if len(parts) == 0 {
		return "Aucune action urgente détectée."
	}
	return strings.Join(parts, " ")
}
