package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRateUnit(t *testing.T) {
	u, err := NormalizeRateUnit("")
	require.NoError(t, err)
	require.Equal(t, RateUnitTJM, u)

	u, err = NormalizeRateUnit("hourly")
	require.NoError(t, err)
	require.Equal(t, RateUnitHourly, u)

	_, err = NormalizeRateUnit("weekly")
	require.ErrorIs(t, err, ErrInvalidRateUnit)
}

func TestNewMission_defaultsTJM(t *testing.T) {
	m := NewMission(kernel.NewTenantID(uuid.New()), uuid.New(), time.Now().UTC(), 50000)
	require.Equal(t, RateUnitTJM, m.RateUnit)
	require.Empty(t, m.Title)
}
