package project

import (
	"embed"

	"github.com/kore/kore/internal/platform/db"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrations() db.ModuleMigration {
	return db.ModuleMigration{
		Module: "project",
		Schema: "project",
		FS:     migrations,
		Dir:    "migrations",
	}
}
