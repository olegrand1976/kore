package app

import (
	"context"
	"fmt"

	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

func (s *Service) ExportPrestationsXML(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]ports.PrestationExportRow, error) {
	summaries, err := s.ListPrestations(ctx, tenant, month)
	if err != nil {
		return nil, err
	}
	out := make([]ports.PrestationExportRow, 0, len(summaries))
	for _, summary := range summaries {
		billableMinutes, err := s.BillableMinutesForMonth(ctx, tenant, summary.UserID, month)
		if err != nil {
			billableMinutes = 0
		}
		name := summary.UserPrenom
		if summary.UserNom != "" {
			if name != "" {
				name += " "
			}
			name += summary.UserNom
		}
		out = append(out, ports.PrestationExportRow{
			UserLogin:     summary.UserLogin,
			UserName:      name,
			Month:         string(month),
			Status:        string(summary.Status),
			TotalHours:    float64(summary.TotalMinutes) / 60,
			BillableHours: float64(billableMinutes) / 60,
			WeeksRatio:    fmt.Sprintf("%d/%d", summary.WeeksSubmitted, summary.WeeksTotal),
		})
	}
	return out, nil
}

func (s *Service) BillableSummary(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]ports.BillableUserSummary, error) {
	summaries, err := s.ListPrestations(ctx, tenant, month)
	if err != nil {
		return nil, err
	}
	out := make([]ports.BillableUserSummary, 0, len(summaries))
	for _, summary := range summaries {
		billableMinutes, err := s.BillableMinutesForMonth(ctx, tenant, summary.UserID, month)
		if err != nil || billableMinutes <= 0 {
			continue
		}
		out = append(out, ports.BillableUserSummary{
			UserID:        summary.UserID,
			UserLogin:     summary.UserLogin,
			BillableHours: float64(billableMinutes) / 60,
		})
	}
	return out, nil
}
