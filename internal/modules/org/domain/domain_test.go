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
		{"valid with prefix", "ABC_jean", true},
		{"valid without prefix", "olivier", true},
		{"valid dotted", "jean.dupont", true},
		{"valid hyphen", "jean-dupont", true},
		{"valid mixed case", "JeanDupont", true},
		{"valid seed style", "ADM_admin", true},
		{"invalid too short", "ab", false},
		{"invalid starts with digit", "1olivier", false},
		{"invalid space", "jean dupont", false},
		{"invalid special char", "jean@dupont", false},
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

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name string
		pwd  string
		ok   bool
	}{
		{"strong", "Admin123!", true},
		{"too short", "Ab1!", false},
		{"no upper", "admin123!", false},
		{"no lower", "ADMIN123!", false},
		{"no digit", "AdminPass!", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pwd)
			if tt.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrWeakPassword)
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
