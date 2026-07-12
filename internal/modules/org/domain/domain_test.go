package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoginValid(t *testing.T) {
	tests := []struct {
		name  string
		login string
		ok    bool
	}{
		{"valid", "ABC_jean", true},
		{"invalid lowercase prefix", "abc_jean", false},
		{"invalid no underscore", "ABCjean", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLogin(tt.login)
			if tt.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrInvalidLogin)
			}
		})
	}
}

func TestActivationPeriodExpired(t *testing.T) {
	exp := time.Now().Add(-24 * time.Hour)
	period := ActivationPeriod{
		Activation: time.Now().Add(-48 * time.Hour),
		Expiration: &exp,
	}
	require.False(t, period.IsActive(time.Now()))
}
