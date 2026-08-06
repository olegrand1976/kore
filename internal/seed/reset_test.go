package seed

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrSeedProtected_isSentinel(t *testing.T) {
	err := fmt.Errorf("%w : seed-reset impossible tant qu'une société seed_protected existe", ErrSeedProtected)
	if !errors.Is(err, ErrSeedProtected) {
		t.Fatalf("errors.Is = false for %v", err)
	}
}

func TestGenerateStrongPassword(t *testing.T) {
	a, err := generateStrongPassword()
	if err != nil {
		t.Fatal(err)
	}
	b, err := generateStrongPassword()
	if err != nil {
		t.Fatal(err)
	}
	if a == "" || b == "" || a == b {
		t.Fatalf("expected distinct non-empty passwords, got %q / %q", a, b)
	}
	if len(a) < 16 {
		t.Fatalf("password too short: %d", len(a))
	}
}
