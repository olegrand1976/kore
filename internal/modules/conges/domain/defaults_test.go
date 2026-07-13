package domain_test

import (
	"testing"

	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLeaveTypesForCountry_FR(t *testing.T) {
	types, err := domain.DefaultLeaveTypesForCountry("FR")
	require.NoError(t, err)
	require.Len(t, types, 3)
	assert.Equal(t, "conges_payes", types[0].Code)
	assert.Equal(t, "rtt", types[1].Code)
	assert.Equal(t, "maladie", types[2].Code)
	assert.True(t, types[0].TracksBalance)
	assert.False(t, types[2].TracksBalance)
}

func TestDefaultLeaveTypesForCountry_BE(t *testing.T) {
	types, err := domain.DefaultLeaveTypesForCountry("be")
	require.NoError(t, err)
	require.Len(t, types, 3)
	assert.Equal(t, "conges_annuels", types[0].Code)
	assert.Equal(t, "recuperation", types[1].Code)
}

func TestDefaultLeaveTypesForCountry_Unsupported(t *testing.T) {
	_, err := domain.DefaultLeaveTypesForCountry("DE")
	assert.ErrorIs(t, err, domain.ErrUnsupportedCountry)
}
