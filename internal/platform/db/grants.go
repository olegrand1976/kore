package db

import (
	"context"
	"embed"
	"fmt"
)

//go:embed grants_kore_app.sql
var grantsKoreAppSQL embed.FS

// GrantsKoreAppSQL returns the canonical kore_app grants script (for tests / ops).
func GrantsKoreAppSQL() (string, error) {
	b, err := grantsKoreAppSQL.ReadFile("grants_kore_app.sql")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ApplyKoreAppGrants grants USAGE/CRUD on module schemas to kore_app.
// No-op when the kore_app role does not exist (local docker).
func ApplyKoreAppGrants(ctx context.Context, pool *Pool) error {
	sql, err := GrantsKoreAppSQL()
	if err != nil {
		return fmt.Errorf("read kore_app grants: %w", err)
	}
	if _, err := pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("apply kore_app grants: %w", err)
	}
	return nil
}
