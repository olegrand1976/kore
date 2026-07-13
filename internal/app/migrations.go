package app

import (
	"github.com/kore/kore/internal/modules/admin"
	"github.com/kore/kore/internal/modules/ai"
	"github.com/kore/kore/internal/modules/billing"
	"github.com/kore/kore/internal/modules/budget"
	"github.com/kore/kore/internal/modules/conges"
	"github.com/kore/kore/internal/modules/cra"
	"github.com/kore/kore/internal/modules/ett"
	"github.com/kore/kore/internal/modules/integrations"
	"github.com/kore/kore/internal/modules/invoicing"
	"github.com/kore/kore/internal/modules/maintenance"
	"github.com/kore/kore/internal/modules/notifications"
	"github.com/kore/kore/internal/modules/org"
	"github.com/kore/kore/internal/modules/publicsite"
	"github.com/kore/kore/internal/modules/reporting"
	"github.com/kore/kore/internal/modules/ssii"
	"github.com/kore/kore/internal/modules/support"
	"github.com/kore/kore/internal/modules/tma"
	"github.com/kore/kore/internal/modules/workflow"
	"github.com/kore/kore/internal/platform/db"
)

func AllModuleMigrations() []db.ModuleMigration {
	return []db.ModuleMigration{
		org.Migrations(),
		workflow.Migrations(),
		cra.Migrations(),
		notifications.Migrations(),
		conges.Migrations(),
		budget.Migrations(),
		tma.Migrations(),
		ssii.Migrations(),
		support.Migrations(),
		maintenance.Migrations(),
		invoicing.Migrations(),
		ett.Migrations(),
		reporting.Migrations(),
		admin.Migrations(),
		integrations.Migrations(),
		ai.Migrations(),
		billing.Migrations(),
		publicsite.Migrations(),
	}
}
