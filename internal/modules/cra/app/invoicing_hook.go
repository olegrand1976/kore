package app

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	invoicingdomain "github.com/kore/kore/internal/modules/invoicing/domain"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
)

func (s *Service) tryPublishValidationInvoice(ctx context.Context, ts domain.Timesheet) ports.InvoiceDraftOutcome {
	tsID := ts.ID
	withTS := func(o ports.InvoiceDraftOutcome) ports.InvoiceDraftOutcome {
		o.TimesheetID = &tsID
		return o
	}
	if s.invoices == nil {
		return withTS(ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftUnavailable,
			Reason: "invoicing_not_configured",
		})
	}
	clientID := s.resolveClientID(ctx, ts)
	if clientID == nil || *clientID == uuid.Nil {
		return withTS(ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "client_unresolved",
		})
	}
	billableMinutes, err := s.BillableMinutesForMonth(ctx, ts.TenantID, ts.UserID, ts.Month)
	if err != nil {
		return withTS(ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "billable_hours_error",
		})
	}
	if billableMinutes <= 0 {
		return withTS(ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "no_billable_hours",
		})
	}
	if reason := s.billingModeSkipReason(ctx, ts); reason != "" {
		return withTS(ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: reason,
		})
	}
	userLabel := userLabelForTimesheet(ctx, s, ts)
	unitPrice, currency := s.resolveSellUnitPriceCents(ctx, ts)
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
		TaxRate:        20,
	})
	if err != nil {
		return withTS(mapPublishError(err))
	}
	if invoiceID == uuid.Nil {
		return withTS(ports.InvoiceDraftOutcome{
			Status: ports.InvoiceDraftSkipped,
			Reason: "already_exists_or_empty",
		})
	}
	return withTS(ports.InvoiceDraftOutcome{
		Status:    ports.InvoiceDraftCreated,
		InvoiceID: &invoiceID,
	})
}

func mapPublishError(err error) ports.InvoiceDraftOutcome {
	switch {
	case errors.Is(err, invoicingdomain.ErrInvoicingDisabled):
		return ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "invoicing_disabled"}
	case errors.Is(err, invoicingdomain.ErrAlreadyInvoiced):
		return ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "already_exists"}
	case errors.Is(err, invoicingdomain.ErrZeroUnitPrice):
		return ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "zero_unit_price"}
	case errors.Is(err, invoicingdomain.ErrNoBillableContent):
		return ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "no_billable_hours"}
	default:
		return ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "publish_failed"}
	}
}

func (s *Service) CreateInvoicesFromTimesheets(ctx context.Context, tenant kernel.TenantID, ids []uuid.UUID) ([]ports.InvoiceDraftOutcome, error) {
	out := make([]ports.InvoiceDraftOutcome, 0, len(ids))
	for _, id := range ids {
		idCopy := id
		ts, err := s.repo.GetByID(ctx, tenant, id)
		if err != nil {
			out = append(out, ports.InvoiceDraftOutcome{
				Status:      ports.InvoiceDraftSkipped,
				Reason:      "timesheet_not_found",
				TimesheetID: &idCopy,
			})
			continue
		}
		if ts.Status != domain.StatusDefinitif {
			out = append(out, ports.InvoiceDraftOutcome{
				Status:      ports.InvoiceDraftSkipped,
				Reason:      "not_definitive",
				TimesheetID: &idCopy,
			})
			continue
		}
		out = append(out, s.tryPublishValidationInvoice(ctx, ts))
	}
	return out, nil
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
	bestMinutes := -1
	for id, minutes := range minutesByMission {
		if minutes > bestMinutes || (minutes == bestMinutes && (bestID == "" || id < bestID)) {
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

func dominantApplicationFromLines(ts domain.Timesheet) uuid.UUID {
	minutesByApp := make(map[string]int)
	for _, week := range ts.Weeks {
		for _, line := range week.Lines {
			if line.Source.Type != "application" || !line.Billable || line.Duration.Minutes <= 0 {
				continue
			}
			minutesByApp[line.Source.ID] += line.Duration.Minutes
		}
	}
	var bestID string
	bestMinutes := -1
	for id, minutes := range minutesByApp {
		if minutes > bestMinutes || (minutes == bestMinutes && (bestID == "" || id < bestID)) {
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

func (s *Service) billingModeSkipReason(ctx context.Context, ts domain.Timesheet) string {
	if s.apps == nil {
		return ""
	}
	appID := dominantApplicationFromLines(ts)
	if appID == uuid.Nil {
		return ""
	}
	info, err := s.apps.GetApplicationBilling(ctx, ts.TenantID, appID)
	if err != nil {
		return ""
	}
	switch info.ModeFacturation {
	case orgdomain.ModeFacturationNon:
		return "billing_mode_disabled"
	case orgdomain.ModeFacturationForfait:
		return "billing_mode_forfait"
	default:
		return ""
	}
}

// resolveSellUnitPriceCents: mission TJM > application default TJM > société default TJM → hourly cents.
func (s *Service) resolveSellUnitPriceCents(ctx context.Context, ts domain.Timesheet) (int64, string) {
	cap := s.settingsForUser(ctx, ts.TenantID, ts.UserID).DayCapacityMinutes
	if cap <= 0 {
		cap = domain.DefaultDayCapacityMinutes
	}
	toHourly := func(tjm int64, currency string) (int64, string) {
		if tjm <= 0 {
			return 0, currency
		}
		if currency == "" {
			currency = "EUR"
		}
		return tjm * 60 / int64(cap), currency
	}

	if missionID := s.resolveMissionID(ctx, ts); missionID != nil && s.missions != nil {
		rate, err := s.missions.GetMissionRate(ctx, ts.TenantID, *missionID)
		if err == nil && rate.TJMAmount > 0 {
			return toHourly(rate.TJMAmount, rate.Currency)
		}
	}
	if s.apps != nil {
		if appID := dominantApplicationFromLines(ts); appID != uuid.Nil {
			info, err := s.apps.GetApplicationBilling(ctx, ts.TenantID, appID)
			if err == nil && info.DefaultTJMCents > 0 {
				return toHourly(info.DefaultTJMCents, "EUR")
			}
		}
	}
	settings := s.settingsForUser(ctx, ts.TenantID, ts.UserID)
	if settings.DefaultTJMCents > 0 {
		return toHourly(settings.DefaultTJMCents, "EUR")
	}
	return 0, "EUR"
}
