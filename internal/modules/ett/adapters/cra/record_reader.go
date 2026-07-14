package cra

import (
	"context"
	"time"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	ettports "github.com/kore/kore/internal/modules/ett/ports"
	"github.com/kore/kore/pkg/kernel"
)

type RecordReader struct {
	repo ettports.ETTRepository
}

func NewRecordReader(repo ettports.ETTRepository) craports.ETTRecordReader {
	return &RecordReader{repo: repo}
}

func (r *RecordReader) ListUserDayHours(
	ctx context.Context,
	tenant kernel.TenantID,
	userID uuid.UUID,
	from, to time.Time,
) ([]craports.ETTDayHours, error) {
	records, err := r.repo.ListRecords(ctx, ettports.RecordsQuery{
		TenantID: tenant,
		UserID:   &userID,
		From:     &from,
		To:       &to,
	})
	if err != nil {
		return nil, err
	}
	out := make([]craports.ETTDayHours, 0, len(records))
	for _, rec := range records {
		hours := rec.EffectiveHours
		if rec.ClockIn != nil && rec.ClockOut != nil {
			hours = rec.ClockOut.Sub(*rec.ClockIn).Hours()
		}
		if hours <= 0 {
			continue
		}
		out = append(out, craports.ETTDayHours{
			WorkDate: rec.WorkDate,
			Hours:    hours,
		})
	}
	return out, nil
}

var _ craports.ETTRecordReader = (*RecordReader)(nil)
