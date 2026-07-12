package app_test

import (
	"testing"

	"github.com/kore/kore/internal/modules/org/app"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/stretchr/testify/require"
)

func TestArgon2HashVerify(t *testing.T) {
	hasher := app.NewArgon2Hasher()
	hash, err := hasher.Hash("secret")
	require.NoError(t, err)
	require.True(t, hasher.Verify(hash, "secret"))
	require.False(t, hasher.Verify(hash, "wrong"))
}

func TestDefaultPermissionsAdmin(t *testing.T) {
	perms := app.DefaultPermissions()
	require.True(t, perms[string(domain.ProfileAdmin)]["org"]["E"])
}
