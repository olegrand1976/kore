package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ModuleMigration describes migrations for one module/schema.
type ModuleMigration struct {
	Module string
	Schema string
	FS     embed.FS
	Dir    string
}

// MigrationRunner applies SQL migrations per schema in module order.
type MigrationRunner struct {
	pool    *Pool
	modules []ModuleMigration
}

func NewMigrationRunner(pool *Pool, modules []ModuleMigration) *MigrationRunner {
	return &MigrationRunner{pool: pool, modules: modules}
}

func (r *MigrationRunner) Up(ctx context.Context) error {
	if err := ensureRegistry(ctx, r.pool); err != nil {
		return err
	}
	for _, mod := range r.modules {
		if err := r.applyModule(ctx, mod); err != nil {
			return fmt.Errorf("module %s: %w", mod.Module, err)
		}
	}
	return nil
}

func (r *MigrationRunner) applyModule(ctx context.Context, mod ModuleMigration) error {
	files, err := fs.Glob(mod.FS, filepath.Join(mod.Dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		version := migrationVersion(file)
		applied, err := isApplied(ctx, r.pool, mod.Schema, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		content, err := mod.FS.ReadFile(file)
		if err != nil {
			return err
		}
		if err := r.pool.WithTx(ctx, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, string(content)); err != nil {
				return err
			}
			return markApplied(ctx, tx, mod.Schema, version, filepath.Base(file))
		}); err != nil {
			return fmt.Errorf("apply %s: %w", file, err)
		}
	}
	return nil
}

func migrationVersion(path string) string {
	base := filepath.Base(path)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) == 0 {
		return base
	}
	return parts[0]
}

func ensureRegistry(ctx context.Context, pool *Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS platform;
		CREATE TABLE IF NOT EXISTS platform.schema_migrations (
			schema_name TEXT NOT NULL,
			version TEXT NOT NULL,
			filename TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (schema_name, version)
		);
	`)
	return err
}

func isApplied(ctx context.Context, pool *Pool, schema, version string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM platform.schema_migrations
			WHERE schema_name = $1 AND version = $2
		)
	`, schema, version).Scan(&exists)
	return exists, err
}

func markApplied(ctx context.Context, tx pgx.Tx, schema, version, filename string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO platform.schema_migrations (schema_name, version, filename)
		VALUES ($1, $2, $3)
	`, schema, version, filename)
	return err
}
