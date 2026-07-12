package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

func TestSubscriptionIsModuleEnabled(t *testing.T) {
	sub := Subscription{
		TenantID: kernel.NewTenantID(uuid.New()),
		Status:   StatusActive,
		Modules: []ModuleEntitlement{
			{ModuleCode: ModuleCRA, Enabled: true},
		},
	}
	if !sub.IsModuleEnabled(ModuleCRA) {
		t.Fatal("expected cra enabled")
	}
	sub.Status = StatusSuspended
	if sub.Status.AllowsAccess() {
		t.Fatal("suspended must not allow access")
	}
}

func TestSubscriptionStatusAllowsAccess(t *testing.T) {
	cases := []struct {
		status SubscriptionStatus
		want   bool
	}{
		{StatusTrial, true},
		{StatusActive, true},
		{StatusPastDue, true},
		{StatusCanceled, false},
		{StatusSuspended, false},
	}
	for _, tc := range cases {
		if tc.status.AllowsAccess() != tc.want {
			t.Fatalf("status %s access=%v want %v", tc.status, tc.status.AllowsAccess(), tc.want)
		}
	}
}
