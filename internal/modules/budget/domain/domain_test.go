package domain_test

import (
	"testing"

	"github.com/kore/kore/internal/modules/budget/domain"
	"github.com/stretchr/testify/assert"
)

func TestEffectiveEffort_QuoteOverridesEstimate(t *testing.T) {
	estimate := &domain.Estimate{Effort: domain.Effort{Days: 5, UO: 10}}
	quote := &domain.Quote{Effort: domain.Effort{Days: 3, UO: 6}}

	effort := domain.EffectiveEffort(estimate, quote)
	assert.Equal(t, 3.0, effort.Days)
	assert.Equal(t, 6.0, effort.UO)
}

func TestBudget_IsOverrun(t *testing.T) {
	b := domain.Budget{
		Planned:   domain.ConsumptionTriple{Days: 10, UO: 100, Amount: 10000},
		Consumed:  domain.ConsumptionTriple{Days: 11, UO: 50, Amount: 5000},
		Remaining: domain.ConsumptionTriple{Days: -1, UO: 50, Amount: 5000},
	}
	assert.True(t, b.IsOverrun())
}
