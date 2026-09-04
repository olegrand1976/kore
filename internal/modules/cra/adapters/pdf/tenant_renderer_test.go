package pdf

import "testing"

func TestFormatHours(t *testing.T) {
	cases := []struct {
		name    string
		minutes int
		want    string
	}{
		{"whole hours", 480, "8"},
		{"half hour", 450, "7.5"},
		{"quarter hour", 375, "6.25"},
		{"three quarters", 45, "0.75"},
		{"non round minutes", 361, "6.02"},
		{"zero", 0, "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatHours(tc.minutes); got != tc.want {
				t.Fatalf("formatHours(%d) = %q, want %q", tc.minutes, got, tc.want)
			}
		})
	}
}
