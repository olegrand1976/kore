package app

import (
	"os"
	"strings"

	"github.com/kore/kore/internal/modules/org/domain"
)

var secretRefEnv = map[string]string{
	"kore-oidc-google-client-secret": "OIDC_GOOGLE_CLIENT_SECRET",
	"kore-oidc-azure-client-secret":  "OIDC_AZURE_CLIENT_SECRET",
}

// resolveIDPClientSecret préfère les secrets runtime (Secret Manager / env) au secret stocké en DB.
func resolveIDPClientSecret(idp domain.IdentityProvider) string {
	issuer := strings.ToLower(strings.TrimSpace(idp.Issuer))
	switch {
	case strings.Contains(issuer, "accounts.google.com"):
		if secret := strings.TrimSpace(os.Getenv("OIDC_GOOGLE_CLIENT_SECRET")); secret != "" {
			return secret
		}
	case strings.Contains(issuer, "microsoftonline.com"):
		if secret := strings.TrimSpace(os.Getenv("OIDC_AZURE_CLIENT_SECRET")); secret != "" {
			return secret
		}
	}
	secret := strings.TrimSpace(idp.ClientSecret)
	if strings.HasPrefix(secret, "sm://") {
		ref := strings.TrimPrefix(secret, "sm://")
		if envKey, ok := secretRefEnv[ref]; ok {
			if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
				return v
			}
		}
	}
	return secret
}
