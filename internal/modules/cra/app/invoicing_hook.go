package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
)

func (s *Service) WithInvoicePublisher(publisher ports.InvoiceDraftPublisher) *Service {
	if publisher != nil {
		s.invoices = publisher
	}
	return s
}

func (s *Service) publishValidationInvoice(ctx context.Context, ts domain.Timesheet) {
	if s.invoices == nil {
		return
	}
	clientID := s.resolveClientID(ctx, ts)
	if clientID == nil || *clientID == uuid.Nil {
		return
	}
	billableMinutes, err := s.BillableMinutesForMonth(ctx, ts.TenantID, ts.UserID, ts.Month)
	if err != nil || billableMinutes <= 0 {
		return
	}
	userLabel := userLabelForTimesheet(ctx, s, ts)
	_ = s.invoices.PublishCRAValidationDraft(ctx, ports.ValidationInvoiceCommand{
		TenantID:      ts.TenantID,
		TimesheetID:   ts.ID,
		ClientID:      *clientID,
		Month:         ts.Month,
		BillableHours: float64(billableMinutes) / 60,
		MissionLabel:  ts.CommercialInfo.Mission,
		UserLabel:     userLabel,
		Currency:      "EUR",
	})
}

func userLabelForTimesheet(ctx context.Context, s *Service, ts domain.Timesheet) string {
	summaries, err := s.repo.ListSummariesByTenantMonth(ctx, ts.TenantID, ts.Month)
	if err != nil {
		return ts.UserID.String()
	}
	for _, summary := range summaries {
		if summary.ID != ts.ID {
			continue
		}
		name := strings.TrimSpace(summary.UserPrenom + " " + summary.UserNom)
		if name != "" {
			return name
		}
		if summary.UserLogin != "" {
			return summary.UserLogin
		}
	}
	return ts.UserID.String()
}

func (s *Service) resolveClientID(ctx context.Context, ts domain.Timesheet) *uuid.UUID {
	if ts.CommercialInfo.ClientID != nil && *ts.CommercialInfo.ClientID != uuid.Nil {
		return ts.CommercialInfo.ClientID
	}
	summaries, err := s.repo.ListSummariesByTenantMonth(ctx, ts.TenantID, ts.Month)
	if err != nil {
		return nil
	}
	for _, summary := range summaries {
		if summary.ID == ts.ID {
			return summary.ClientID
		}
	}
	return nil
}
