package app

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
)

func (s *Service) tryPublishValidationInvoice(ctx context.Context, ts domain.Timesheet) ports.InvoiceDraftOutcome {
	if s.invoices == nil {
		return ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftUnavailable,
			Reason: "invoicing_not_configured",
		}
	}
	clientID := s.resolveClientID(ctx, ts)
	if clientID == nil || *clientID == uuid.Nil {
		return ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "client_unresolved",
		}
	}
	billableMinutes, err := s.BillableMinutesForMonth(ctx, ts.TenantID, ts.UserID, ts.Month)
	if err != nil {
		return ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "billable_hours_error",
		}
	}
	if billableMinutes <= 0 {
		return ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "no_billable_hours",
		}
	}
	userLabel := userLabelForTimesheet(ctx, s, ts)
	unitPrice := int64(0)
	currency := "EUR"
	if missionID := s.resolveMissionID(ctx, ts); missionID != nil {
		unitPrice, currency = s.resolveUnitPriceCents(ctx, ts, *missionID)
	}
	invoiceID, err := s.invoices.PublishCRAValidationDraft(ctx, ports.ValidationInvoiceCommand{
		TenantID:       ts.TenantID,
		TimesheetID:    ts.ID,
		ClientID:       *clientID,
		Month:          ts.Month,
		BillableHours:  float64(billableMinutes) / 60,
		MissionLabel:   ts.CommercialInfo.Mission,
		UserLabel:      userLabel,
		Currency:       currency,
		UnitPriceCents: unitPrice,
	})
	if err != nil {
		return ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "publish_failed",
		}
	}
	if invoiceID == uuid.Nil {
		return ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "already_exists_or_empty",
		}
	}
	return ports.InvoiceDraftOutcome{
		Status:    ports.InvoiceDraftCreated,
		InvoiceID: &invoiceID,
	}
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

func (s *Service) resolveMissionID(ctx context.Context, ts domain.Timesheet) *uuid.UUID {
	if ts.CommercialInfo.MissionID != nil && *ts.CommercialInfo.MissionID != uuid.Nil {
		return ts.CommercialInfo.MissionID
	}
	if id := dominantMissionFromLines(ts); id != uuid.Nil {
		return &id
	}
	summaries, err := s.repo.ListSummariesByTenantMonth(ctx, ts.TenantID, ts.Month)
	if err != nil {
		return nil
	}
	for _, summary := range summaries {
		if summary.ID == ts.ID {
			return summary.MissionID
		}
	}
	return nil
}

func dominantMissionFromLines(ts domain.Timesheet) uuid.UUID {
	minutesByMission := make(map[string]int)
	for _, week := range ts.Weeks {
		for _, line := range week.Lines {
			if line.Source.Type != "mission" || !line.Billable || line.Duration.Minutes <= 0 {
				continue
			}
			minutesByMission[line.Source.ID] += line.Duration.Minutes
		}
	}
	var bestID string
	var bestMinutes int
	for id, minutes := range minutesByMission {
		if minutes > bestMinutes {
			bestMinutes = minutes
			bestID = id
		}
	}
	if bestID == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(bestID)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func (s *Service) resolveUnitPriceCents(ctx context.Context, ts domain.Timesheet, missionID uuid.UUID) (int64, string) {
	if s.missions == nil {
		return 0, "EUR"
	}
	rate, err := s.missions.GetMissionRate(ctx, ts.TenantID, missionID)
	if err != nil || rate.TJMAmount <= 0 {
		return 0, "EUR"
	}
	cap := s.settingsForUser(ctx, ts.TenantID, ts.UserID).DayCapacityMinutes
	if cap <= 0 {
		cap = domain.DefaultDayCapacityMinutes
	}
	currency := rate.Currency
	if currency == "" {
		currency = "EUR"
	}
	return rate.TJMAmount * 60 / int64(cap), currency
}
