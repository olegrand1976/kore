package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kore/kore/internal/modules/notifications/domain"
	"github.com/kore/kore/internal/modules/notifications/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveRule(ctx context.Context, rule domain.NotificationRule) error {
	policy, err := json.Marshal(rule.RecipientsPolicy)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO notifications.rules (
			id, tenant_id, code, trigger, frequency, recipient_policy, template, attach_pdf
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id, code) DO UPDATE SET
			trigger = EXCLUDED.trigger,
			frequency = EXCLUDED.frequency,
			recipient_policy = EXCLUDED.recipient_policy,
			template = EXCLUDED.template,
			attach_pdf = EXCLUDED.attach_pdf
	`, rule.ID, rule.TenantID.UUID(), rule.Code, rule.Trigger, string(rule.Frequency), policy, rule.Template, rule.AttachPDF)
	return err
}

func (r *Repository) GetRuleByTrigger(ctx context.Context, tenant kernel.TenantID, trigger string) (domain.NotificationRule, error) {
	return r.scanRule(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, code, trigger, frequency, recipient_policy, template, attach_pdf
		FROM notifications.rules
		WHERE tenant_id = $1 AND trigger = $2
	`, tenant.UUID(), trigger))
}

func (r *Repository) ListRules(ctx context.Context, tenant kernel.TenantID) ([]domain.NotificationRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, code, trigger, frequency, recipient_policy, template, attach_pdf
		FROM notifications.rules
		WHERE tenant_id = $1
		ORDER BY code
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.NotificationRule
	for rows.Next() {
		rule, err := scanRuleRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (r *Repository) SaveMessage(ctx context.Context, m domain.NotificationMessage) error {
	recipients, err := json.Marshal(m.Recipients)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO notifications.messages (
			id, tenant_id, rule_code, recipients, subject, body, status, attempts, sent_at, scheduled_for
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			recipients = EXCLUDED.recipients,
			subject = EXCLUDED.subject,
			body = EXCLUDED.body,
			status = EXCLUDED.status,
			attempts = EXCLUDED.attempts,
			sent_at = EXCLUDED.sent_at,
			scheduled_for = EXCLUDED.scheduled_for
	`, m.ID, m.TenantID.UUID(), nullIfEmpty(m.RuleCode), recipients, m.Subject, m.Body, string(m.Status), m.Attempts, m.SentAt, m.ScheduledFor)
	return err
}

func (r *Repository) ListMessages(ctx context.Context, filter ports.SentFilter) ([]domain.NotificationMessage, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, rule_code, recipients, subject, body, status, attempts, sent_at, scheduled_for
		FROM notifications.messages
		WHERE tenant_id = $1
	`
	args := []any{filter.TenantID.UUID()}
	if filter.Status != nil {
		query += ` AND status = $2`
		args = append(args, string(*filter.Status))
		query += ` ORDER BY created_at DESC LIMIT $3`
		args = append(args, limit)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.NotificationMessage
	for rows.Next() {
		msg, err := scanMessageRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, rows.Err()
}

func (r *Repository) ListPending(ctx context.Context, tenant kernel.TenantID) ([]domain.NotificationMessage, error) {
	status := domain.MessageStatusPending
	return r.ListMessages(ctx, ports.SentFilter{
		TenantID: tenant,
		Status:   &status,
		Limit:    500,
	})
}

func (r *Repository) ListDue(ctx context.Context, now time.Time, limit int) ([]domain.NotificationMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, rule_code, recipients, subject, body, status, attempts, sent_at, scheduled_for
		FROM notifications.messages
		WHERE status = $1 AND (scheduled_for IS NULL OR scheduled_for <= $2)
		ORDER BY created_at ASC
		LIMIT $3
	`, string(domain.MessageStatusPending), now.UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.NotificationMessage
	for rows.Next() {
		msg, err := scanMessageRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, rows.Err()
}

func (r *Repository) scanRule(row pgx.Row) (domain.NotificationRule, error) {
	rule, err := scanRuleRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NotificationRule{}, domain.ErrRuleNotFound
		}
		return domain.NotificationRule{}, err
	}
	return rule, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanRuleRow(row scannable) (domain.NotificationRule, error) {
	var rule domain.NotificationRule
	var tenantID uuid.UUID
	var frequency string
	var policyJSON []byte
	err := row.Scan(
		&rule.ID, &tenantID, &rule.Code, &rule.Trigger, &frequency,
		&policyJSON, &rule.Template, &rule.AttachPDF,
	)
	if err != nil {
		return domain.NotificationRule{}, err
	}
	rule.TenantID = kernel.NewTenantID(tenantID)
	parsed, err := domain.ParseFrequency(frequency)
	if err != nil {
		return domain.NotificationRule{}, err
	}
	rule.Frequency = parsed
	if len(policyJSON) > 0 {
		if err := json.Unmarshal(policyJSON, &rule.RecipientsPolicy); err != nil {
			return domain.NotificationRule{}, fmt.Errorf("decode recipient policy: %w", err)
		}
	}
	return rule, nil
}

func scanMessageRow(row scannable) (domain.NotificationMessage, error) {
	var msg domain.NotificationMessage
	var tenantID uuid.UUID
	var ruleCode *string
	var recipientsJSON []byte
	var status string
	err := row.Scan(
		&msg.ID, &tenantID, &ruleCode, &recipientsJSON, &msg.Subject, &msg.Body,
		&status, &msg.Attempts, &msg.SentAt, &msg.ScheduledFor,
	)
	if err != nil {
		return domain.NotificationMessage{}, err
	}
	msg.TenantID = kernel.NewTenantID(tenantID)
	if ruleCode != nil {
		msg.RuleCode = *ruleCode
	}
	if len(recipientsJSON) > 0 {
		if err := json.Unmarshal(recipientsJSON, &msg.Recipients); err != nil {
			return domain.NotificationMessage{}, fmt.Errorf("decode recipients: %w", err)
		}
	}
	msg.Status = domain.MessageStatus(status)
	return msg, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

var _ ports.NotificationRepository = (*Repository)(nil)
