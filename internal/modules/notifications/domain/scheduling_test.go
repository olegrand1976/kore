package domain

import (
	"testing"
	"time"
)

func TestNextRun(t *testing.T) {
	// Reference: Wednesday 2026-01-07 10:00 UTC.
	now := time.Date(2026, time.January, 7, 10, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		freq Frequency
		want time.Time
	}{
		{
			name: "immediate returns now",
			freq: FrequencyImmediate,
			want: now,
		},
		{
			name: "morning next day 08h (past today's 08h)",
			freq: FrequencyMorning,
			want: time.Date(2026, time.January, 8, scheduledSendHour, 0, 0, 0, time.UTC),
		},
		{
			name: "monday next monday",
			freq: FrequencyMonday,
			want: time.Date(2026, time.January, 12, scheduledSendHour, 0, 0, 0, time.UTC),
		},
		{
			name: "friday next friday",
			freq: FrequencyFriday,
			want: time.Date(2026, time.January, 9, scheduledSendHour, 0, 0, 0, time.UTC),
		},
		{
			name: "last monday of january 2026",
			freq: FrequencyLastMondayOfMonth,
			want: time.Date(2026, time.January, 26, scheduledSendHour, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NextRun(tc.freq, now)
			if !got.Equal(tc.want) {
				t.Fatalf("NextRun(%s) = %s, want %s", tc.freq, got, tc.want)
			}
			if tc.freq != FrequencyImmediate && !got.After(now) {
				t.Fatalf("NextRun(%s) = %s must be strictly after now %s", tc.freq, got, now)
			}
		})
	}
}

func TestNextRunLastMondayRollsToNextMonth(t *testing.T) {
	// After the last monday of january (2026-01-26 08:00), the next occurrence
	// must roll over to february's last monday (2026-02-23).
	now := time.Date(2026, time.January, 26, 9, 0, 0, 0, time.UTC)
	got := NextRun(FrequencyLastMondayOfMonth, now)
	want := time.Date(2026, time.February, 23, scheduledSendHour, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("NextRun rollover = %s, want %s", got, want)
	}
}
