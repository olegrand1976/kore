//go:build integration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/invoicing/adapters/postgres"
	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestInvoicing_ProformaTokenRoundTrip(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	inv := domain.NewInvoice(tenant, uuid.New(), domain.InvoiceTypeStandard, "EUR")
	inv.Status = domain.InvoiceStatusPreparee
	inv.TotalAmount = 10000
	inv.TaxAmount = 2000
	now := time.Now().UTC()
	token := "test-proforma-token"
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])
	require.NoError(t, inv.EmitProforma(hash, "marie.dupont@acme.test", now))
	require.NoError(t, repo.SaveInvoice(ctx, inv))

	got, err := repo.GetInvoiceByProformaTokenHash(ctx, hash)
	require.NoError(t, err)
	require.Equal(t, inv.ID, got.ID)
	require.Equal(t, domain.InvoiceStatusProforma, got.Status)
	require.Equal(t, "marie.dupont@acme.test", got.ProformaRecipientEmail)

	require.NoError(t, got.ValidateProforma(now.Add(time.Minute)))
	got.MarkInvoiceSent(now.Add(2 * time.Minute))
	require.NoError(t, repo.SaveInvoice(ctx, got))

	reloaded, err := repo.GetInvoice(ctx, tenant, inv.ID)
	require.NoError(t, err)
	require.Equal(t, domain.InvoiceStatusPreparee, reloaded.Status)
	require.Empty(t, reloaded.ProformaTokenHash)
	require.NotNil(t, reloaded.ProformaValidatedAt)
	require.NotNil(t, reloaded.InvoiceSentAt)

	_, err = repo.GetInvoiceByProformaTokenHash(ctx, hash)
	require.ErrorIs(t, err, domain.ErrInvoiceNotFound)
}
