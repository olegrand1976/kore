package domain_test

import (
	"testing"

	"github.com/kore/kore/internal/modules/invoicing/domain"
	"github.com/stretchr/testify/require"
)

func TestLineNetCents(t *testing.T) {
	net, err := domain.LineNetCents(5000, 2)
	require.NoError(t, err)
	require.Equal(t, int64(10000), net)

	net, err = domain.LineNetCents(10000, 1.5)
	require.NoError(t, err)
	require.Equal(t, int64(15000), net)

	_, err = domain.LineNetCents(0, 1)
	require.ErrorIs(t, err, domain.ErrInvalidInvoiceLine)
	_, err = domain.LineNetCents(100, 0)
	require.ErrorIs(t, err, domain.ErrInvalidInvoiceLine)
}

func TestLineTaxCents(t *testing.T) {
	tax, err := domain.LineTaxCents(10000, 20)
	require.NoError(t, err)
	require.Equal(t, int64(2000), tax)
}
