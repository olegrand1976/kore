package postgres

import (
	"context"

	"github.com/kore/kore/internal/platform/db"
)

type craSchema struct {
	hasRejectReason bool
	hasLineBillable bool
	hasLineOrigin   bool
	hasSSMissions   bool
}

func probeCraSchema(ctx context.Context, pool *db.Pool) craSchema {
	return craSchema{
		hasRejectReason: columnExists(ctx, pool, "cra", "timesheets", "reject_reason"),
		hasLineBillable: columnExists(ctx, pool, "cra", "time_lines", "billable"),
		hasLineOrigin:   columnExists(ctx, pool, "cra", "time_lines", "origin"),
		hasSSMissions:   tableExists(ctx, pool, "ssii.missions"),
	}
}

func columnExists(ctx context.Context, pool *db.Pool, schema, table, column string) bool {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = $2 AND column_name = $3
		)
	`, schema, table, column).Scan(&exists)
	return err == nil && exists
}

func tableExists(ctx context.Context, pool *db.Pool, qualified string) bool {
	var exists bool
	err := pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, qualified).Scan(&exists)
	return err == nil && exists
}
