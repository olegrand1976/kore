package app_test

import (
	"testing"

	"github.com/kore/kore/internal/modules/org/app"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/platform/authx"
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
	require.True(t, perms[string(domain.ProfileAdmin)]["org"][authx.ActionWrite])
	require.True(t, perms[string(domain.ProfileAdmin)]["conges"][authx.ActionValidate])
	require.True(t, perms[string(domain.ProfileAdmin)]["budget"][authx.ActionRead])
}

func TestDefaultPermissionsResponsableCongesValidate(t *testing.T) {
	perms := app.DefaultPermissions()
	require.True(t, perms["Responsable de service"]["org"][authx.ActionRead])
	require.True(t, perms["Responsable de service"]["conges"][authx.ActionValidate])
	require.True(t, perms["Responsable de service"]["budget"][authx.ActionWrite])
}

func TestDefaultPermissionsChefEquipeOrgRead(t *testing.T) {
	perms := app.DefaultPermissions()
	require.True(t, perms["Chef d'équipe"]["org"][authx.ActionRead])
	require.False(t, perms["Chef d'équipe"]["conges"][authx.ActionValidate])
}

func TestDefaultPermissionsCollaborateur(t *testing.T) {
	perms := app.DefaultPermissions()
	require.True(t, perms[string(domain.ProfileCollaborateur)]["conges"][authx.ActionWrite])
	require.True(t, perms[string(domain.ProfileCollaborateur)]["tma"][authx.ActionWrite])
}
