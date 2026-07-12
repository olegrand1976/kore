package tma

import (
	"embed"

	"github.com/kore/kore/internal/platform/db"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrations() db.ModuleMigration {
	return db.ModuleMigration{
		Module: "tma",
		Schema: "tma",
		FS:     migrations,
		Dir:    "migrations",
	}
}
