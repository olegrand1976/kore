package seed

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
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

func TestLLITTenantIsolatedFromDemo(t *testing.T) {
	if LLITTenantID == DemoTenantID {
		t.Fatal("LLITTenantID must differ from DemoTenantID")
	}
	if LLITSocieteID == DemoSocieteID {
		t.Fatal("LLITSocieteID must differ from DemoSocieteID")
	}
	if ProdAdminLogin == AdminLogin {
		t.Fatal("prod and demo admin logins must differ")
	}
}

func TestShouldRestoreDemoSociete(t *testing.T) {
	if !shouldRestoreDemoSociete(LLITSocieteName) {
		t.Fatal("expected restore when demo société was renamed to LL-IT")
	}
	if shouldRestoreDemoSociete(DemoSocieteName) {
		t.Fatal("must not restore a correctly named demo société")
	}
	if shouldRestoreDemoSociete("Autre SAS") {
		t.Fatal("must not restore an unrelated société name")
	}
}

func TestIsNotFound(t *testing.T) {
	if !isNotFound(pgx.ErrNoRows) {
		t.Fatal("raw pgx.ErrNoRows")
	}
	if !isNotFound(fmt.Errorf("user not found: %w", pgx.ErrNoRows)) {
		t.Fatal("wrapped pgx.ErrNoRows")
	}
	if isNotFound(errors.New("connection refused")) {
		t.Fatal("non-not-found error treated as not found")
	}
	if isNotFound(nil) {
		t.Fatal("nil must not be not-found")
	}
}
