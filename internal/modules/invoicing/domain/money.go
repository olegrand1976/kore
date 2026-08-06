package domain

import "math"

// LineNetCents computes line HT in cents from unit price (cents) and quantity
// using milli-units to avoid float accumulation drift.
func LineNetCents(unitPriceCents int64, quantity float64) (int64, error) {
	if unitPriceCents <= 0 || quantity <= 0 || math.IsNaN(quantity) || math.IsInf(quantity, 0) {
		return 0, ErrInvalidInvoiceLine
	}
	milliQty := int64(math.Round(quantity * 1000))
	if milliQty <= 0 {
		return 0, ErrInvalidInvoiceLine
	}
	return unitPriceCents * milliQty / 1000, nil
}

// LineTaxCents computes VAT cents from net amount and tax rate percent (e.g. 20).
func LineTaxCents(netCents int64, taxRatePercent float64) (int64, error) {
	if netCents < 0 || taxRatePercent < 0 || math.IsNaN(taxRatePercent) || math.IsInf(taxRatePercent, 0) {
		return 0, ErrInvalidInvoiceLine
	}
	// Basis points of percent: 20% → 2000; supports two decimal places on the rate.
	rateBP := int64(math.Round(taxRatePercent * 100))
	return netCents * rateBP / 10000, nil
}
