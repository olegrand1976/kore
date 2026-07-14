package fec

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/kore/kore/pkg/kernel"
)

// Exporter produces FEC-compatible CSV exports (1st accounting connector).
type Exporter struct{}

func NewExporter() *Exporter {
	return &Exporter{}
}

// Export generates a minimal FEC CSV stub for the given tenant and period label.
func (e *Exporter) Export(_ context.Context, tenant kernel.TenantID, periodLabel string, recordCount int) ([]byte, int, error) {
	if recordCount <= 0 {
		recordCount = 1
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = '|'
	header := []string{
		"JournalCode", "JournalLib", "EcritureNum", "EcritureDate",
		"CompteNum", "CompteLib", "CompAuxNum", "CompAuxLib",
		"PieceRef", "PieceDate", "EcritureLib", "Debit", "Credit",
		"EcritureLet", "DateLet", "ValidDate", "Montantdevise", "Idevise",
	}
	if err := w.Write(header); err != nil {
		return nil, 0, err
	}
	now := time.Now().UTC().Format("20060102")
	for i := 0; i < recordCount; i++ {
		row := []string{
			"VT", "Ventes", fmt.Sprintf("%d", i+1), now,
			"411000", "Clients", "", "",
			fmt.Sprintf("FAC-%s-%d", tenant.String()[:8], i+1), now,
			fmt.Sprintf("Export FEC %s", periodLabel), "0,00", "100,00",
			"", "", now, "", "EUR",
		}
		if err := w.Write(row); err != nil {
			return nil, 0, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, 0, err
	}
	return buf.Bytes(), recordCount, nil
}
