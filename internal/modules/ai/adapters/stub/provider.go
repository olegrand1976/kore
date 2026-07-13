package stub

import (
	"context"
	"fmt"
	"strings"

	"github.com/kore/kore/internal/modules/ai/ports"
)

const ModelName = "stub-v1"

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) Complete(_ context.Context, req ports.CompletionRequest) (ports.CompletionResponse, error) {
	return ports.CompletionResponse{
		Text:  strings.TrimSpace(req.UserPrompt),
		Model: ModelName,
	}, nil
}

func BuildAnalysisDraft(subject string) ports.CompletionResponse {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = "demande TMA"
	}
	text := fmt.Sprintf(`FUNCTIONAL|Analyse fonctionnelle du sujet « %s » : comportement attendu à restaurer, périmètre utilisateur impacté, scénario de reproduction.
TECHNICAL|Investigation technique : logs applicatifs, couche concernée (API/UI/batch), hypothèse root cause à confirmer.
RISKS|Risques de régression sur le périmètre adjacent ; vérifier jeux de données et déploiements récents.
TESTS|1. Reproduire le cas nominal. 2. Valider le correctif. 3. Non-régression smoke sur module lié.`, subject)
	return ports.CompletionResponse{Text: text, Model: ModelName}
}

func ClassifySubject(subject string) (category string, confidence float64) {
	lower := strings.ToLower(subject)
	switch {
	case strings.Contains(lower, "régression") || strings.Contains(lower, "regression"):
		return "regression", 0.85
	case strings.Contains(lower, "évolution") || strings.Contains(lower, "feature"):
		return "evolution", 0.75
	case strings.Contains(lower, "question") || strings.Contains(strings.ToLower(subject), "?"):
		return "question", 0.7
	default:
		return "incident", 0.8
	}
}

func PublicChatReply(message string) string {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "cra") || strings.Contains(lower, "activité"):
		return "Kore unifie la saisie CRA, les congés et le suivi TMA/budget sans double saisie. Souhaitez-vous une démo ?"
	case strings.Contains(lower, "prix") || strings.Contains(lower, "tarif"):
		return "Consultez la page Tarifs pour les offres Starter/Pro. Je peux vous orienter vers une réservation démo."
	case strings.Contains(lower, "tma"):
		return "Le module TMA couvre le cycle incident complet : gate chef utilisateur, analyse, budget UO et export XML."
	default:
		return "Je suis l'assistant IA Kore. Je peux présenter les modules CRA, TMA, congés et budget. Que recherchez-vous ?"
	}
}
