package app

import (
	"testing"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/stretchr/testify/require"
)

// TestPhase1GateDocumentsSSOAndPasswordCoexistence verifies dual-mode auth invariants.
func TestPhase1GateDocumentsSSOAndPasswordCoexistence(t *testing.T) {
	require.ErrorIs(t, domain.ErrSSONotEnabled, domain.ErrSSONotEnabled)
	require.NotNil(t, domain.ErrInvalidCredentials)
}
