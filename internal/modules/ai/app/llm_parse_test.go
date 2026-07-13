package app

import (
	"testing"

	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/stretchr/testify/assert"
)

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
