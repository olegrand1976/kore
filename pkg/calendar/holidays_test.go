package calendar

import (
	"testing"
	"time"
)

func TestIsPublicHoliday_FR2026(t *testing.T) {
	cases := []struct {
		date string
		want bool
	}{
		{"2026-01-01", true},
		{"2026-04-06", true},
		{"2026-07-14", true},
		{"2026-07-15", false},
	}
	for _, tc := range cases {
		day, err := time.Parse("2006-01-02", tc.date)
		if err != nil {
			t.Fatal(err)
		}
		if got := IsPublicHoliday(day, "FR"); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.date, got, tc.want)
		}
	}
}
