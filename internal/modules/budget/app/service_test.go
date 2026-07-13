package app

import (
	"testing"

	"github.com/kore/kore/internal/modules/budget/domain"
	"github.com/stretchr/testify/require"
)

func TestNewBudgetService(t *testing.T) {
	svc := NewService(nil, nil)
	require.NotNil(t, svc)
}

func TestBudgetTypeDefault(t *testing.T) {
	require.Equal(t, domain.BudgetTypeDefault, domain.BudgetType("defaut"))
}
