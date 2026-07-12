package cra

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/budget/ports"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ReaderAdapter struct {
	reader craports.CRAReader
}

func NewReaderAdapter(reader craports.CRAReader) ports.CRAReader {
	return &ReaderAdapter{reader: reader}
}

func (a *ReaderAdapter) ConsumedByApplication(ctx context.Context, tenant kernel.TenantID, appID uuid.UUID, period kernel.Period) ([]ports.CRAConsumption, error) {
	items, err := a.reader.ConsumedByApplication(ctx, tenant, appID, period)
	if err != nil {
		return nil, err
	}
	var days float64
	for _, item := range items {
		days += float64(item.Duration.Minutes) / 480.0
	}
	if days == 0 {
		return nil, nil
	}
	return []ports.CRAConsumption{{Days: days}}, nil
}
