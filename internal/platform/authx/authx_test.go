package authx

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

func TestRBACAuthorizer_UnionOfProfiles(t *testing.T) {
	perms := map[string]map[Module]map[Action]bool{
		"Collaborateur": {
			"cra": {ActionRead: true, ActionWrite: true},
		},
		"Chef d'équipe": {
			"cra": {ActionRead: true, ActionWrite: true, ActionValidate: true},
		},
	}
	authz := NewRBACAuthorizer(perms)
	ctx := WithIdentity(context.Background(), Identity{
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Profile:  "Collaborateur",
		Profiles: []Profile{"Collaborateur", "Chef d'équipe"},
	})
	if !authz.Can(ctx, "cra", ActionValidate) {
		t.Fatal("expected union of profiles to grant Validate from Chef d'équipe")
	}
}

func TestIdentity_EffectiveProfilesFallback(t *testing.T) {
	id := Identity{Profile: ProfileAdmin}
	got := id.EffectiveProfiles()
	if len(got) != 1 || got[0] != ProfileAdmin {
		t.Fatalf("got %v", got)
	}
}
