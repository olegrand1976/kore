package app

import (
	"github.com/kore/kore/internal/modules/billing"
	"github.com/kore/kore/internal/modules/budget"
	"github.com/kore/kore/internal/modules/conges"
	"github.com/kore/kore/internal/modules/cra"
	"github.com/kore/kore/internal/modules/ai"
	"github.com/kore/kore/internal/modules/notifications"
	"github.com/kore/kore/internal/modules/org"
	"github.com/kore/kore/internal/modules/publicsite"
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
		ai.Migrations(),
		billing.Migrations(),
		publicsite.Migrations(),
	}
}
