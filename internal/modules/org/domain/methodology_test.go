package domain_test

import (
	"testing"

	"github.com/kore/kore/internal/modules/org/domain"
)

func TestNormalizeMethodologyProfile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want domain.MethodologyProfile
		err  bool
	}{
		{"", domain.MethodologyPSA, false},
		{"psa", domain.MethodologyPSA, false},
		{"agile_scrum", domain.MethodologyAgileScrum, false},
		{"agile_kanban", domain.MethodologyAgileKanban, false},
		{"scrum", "", true},
	}
	for _, tc := range cases {
		got, err := domain.NormalizeMethodologyProfile(tc.in)
		if tc.err {
			if err == nil {
				t.Fatalf("expected error for %q", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("got %q want %q", got, tc.want)
		}
	}
}

func TestMethodologyProfileIsAgile(t *testing.T) {
	t.Parallel()
	if domain.MethodologyPSA.IsAgile() {
		t.Fatal("psa should not be agile")
	}
	if !domain.MethodologyAgileScrum.IsAgile() {
		t.Fatal("scrum should be agile")
	}
}
