package oidc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoogleAuthorizeURL(t *testing.T) {
	url, err := BuildAuthorizeURL(
		GoogleIssuer,
		"client-id",
		"http://localhost:3001/login",
		"openid profile email",
		"state123",
		"challenge",
		"nonce123",
	)
	require.NoError(t, err)
	require.Contains(t, url, GoogleAuthURL)
	require.Contains(t, url, "client_id=client-id")
	require.Contains(t, url, "code_challenge=challenge")
	require.Contains(t, url, "code_challenge_method=S256")
	require.Contains(t, url, "nonce=nonce123")
}

func TestAzureAuthorizeURL(t *testing.T) {
	url, err := BuildAuthorizeURL(
		"https://login.microsoftonline.com/tenant-id",
		"client-id",
		"http://localhost:3001/login",
		"openid profile email",
		"state123",
		"challenge",
		"nonce456",
	)
	require.NoError(t, err)
	require.Contains(t, url, "https://login.microsoftonline.com/tenant-id/oauth2/v2.0/authorize")
}
