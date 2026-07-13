package postgres

import (
	"context"
	"time"

	"github.com/kore/kore/internal/modules/org/ports"
)

func (r *Repository) ListTenantsUsage(ctx context.Context) ([]ports.TenantUsageSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id,
			t.name,
			t.created_at,
			COALESCE((
				SELECT s.raison_sociale FROM org.societes s
				WHERE s.tenant_id = t.id ORDER BY s.created_at LIMIT 1
			), t.name) AS societe_name,
			COALESCE(bs.status, 'none') AS subscription_status,
			COALESCE(bs.seats, 0) AS seat_limit,
			COALESCE(uc.cnt, 0) AS active_users,
			COALESCE(mc.cnt, 0) AS modules_enabled,
			COALESCE(cra.cnt, 0) AS cra_count,
			COALESCE(tma.cnt, 0) AS tma_count,
			COALESCE(tma.open_cnt, 0) AS tma_open,
			COALESCE(bud.cnt, 0) AS budget_count,
			COALESCE(cong.cnt, 0) AS leave_count,
			COALESCE(ai.cnt, 0) AS ai_requests_30d,
			GREATEST(
				t.created_at,
				COALESCE(cra.last_at, t.created_at),
				COALESCE(tma.last_at, t.created_at),
				COALESCE(ai.last_at, t.created_at)
			) AS last_activity_at
		FROM org.tenants t
		LEFT JOIN billing.subscriptions bs ON bs.tenant_id = t.id
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt FROM org.users u
			WHERE u.tenant_id = t.id AND u.active = TRUE AND u.deleted_at IS NULL
		) uc ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt FROM billing.module_entitlements me
			WHERE me.tenant_id = t.id AND me.enabled = TRUE
		) mc ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt, MAX(updated_at) AS last_at
			FROM cra.timesheets ts WHERE ts.tenant_id = t.id
		) cra ON TRUE
		LEFT JOIN LATERAL (
			SELECT
				COUNT(*) AS cnt,
				COUNT(*) FILTER (WHERE status IN ('ouverte', 'affectee', 'en_cours', 'rework')) AS open_cnt,
				MAX(created_at) AS last_at
			FROM tma.demands d WHERE d.tenant_id = t.id
		) tma ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt FROM budget.budgets b WHERE b.tenant_id = t.id
		) bud ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt FROM conges.leave_requests lr WHERE lr.tenant_id = t.id
		) cong ON TRUE
		LEFT JOIN LATERAL (
			SELECT COUNT(*) AS cnt, MAX(created_at) AS last_at
			FROM ai.ai_request_log ar
			WHERE ar.tenant_id = t.id AND ar.created_at >= NOW() - INTERVAL '30 days'
		) ai ON TRUE
		ORDER BY t.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cutoff := time.Now().UTC().Add(-30 * 24 * time.Hour)
	out := make([]ports.TenantUsageSummary, 0)
	for rows.Next() {
		var item ports.TenantUsageSummary
		var lastActivity time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.CreatedAt,
			&item.SocieteName,
			&item.SubscriptionStatus,
			&item.SeatLimit,
			&item.ActiveUsers,
			&item.ModulesEnabled,
			&item.CraCount,
			&item.TmaCount,
			&item.TmaOpen,
			&item.BudgetCount,
			&item.LeaveCount,
			&item.AIRequests30d,
			&lastActivity,
		); err != nil {
			return nil, err
		}
		if item.SeatLimit > 0 {
			item.SeatUsagePct = float64(item.ActiveUsers) / float64(item.SeatLimit) * 100
		}
		item.LastActivityAt = &lastActivity
		item.ActiveLast30d = !lastActivity.Before(cutoff)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

var _ ports.PlatformRepository = (*Repository)(nil)
