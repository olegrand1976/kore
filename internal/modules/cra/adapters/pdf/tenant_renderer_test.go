package pdf

import "testing"

func TestTrimJoin(t *testing.T) {
	cases := []struct {
		name  string
		parts []string
		want  string
	}{
		{"no part", nil, ""},
		{"only empty parts", []string{"", ""}, ""},
		{"single part", []string{"ACME"}, "ACME"},
		{"drops empty parts in between", []string{"ACME", "", "Bruxelles"}, "ACME · Bruxelles"},
		{"all parts", []string{"ACME", "Mission", "Bruxelles"}, "ACME · Mission · Bruxelles"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := trimJoin(tc.parts...); got != tc.want {
				t.Fatalf("trimJoin(%q) = %q, want %q", tc.parts, got, tc.want)
			}
		})
	}
}

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
