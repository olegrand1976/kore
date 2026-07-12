package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/billing/domain"
	"github.com/kore/kore/internal/modules/billing/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, s domain.Subscription) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO billing.subscriptions (
			id, tenant_id, stripe_customer_id, stripe_subscription_id, status, seats, current_period_end, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			stripe_customer_id = EXCLUDED.stripe_customer_id,
			stripe_subscription_id = EXCLUDED.stripe_subscription_id,
			status = EXCLUDED.status,
			seats = EXCLUDED.seats,
			current_period_end = EXCLUDED.current_period_end,
			updated_at = NOW()
	`, s.ID, s.TenantID.UUID(), nullIfEmpty(s.StripeCustomerID), nullIfEmpty(s.StripeSubscriptionID),
		string(s.Status), s.Seats, s.CurrentPeriodEnd)
	return err
}

func (r *Repository) GetByTenant(ctx context.Context, tenantID kernel.TenantID) (domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, COALESCE(stripe_customer_id, ''), COALESCE(stripe_subscription_id, ''),
		       status, seats, current_period_end
		FROM billing.subscriptions WHERE tenant_id = $1
	`, tenantID.UUID())
	s, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subscription{}, domain.ErrSubscriptionNotFound
		}
		return domain.Subscription{}, err
	}
	modules, err := r.ListEntitlements(ctx, tenantID)
	if err != nil {
		return domain.Subscription{}, err
	}
	s.Modules = modules
	return s, nil
}

func (r *Repository) GetByStripeCustomer(ctx context.Context, customerID string) (domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, COALESCE(stripe_customer_id, ''), COALESCE(stripe_subscription_id, ''),
		       status, seats, current_period_end
		FROM billing.subscriptions WHERE stripe_customer_id = $1
	`, customerID)
	s, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Subscription{}, domain.ErrSubscriptionNotFound
		}
		return domain.Subscription{}, err
	}
	modules, err := r.ListEntitlements(ctx, s.TenantID)
	if err != nil {
		return domain.Subscription{}, err
	}
	s.Modules = modules
	return s, nil
}

func (r *Repository) ListEntitlements(ctx context.Context, tenantID kernel.TenantID) ([]domain.ModuleEntitlement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tenant_id, module_code, enabled FROM billing.module_entitlements WHERE tenant_id = $1
	`, tenantID.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ModuleEntitlement
	for rows.Next() {
		var m domain.ModuleEntitlement
		var tid uuid.UUID
		var code string
		if err := rows.Scan(&tid, &code, &m.Enabled); err != nil {
			return nil, err
		}
		m.TenantID = kernel.NewTenantID(tid)
		m.ModuleCode = domain.ModuleCode(code)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) SaveEntitlements(ctx context.Context, tenantID kernel.TenantID, modules []domain.ModuleEntitlement) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM billing.module_entitlements WHERE tenant_id = $1`, tenantID.UUID()); err != nil {
			return err
		}
		for _, m := range modules {
			if _, err := tx.Exec(ctx, `
				INSERT INTO billing.module_entitlements (id, tenant_id, module_code, enabled)
				VALUES ($1, $2, $3, $4)
			`, uuid.New(), tenantID.UUID(), string(m.ModuleCode), m.Enabled); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) MarkEventProcessed(ctx context.Context, eventID, eventType string) (bool, error) {
	var firstTime bool
	err := r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM billing.webhook_events WHERE event_id = $1)`, eventID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			firstTime = false
			return nil
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO billing.webhook_events (event_id, event_type) VALUES ($1, $2)
		`, eventID, eventType); err != nil {
			return err
		}
		firstTime = true
		return nil
	})
	return firstTime, err
}

func scanSubscription(row pgx.Row) (domain.Subscription, error) {
	var s domain.Subscription
	var tenantID uuid.UUID
	var status string
	var customerID, subID string
	err := row.Scan(&s.ID, &tenantID, &customerID, &subID, &status, &s.Seats, &s.CurrentPeriodEnd)
	if err != nil {
		return domain.Subscription{}, err
	}
	s.TenantID = kernel.NewTenantID(tenantID)
	s.StripeCustomerID = customerID
	s.StripeSubscriptionID = subID
	s.Status = domain.SubscriptionStatus(status)
	return s, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

var _ ports.SubscriptionRepository = (*Repository)(nil)

// EnsureSubscription creates a trial subscription if none exists (dev/bootstrap).
func (r *Repository) EnsureTrial(ctx context.Context, tenantID kernel.TenantID, seats int, modules []domain.ModuleCode) error {
	_, err := r.GetByTenant(ctx, tenantID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, domain.ErrSubscriptionNotFound) {
		return err
	}
	end := time.Now().UTC().Add(14 * 24 * time.Hour)
	sub := domain.Subscription{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Status:           domain.StatusTrial,
		Seats:            seats,
		CurrentPeriodEnd: &end,
	}
	if err := r.Save(ctx, sub); err != nil {
		return err
	}
	entitlements := make([]domain.ModuleEntitlement, 0, len(modules))
	for _, code := range modules {
		entitlements = append(entitlements, domain.ModuleEntitlement{
			TenantID:   tenantID,
			ModuleCode: code,
			Enabled:    true,
		})
	}
	return r.SaveEntitlements(ctx, tenantID, entitlements)
}

func (r *Repository) GetSubscriptionStatus(ctx context.Context, tenantID kernel.TenantID) (domain.SubscriptionStatus, error) {
	var status string
	err := r.pool.QueryRow(ctx, `SELECT status FROM billing.subscriptions WHERE tenant_id = $1`, tenantID.UUID()).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrSubscriptionNotFound
		}
		return "", err
	}
	return domain.SubscriptionStatus(status), nil
}

func (r *Repository) UpdateStatus(ctx context.Context, tenantID kernel.TenantID, status domain.SubscriptionStatus, seats int, periodEnd *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE billing.subscriptions
		SET status = $2, seats = $3, current_period_end = $4, updated_at = NOW()
		WHERE tenant_id = $1
	`, tenantID.UUID(), string(status), seats, periodEnd)
	if err != nil {
		return fmt.Errorf("update subscription status: %w", err)
	}
	return nil
}
