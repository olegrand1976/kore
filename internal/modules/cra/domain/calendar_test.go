package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeekDaysInMonth_July2026Monday(t *testing.T) {
	month := Month("2026-07")
	days, err := WeekDaysInMonth(month, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"2026-07-01", "2026-07-02", "2026-07-03", "2026-07-04", "2026-07-05"}, days)

	days2, err := WeekDaysInMonth(month, 2, 1)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"2026-07-06", "2026-07-07", "2026-07-08", "2026-07-09",
		"2026-07-10", "2026-07-11", "2026-07-12",
	}, days2)
}

func TestMonthWeekNumber_July2026Monday(t *testing.T) {
	month := Month("2026-07")
	day := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	week, err := MonthWeekNumber(day, month, 1)
	require.NoError(t, err)
	assert.Equal(t, WeekNumber(2), week)
}

func TestWeekDaysInMonth_July2026Sunday(t *testing.T) {
	month := Month("2026-07")
	days, err := WeekDaysInMonth(month, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"2026-07-01", "2026-07-02", "2026-07-03", "2026-07-04"}, days)

	days2, err := WeekDaysInMonth(month, 2, 0)
	require.NoError(t, err)
	assert.Contains(t, days2, "2026-07-05")
	assert.Contains(t, days2, "2026-07-06")
}
