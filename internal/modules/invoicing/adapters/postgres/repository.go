package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/pkg/kernel"
)

type Repository struct {
	pool *db.Pool
}

func NewRepository(pool *db.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SaveInvoice(ctx context.Context, inv domain.Invoice) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invoicing.invoices (
			id, tenant_id, client_id, type, status, currency,
			total_amount, tax_amount, pdp_receipt_id, transmitted_at, created_at, source_timesheet_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			total_amount = EXCLUDED.total_amount,
			tax_amount = EXCLUDED.tax_amount,
			pdp_receipt_id = EXCLUDED.pdp_receipt_id,
			transmitted_at = EXCLUDED.transmitted_at,
			source_timesheet_id = COALESCE(EXCLUDED.source_timesheet_id, invoicing.invoices.source_timesheet_id)
	`, inv.ID, inv.TenantID.UUID(), inv.ClientID, string(inv.Type), string(inv.Status),
		inv.Currency, inv.TotalAmount, inv.TaxAmount, inv.PDPReceiptID, inv.TransmittedAt, inv.CreatedAt,
		inv.SourceTimesheetID)
	return mapInvoiceSaveError(err)
}

func mapInvoiceSaveError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "source_timesheet") {
			return domain.ErrAlreadyInvoiced
		}
	}
	return err
}

func (r *Repository) GetInvoice(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Invoice, error) {
	return r.scanInvoice(r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, client_id, type, status, currency,
			total_amount, tax_amount, pdp_receipt_id, transmitted_at, created_at, source_timesheet_id
		FROM invoicing.invoices WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), id))
}

func (r *Repository) ListInvoices(ctx context.Context, tenant kernel.TenantID) ([]domain.Invoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, client_id, type, status, currency,
			total_amount, tax_amount, pdp_receipt_id, transmitted_at, created_at, source_timesheet_id
		FROM invoicing.invoices WHERE tenant_id = $1 ORDER BY created_at DESC
	`, tenant.UUID())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Invoice
	for rows.Next() {
		inv, err := r.scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (r *Repository) SaveInvoiceLine(ctx context.Context, line domain.InvoiceLine) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invoicing.invoice_lines (
			id, tenant_id, invoice_id, description, quantity, unit_price, tax_rate
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, line.ID, line.TenantID.UUID(), line.InvoiceID, line.Description,
		line.Quantity, line.UnitPrice, line.TaxRate)
	return err
}

func (r *Repository) SaveInvoiceWithLines(ctx context.Context, inv domain.Invoice, lines []domain.InvoiceLine) error {
	return r.pool.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO invoicing.invoices (
				id, tenant_id, client_id, type, status, currency,
				total_amount, tax_amount, pdp_receipt_id, transmitted_at, created_at, source_timesheet_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
				status = EXCLUDED.status,
				total_amount = EXCLUDED.total_amount,
				tax_amount = EXCLUDED.tax_amount,
				pdp_receipt_id = EXCLUDED.pdp_receipt_id,
				transmitted_at = EXCLUDED.transmitted_at,
				source_timesheet_id = COALESCE(EXCLUDED.source_timesheet_id, invoicing.invoices.source_timesheet_id)
		`, inv.ID, inv.TenantID.UUID(), inv.ClientID, string(inv.Type), string(inv.Status),
			inv.Currency, inv.TotalAmount, inv.TaxAmount, inv.PDPReceiptID, inv.TransmittedAt, inv.CreatedAt,
			inv.SourceTimesheetID)
		if err != nil {
			return mapInvoiceSaveError(err)
		}
		for _, line := range lines {
			_, err := tx.Exec(ctx, `
				INSERT INTO invoicing.invoice_lines (
					id, tenant_id, invoice_id, description, quantity, unit_price, tax_rate
				) VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, line.ID, line.TenantID.UUID(), line.InvoiceID, line.Description,
				line.Quantity, line.UnitPrice, line.TaxRate)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListInvoiceLines(ctx context.Context, tenant kernel.TenantID, invoiceID uuid.UUID) ([]domain.InvoiceLine, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, invoice_id, description, quantity, unit_price, tax_rate
		FROM invoicing.invoice_lines WHERE tenant_id = $1 AND invoice_id = $2
	`, tenant.UUID(), invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.InvoiceLine
	for rows.Next() {
		var line domain.InvoiceLine
		var tenantID uuid.UUID
		if err := rows.Scan(&line.ID, &tenantID, &line.InvoiceID, &line.Description,
			&line.Quantity, &line.UnitPrice, &line.TaxRate); err != nil {
			return nil, err
		}
		line.TenantID = kernel.NewTenantID(tenantID)
		out = append(out, line)
	}
	return out, rows.Err()
}

func (r *Repository) SavePDPQueueItem(ctx context.Context, item domain.PDPQueueItem) error {
	payload, err := json.Marshal(item.Payload)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO invoicing.pdp_queue (
			id, tenant_id, invoice_id, payload, status, attempts, last_error, created_at, next_retry_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, item.ID, item.TenantID.UUID(), item.InvoiceID, payload,
		item.Status, item.Attempts, item.LastError, item.CreatedAt, item.NextRetryAt)
	return err
}

func (r *Repository) InvoiceExistsForTimesheet(ctx context.Context, tenant kernel.TenantID, timesheetID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM invoicing.invoices
			WHERE tenant_id = $1 AND source_timesheet_id = $2
		)
	`, tenant.UUID(), timesheetID).Scan(&exists)
	return exists, err
}

func (r *Repository) SumNonVirtualInvoicesInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) (int64, int, string, error) {
	var total int64
	var count int
	var currency string
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_amount), 0)::bigint,
		       COUNT(*)::int,
		       COALESCE(NULLIF(MAX(currency), ''), 'EUR')
		FROM invoicing.invoices
		WHERE tenant_id = $1
		  AND status <> $2
		  AND created_at >= $3
		  AND created_at < ($4::timestamptz + INTERVAL '1 day')
	`, tenant.UUID(), string(domain.InvoiceStatusVirtuelle), period.Start, period.End).Scan(&total, &count, &currency)
	return total, count, currency, err
}

func (r *Repository) scanInvoice(row pgx.Row) (domain.Invoice, error) {
	var inv domain.Invoice
	var tenantID uuid.UUID
	var invType, status string
	err := row.Scan(&inv.ID, &tenantID, &inv.ClientID, &invType, &status, &inv.Currency,
		&inv.TotalAmount, &inv.TaxAmount, &inv.PDPReceiptID, &inv.TransmittedAt, &inv.CreatedAt, &inv.SourceTimesheetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Invoice{}, domain.ErrInvoiceNotFound
		}
		return domain.Invoice{}, err
	}
	inv.TenantID = kernel.NewTenantID(tenantID)
	inv.Type = domain.InvoiceType(invType)
	inv.Status = domain.InvoiceStatus(status)
	return inv, nil
}

var _ ports.InvoicingRepository = (*Repository)(nil)
