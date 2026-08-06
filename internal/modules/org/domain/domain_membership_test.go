package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestSyncPrimaryMemberships_keepsEquipeWhenStillMember(t *testing.T) {
	primary := uuid.New()
	other := uuid.New()
	// Deliberately put other first (UUID-sort style) to ensure we keep primary.
	u := User{
		Profile:   ProfileCollaborateur,
		Profiles:  []Profile{ProfileCollaborateur, ProfileChefEquipe},
		EquipeID:  &primary,
		EquipeIDs: []uuid.UUID{other, primary},
	}
	u.SyncPrimaryMemberships()
	if u.EquipeID == nil || *u.EquipeID != primary {
		t.Fatalf("EquipeID = %v, want primary %v", u.EquipeID, primary)
	}
	if u.Profile != ProfileChefEquipe {
		t.Fatalf("Profile = %s, want Chef d'équipe as primary among profiles", u.Profile)
	}
}

func TestSyncPrimaryMemberships_detachesEmptySlice(t *testing.T) {
	primary := uuid.New()
	u := User{
		Profile:   ProfileAdmin,
		EquipeID:  &primary,
		EquipeIDs: []uuid.UUID{},
	}
	u.SyncPrimaryMemberships()
	if u.EquipeID != nil {
		t.Fatalf("expected detach, got %v", u.EquipeID)
	}
}

func TestSyncPrimaryMemberships_setsPrimaryFromEquipeIDsWhenNil(t *testing.T) {
	a := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	b := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	u := User{
		EquipeIDs: []uuid.UUID{a, b},
	}
	u.SyncPrimaryMemberships()
	if u.EquipeID == nil || *u.EquipeID != a {
		t.Fatalf("EquipeID = %v, want %v", u.EquipeID, a)
	}
}

func TestValidateProfiles(t *testing.T) {
	if err := ValidateProfiles(nil); err != ErrProfilesRequired {
		t.Fatalf("nil: %v", err)
	}
	if err := ValidateProfiles([]Profile{"Inconnu"}); err != ErrInvalidProfile {
		t.Fatalf("unknown: %v", err)
	}
	if err := ValidateProfiles([]Profile{ProfileCollaborateur}); err != nil {
		t.Fatalf("valid: %v", err)
	}
}
