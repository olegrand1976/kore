package fec

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

func TestExportProducesFECRows(t *testing.T) {
	exp := NewExporter()
	data, count, err := exp.Export(context.Background(), kernel.NewTenantID(uuid.New()), "2026-Q1", 2)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 records, got %d", count)
	}
	csv := string(data)
	if !strings.Contains(csv, "JournalCode") {
		t.Fatal("expected FEC header")
	}
	if strings.Count(csv, "\n") < 3 {
		t.Fatalf("expected header + 2 rows, got:\n%s", csv)
	}
}
