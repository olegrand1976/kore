package app

import (
	"testing"

	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/stretchr/testify/assert"
)

func TestParseClassifyLLM(t *testing.T) {
	cat, conf, ok := parseClassifyLLM("CATEGORY|regression\nCONFIDENCE|0.92")
	if !ok || cat != "regression" || conf != 0.92 {
		t.Fatalf("parseClassifyLLM: cat=%s conf=%f ok=%v", cat, conf, ok)
	}
}

func TestParseExecutiveSummaryLLM(t *testing.T) {
	summary, highlights, ok := parseExecutiveSummaryLLM("SUMMARY|Synthèse TMA\nHIGHLIGHT|3 ouverts\nHIGHLIGHT|1 bloquant")
	if !ok || summary == "" || len(highlights) != 2 {
		t.Fatalf("parseExecutiveSummaryLLM: summary=%q highlights=%v ok=%v", summary, highlights, ok)
	}
}

func TestParseAnalysisDraftLLM(t *testing.T) {
	text := `FUNCTIONAL|Comportement attendu restauré.
TECHNICAL|Logs API à analyser.
RISKS|Régression possible.
TESTS|Reproduire puis valider.`

	draft, ok := parseAnalysisDraftLLM(text)
	assert.True(t, ok)
	assert.Contains(t, draft.Functional, "Comportement")
	assert.Contains(t, draft.Technical, "Logs")
	assert.Contains(t, draft.Risks, "Régression")
	assert.Contains(t, draft.TestScenario, "Reproduire")
}

func TestBuildBriefingFallback_empty(t *testing.T) {
	text := buildBriefingFallback(ports.DashboardBriefingCommand{})
	assert.Equal(t, "Aucune action urgente détectée.", text)
}
