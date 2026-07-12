package billing

import (
	"embed"

	"github.com/kore/kore/internal/platform/db"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrations() db.ModuleMigration {
	return db.ModuleMigration{
		Module: "billing",
		Schema: "billing",
		FS:     migrations,
		Dir:    "migrations",
	}
}
