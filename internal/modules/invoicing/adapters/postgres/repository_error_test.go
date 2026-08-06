package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kore/kore/internal/modules/invoicing/domain"
)

func TestMapInvoiceSaveError_UniqueSourceTimesheet(t *testing.T) {
	err := mapInvoiceSaveError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "idx_invoicing_invoices_source_timesheet",
	})
	if !errors.Is(err, domain.ErrAlreadyInvoiced) {
		t.Fatalf("expected ErrAlreadyInvoiced, got %v", err)
	}
}

func TestMapInvoiceSaveError_OtherUnique(t *testing.T) {
	raw := &pgconn.PgError{Code: "23505", ConstraintName: "invoices_pkey"}
	err := mapInvoiceSaveError(raw)
	if errors.Is(err, domain.ErrAlreadyInvoiced) {
		t.Fatal("pkey unique should not map to ErrAlreadyInvoiced")
	}
	if err != raw {
		t.Fatalf("expected original error, got %v", err)
	}
}
