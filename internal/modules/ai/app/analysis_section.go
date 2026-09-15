package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
)

func (s *Service) SuggestAnalysisSection(ctx context.Context, cmd ports.AnalysisSectionCommand) (ports.AnalysisSectionResult, error) {
	const capCode = "tma.analysis_section"
	if err := s.ensureAI(ctx, cmd.TenantID, capCode); err != nil {
		return ports.AnalysisSectionResult{}, err
	}
	section := strings.TrimSpace(cmd.Section)
	if !domain.ValidAnalysisSection(section) {
		return ports.AnalysisSectionResult{}, domain.ErrInvalidAnalysisSection
	}
	prompt := strings.TrimSpace(cmd.Prompt)
	if prompt == "" {
		return ports.AnalysisSectionResult{}, domain.ErrEmptyAnalysisPrompt
	}
	if cmd.DemandID == uuid.Nil {
		return ports.AnalysisSectionResult{}, domain.ErrAnalysisDemandNotFound
	}
	if s.tma == nil {
		return ports.AnalysisSectionResult{}, domain.ErrAnalysisDemandNotFound
	}
	demand, err := s.tma.GetDemand(ctx, cmd.TenantID, cmd.DemandID)
	if err != nil {
		if errors.Is(err, tmadomain.ErrDemandNotFound) {
			return ports.AnalysisSectionResult{}, domain.ErrAnalysisDemandNotFound
		}
		return ports.AnalysisSectionResult{}, err
	}

	subject := strings.TrimSpace(cmd.Subject)
	if subject == "" {
		subject = strings.TrimSpace(demand.Subject)
	}
	if subject == "" {
		subject = "demande TMA"
	}

	var sources []domain.DocumentChunk
	if cmd.UseRAG {
		query := subject + "\n" + prompt
		retrieved, err := s.retrieveRAG(ctx, cmd.TenantID, cmd.DemandID, query, 6)
		if err != nil {
			return ports.AnalysisSectionResult{}, err
		}
		sources = retrieved
	}

	text := buildSectionDraft(section, subject, prompt)
	model := stub.ModelName
	system := fmt.Sprintf(
		`Tu rédiges UNIQUEMENT la section « %s » d'un dossier d'analyse TMA en français.
Réponds par le texte de la section uniquement, sans préfixe de titre.`,
		sectionLabelFR(section),
	)
	userParts := map[string]string{
		"subject": subject,
		"section": section,
		"prompt":  prompt,
	}
	if rag := formatRAGContext(sources); rag != "" {
		userParts["documents"] = rag
	}
	if resp, err := s.llmComplete(ctx, capCode, system, userParts); err == nil {
		// Stub Complete echoes the sandboxed user prompt — never persist that as section text.
		if resp.Model != stub.ModelName {
			if trimmed := strings.TrimSpace(resp.Text); trimmed != "" {
				text = trimmed
				model = resp.Model
			}
		}
	}

	out, _ := json.Marshal(map[string]any{"text": text, "section": section})
	reqID, err := s.logRequest(ctx, domain.RequestLog{
		TenantID:       cmd.TenantID,
		UserID:         cmd.UserID,
		CapabilityCode: capCode,
		EntityType:     "tma_demand",
		EntityID:       ptrUUID(cmd.DemandID),
		InputHash:      hashInput(cmd),
		OutputJSON:     out,
		Model:          model,
		ExplainContext: map[string]any{
			"subject":  subject,
			"section":  section,
			"useRAG":   cmd.UseRAG,
			"chunkIds": chunkIDs(sources),
			"sources":  ragSources(sources),
		},
	})
	if err != nil {
		return ports.AnalysisSectionResult{}, err
	}
	return ports.AnalysisSectionResult{
		Text:      text,
		RequestID: reqID,
		Sources:   ragSources(sources),
	}, nil
}

func sectionLabelFR(section string) string {
	switch section {
	case domain.AnalysisSectionFunctional:
		return "analyse fonctionnelle"
	case domain.AnalysisSectionTechnical:
		return "analyse technique"
	case domain.AnalysisSectionRisks:
		return "risques"
	case domain.AnalysisSectionTestScenario:
		return "scénario de test"
	default:
		return section
	}
}

func buildSectionDraft(section, subject, prompt string) string {
	base := fmt.Sprintf("Sujet « %s ». Consigne : %s.", subject, prompt)
	switch section {
	case domain.AnalysisSectionFunctional:
		return "Analyse fonctionnelle — " + base
	case domain.AnalysisSectionTechnical:
		return "Investigation technique — " + base
	case domain.AnalysisSectionRisks:
		return "Risques identifiés — " + base
	case domain.AnalysisSectionTestScenario:
		return "Scénario de test — " + base + " 1) Reproduire. 2) Corriger. 3) Non-régression."
	default:
		return base
	}
}

func chunkIDs(chunks []domain.DocumentChunk) []string {
	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		out = append(out, c.ID.String())
	}
	return out
}
