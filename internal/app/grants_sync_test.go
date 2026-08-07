package app_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kore/kore/internal/app"
	"github.com/kore/kore/internal/platform/db"
)

func TestKoreAppGrantsCoverAllModuleSchemas(t *testing.T) {
	sql, err := db.GrantsKoreAppSQL()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range app.AllModuleMigrations() {
		needle := fmt.Sprintf("'%s'", m.Schema)
		if !strings.Contains(sql, needle) {
			t.Errorf("grants_kore_app.sql missing schema %q (module %s)", m.Schema, m.Module)
		}
	}
	for _, sch := range []string{"public", "platform"} {
		needle := fmt.Sprintf("'%s'", sch)
		if !strings.Contains(sql, needle) {
			t.Errorf("grants_kore_app.sql missing base schema %q", sch)
		}
	}
	if !strings.Contains(sql, "FOR ROLE") {
		t.Error("grants_kore_app.sql must set ALTER DEFAULT PRIVILEGES FOR ROLE")
	}
	if !strings.Contains(sql, "ett.audit_journal") {
		t.Error("grants_kore_app.sql must revoke UPDATE/DELETE on ett.audit_journal")
	}
}
