package app

import (
	"testing"

	"github.com/kore/kore/internal/modules/org/domain"
)

func TestResolveIDPClientSecretPrefersEnv(t *testing.T) {
	t.Setenv("OIDC_GOOGLE_CLIENT_SECRET", "from-env")
	idp := domain.IdentityProvider{
		Issuer:       "https://accounts.google.com",
		ClientSecret: "from-db",
	}
	if got := resolveIDPClientSecret(idp); got != "from-env" {
		t.Fatalf("expected env secret, got %q", got)
	}
}

func TestResolveIDPClientSecretFallsBackToDB(t *testing.T) {
	t.Setenv("OIDC_GOOGLE_CLIENT_SECRET", "")
	idp := domain.IdentityProvider{
		Issuer:       "https://accounts.google.com",
		ClientSecret: "from-db",
	}
	if got := resolveIDPClientSecret(idp); got != "from-db" {
		t.Fatalf("expected db secret, got %q", got)
	}
}

func TestResolveIDPClientSecretSMRef(t *testing.T) {
	t.Setenv("OIDC_GOOGLE_CLIENT_SECRET", "from-sm")
	idp := domain.IdentityProvider{
		Issuer:       "https://accounts.google.com",
		ClientSecret: "sm://kore-oidc-google-client-secret",
	}
	if got := resolveIDPClientSecret(idp); got != "from-sm" {
		t.Fatalf("expected sm ref via env, got %q", got)
	}
}
