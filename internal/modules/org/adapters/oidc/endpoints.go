package oidc

import "strings"

const (
	GoogleIssuer   = "https://accounts.google.com"
	GoogleJWKSURI  = "https://www.googleapis.com/oauth2/v3/certs"
	GoogleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	GoogleTokenURL = "https://oauth2.googleapis.com/token"
)

func isGoogleIssuer(issuer string) bool {
	base := strings.TrimRight(strings.ToLower(strings.TrimSpace(issuer)), "/")
	return base == "https://accounts.google.com" ||
		base == "accounts.google.com" ||
		strings.Contains(base, "googleapis.com")
}

func authorizeURLs(issuer string) []string {
	if isGoogleIssuer(issuer) {
		return []string{GoogleAuthURL}
	}
	base := strings.TrimRight(issuer, "/")
	return []string{
		base + "/oauth2/v2.0/authorize",
		base + "/authorize",
		base + "/protocol/openid-connect/auth",
	}
}

func tokenURLs(issuer string) []string {
	if isGoogleIssuer(issuer) {
		return []string{GoogleTokenURL}
	}
	base := strings.TrimRight(issuer, "/")
	return []string{
		base + "/oauth2/v2.0/token",
		base + "/token",
		base + "/protocol/openid-connect/token",
	}
}
