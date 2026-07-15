package seed

import (
	"context"
	"log"

	"github.com/kore/kore/pkg/kernel"
)

// ResetDemoTenant supprime les données métier du tenant demo pour permettre un re-seed complet.
func (r *Runner) ResetDemoTenant(ctx context.Context) error {
	tid := DemoTenantID
	log.Println("seed: reset du tenant demo…")

	stmts := []struct {
		query string
		args  int
	}{
		{`DELETE FROM cra.time_lines WHERE tenant_id = $1`, 1},
		{`DELETE FROM cra.week_entries WHERE tenant_id = $1`, 1},
		{`DELETE FROM cra.timesheets WHERE tenant_id = $1`, 1},
		{`DELETE FROM conges.leave_requests WHERE tenant_id = $1`, 1},
		{`DELETE FROM conges.leave_balances WHERE tenant_id = $1`, 1},
		{`DELETE FROM conges.leave_type_configs WHERE tenant_id = $1`, 1},
		{`DELETE FROM budget.consumptions WHERE tenant_id = $1`, 1},
		{`DELETE FROM budget.quotes WHERE tenant_id = $1`, 1},
		{`DELETE FROM budget.estimates WHERE tenant_id = $1`, 1},
		{`DELETE FROM budget.budgets WHERE tenant_id = $1`, 1},
		{`DELETE FROM tma.analysis_dossiers WHERE tenant_id = $1`, 1},
		{`DELETE FROM tma.delivery_codes WHERE tenant_id = $1`, 1},
		{`DELETE FROM tma.releases WHERE tenant_id = $1`, 1},
		{`DELETE FROM tma.demands WHERE tenant_id = $1`, 1},
		{`DELETE FROM workflow.transition_logs WHERE tenant_id = $1`, 1},
		{`DELETE FROM workflow.instances WHERE tenant_id = $1`, 1},
		{`DELETE FROM workflow.transitions WHERE definition_id IN (SELECT id FROM workflow.definitions WHERE tenant_id = $1)`, 1},
		{`DELETE FROM workflow.states WHERE definition_id IN (SELECT id FROM workflow.definitions WHERE tenant_id = $1)`, 1},
		{`DELETE FROM workflow.definitions WHERE tenant_id = $1`, 1},
		{`DELETE FROM notifications.messages WHERE tenant_id = $1`, 1},
		{`DELETE FROM notifications.rules WHERE tenant_id = $1`, 1},
		{`DELETE FROM publicsite.appointments WHERE commercial_id IN (SELECT id FROM org.users WHERE tenant_id = $1)`, 1},
		{`DELETE FROM publicsite.booking_slots WHERE commercial_id IN (SELECT id FROM org.users WHERE tenant_id = $1)`, 1},
		{`DELETE FROM publicsite.leads WHERE email = 'demo@acme.test'`, 0},
		{`DELETE FROM org.users WHERE tenant_id = $1`, 1},
		{`DELETE FROM org.clients WHERE tenant_id = $1`, 1},
		{`DELETE FROM org.equipes WHERE tenant_id = $1`, 1},
		{`DELETE FROM org.applications WHERE tenant_id = $1`, 1},
		{`DELETE FROM org.services WHERE tenant_id = $1`, 1},
		{`DELETE FROM org.sites WHERE tenant_id = $1`, 1},
		{`DELETE FROM org.societes WHERE tenant_id = $1`, 1},
	}
	for _, stmt := range stmts {
		if stmt.args == 1 {
			if _, err := r.deps.Pool.Exec(ctx, stmt.query, tid); err != nil {
				return err
			}
			continue
		}
		if _, err := r.deps.Pool.Exec(ctx, stmt.query); err != nil {
			return err
		}
	}
	if r.deps.Cache != nil && r.deps.Keys != nil {
		prefix := r.deps.Keys.Key(kernel.NewTenantID(tid), "workflow", "def")
		if err := r.deps.Cache.DeleteByPrefix(ctx, prefix); err != nil {
			return err
		}
	}
	log.Println("seed: tenant demo réinitialisé")
	return nil
}
