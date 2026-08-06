//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/adapters/postgres"
	invoicingapp "github.com/kore/kore/internal/modules/invoicing/app"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/modules/invoicing/ports"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestInvoicing_CreateFromCRA_SourceTimesheetRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	svc := invoicingapp.NewService(repo)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	clientID := uuid.New()
	timesheetID := uuid.New()

	inv, err := svc.CreateFromCRAValidation(ctx, ports.CreateFromCRACommand{
		TenantID:       tenant,
		TimesheetID:    timesheetID,
		ClientID:       clientID,
		Month:          "2026-08",
		BillableHours:  8,
		MissionLabel:   "Mission test",
		UserLabel:      "Collab",
		Currency:       "EUR",
		UnitPriceCents: 10000,
		TaxRate:        20,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, inv.ID)
	require.NotNil(t, inv.SourceTimesheetID)
	require.Equal(t, timesheetID, *inv.SourceTimesheetID)

	exists, err := repo.InvoiceExistsForTimesheet(ctx, tenant, timesheetID)
	require.NoError(t, err)
	require.True(t, exists)

	got, err := svc.Get(ctx, tenant, inv.ID)
	require.NoError(t, err)
	require.Len(t, got.Lines, 1)
	require.NotNil(t, got.SourceTimesheetID)
	require.Equal(t, timesheetID, *got.SourceTimesheetID)

	_, err = svc.CreateFromCRAValidation(ctx, ports.CreateFromCRACommand{
		TenantID:       tenant,
		TimesheetID:    timesheetID,
		ClientID:       clientID,
		Month:          "2026-08",
		BillableHours:  8,
		UnitPriceCents: 10000,
	})
	require.ErrorIs(t, err, domain.ErrAlreadyInvoiced)

	// Unique index also protects concurrent insert after check bypass.
	dup := domain.NewInvoice(tenant, clientID, domain.InvoiceTypeStandard, "EUR")
	dup.SourceTimesheetID = &timesheetID
	dup.Status = domain.InvoiceStatusPreparee
	dup.CreatedAt = time.Now().UTC()
	err = repo.SaveInvoiceWithLines(ctx, dup, nil)
	require.ErrorIs(t, err, domain.ErrAlreadyInvoiced)
}
