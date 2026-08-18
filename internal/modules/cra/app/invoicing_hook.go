package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	invoicingdomain "github.com/kore/kore/internal/modules/invoicing/domain"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
)

const defaultInvoiceTaxRate = 20.0

func (s *Service) rejectIfTimesheetInvoiced(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) error {
	if s.invoices == nil {
		return nil
	}
	exists, err := s.invoices.TimesheetAlreadyInvoiced(ctx, tenant, id)
	if err != nil {
		if errors.Is(err, invoicingdomain.ErrInvoicingDisabled) {
			return nil
		}
		return err
	}
	if exists {
		return domain.ErrCRAAlreadyInvoiced
	}
	return nil
}

// hardInvoiceBlockers cannot be cleared via create-invoices overrides (fail-closed billing rules).
var hardInvoiceBlockers = map[string]struct{}{
	"already_exists":              {},
	"already_exists_check_failed": {},
	"invoicing_not_configured":    {},
	"invoicing_disabled":          {},
	"billing_mode_disabled":       {},
	"billing_mode_forfait":        {},
	"billing_mode_unresolved":     {},
	"billable_hours_error":        {},
	"not_definitive":              {},
	"timesheet_not_found":         {},
}

func firstHardInvoiceBlocker(blockers []string) string {
	for _, b := range blockers {
		if _, ok := hardInvoiceBlockers[b]; ok {
			return b
		}
	}
	return ""
}

// buildInvoiceDraftPreview resolves CRA→invoice fields without persisting.
func (s *Service) buildInvoiceDraftPreview(ctx context.Context, ts domain.Timesheet) ports.InvoiceDraftPreview {
	preview := ports.InvoiceDraftPreview{TimesheetID: ts.ID}
	addBlocker := func(reason string) {
		preview.OK = false
		preview.Blockers = append(preview.Blockers, reason)
	}

	if s.invoices == nil {
		addBlocker("invoicing_not_configured")
		return preview
	}

	if exists, err := s.invoices.TimesheetAlreadyInvoiced(ctx, ts.TenantID, ts.ID); err != nil {
		if errors.Is(err, invoicingdomain.ErrInvoicingDisabled) {
			addBlocker("invoicing_disabled")
		} else {
			addBlocker("already_exists_check_failed")
		}
		return preview
	} else if exists {
		addBlocker("already_exists")
		return preview
	}

	clientID := s.resolveClientID(ctx, ts)
	if clientID == nil || *clientID == uuid.Nil {
		addBlocker("client_unresolved")
	} else {
		preview.ClientID = clientID
	}

	billableMinutes, skipReason, err := s.invoiceableBillableMinutes(ctx, ts)
	if err != nil {
		addBlocker("billable_hours_error")
	} else if skipReason != "" {
		addBlocker(skipReason)
	} else if billableMinutes <= 0 {
		addBlocker("no_billable_hours")
	} else {
		preview.BillableHours = float64(billableMinutes) / 60
	}

	unitPrice, currency := s.resolveSellUnitPriceCents(ctx, ts)
	preview.UnitPriceCents = unitPrice
	preview.Currency = currency
	if unitPrice <= 0 {
		addBlocker("zero_unit_price")
	}

	userLabel := userLabelForTimesheet(ctx, s, ts)
	missionLabel := ts.CommercialInfo.Mission
	preview.UserLabel = userLabel
	preview.MissionLabel = missionLabel
	preview.TaxRate = defaultInvoiceTaxRate
	preview.Description = defaultCRAInvoiceDescription(ts.ID, ts.Month, missionLabel, userLabel)

	preview.OK = len(preview.Blockers) == 0
	return preview
}

func defaultCRAInvoiceDescription(timesheetID uuid.UUID, month domain.Month, missionLabel, userLabel string) string {
	mission := missionLabel
	if mission == "" {
		mission = "Prestation"
	}
	if userLabel != "" {
		mission = fmt.Sprintf("%s — %s", mission, userLabel)
	}
	return fmt.Sprintf("CRA/%s/%s %s", timesheetID, month, mission)
}

func (s *Service) tryPublishValidationInvoice(ctx context.Context, ts domain.Timesheet) ports.InvoiceDraftOutcome {
	return s.publishFromPreview(ctx, ts, s.buildInvoiceDraftPreview(ctx, ts), nil)
}

func (s *Service) publishFromPreview(
	ctx context.Context,
	ts domain.Timesheet,
	preview ports.InvoiceDraftPreview,
	item *ports.CreateInvoiceFromTimesheetItem,
) ports.InvoiceDraftOutcome {
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

	cmd := ports.ValidationInvoiceCommand{
		TenantID:        ts.TenantID,
		TimesheetID:     ts.ID,
		TimesheetUserID: ts.UserID,
		Month:           ts.Month,
		BillableHours:   preview.BillableHours,
		MissionLabel:    preview.MissionLabel,
		UserLabel:       preview.UserLabel,
		Currency:        preview.Currency,
		UnitPriceCents:  preview.UnitPriceCents,
		TaxRate:         preview.TaxRate,
		Description:     preview.Description,
	}
	if preview.ClientID != nil {
		cmd.ClientID = *preview.ClientID
	}

	if item != nil {
		if item.ClientID != nil {
			cmd.ClientID = *item.ClientID
		}
		if item.BillableHours != nil {
			cmd.BillableHours = *item.BillableHours
		}
		if item.UnitPriceCents != nil {
			cmd.UnitPriceCents = *item.UnitPriceCents
		}
		if item.TaxRate != nil {
			cmd.TaxRate = *item.TaxRate
		}
		if item.Currency != nil && *item.Currency != "" {
			cmd.Currency = *item.Currency
		}
		if item.Description != nil {
			cmd.Description = *item.Description
		}
		if item.MissionLabel != nil {
			cmd.MissionLabel = *item.MissionLabel
		}
	}

	// Hard blockers always win (billing mode, already invoiced, …) — overrides cannot clear them.
	if hard := firstHardInvoiceBlocker(preview.Blockers); hard != "" {
		status := ports.InvoiceDraftSkipped
		if hard == "invoicing_not_configured" {
			status = ports.InvoiceDraftUnavailable
		}
		return withTS(ports.InvoiceDraftOutcome{Status: status, Reason: hard})
	}

	// Auto path: require a clean preview. Override path: validate essentials after merge.
	if item == nil {
		if !preview.OK {
			reason := "preview_blocked"
			if len(preview.Blockers) > 0 {
				reason = preview.Blockers[0]
			}
			return withTS(ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: reason})
		}
	} else {
		if cmd.ClientID == uuid.Nil {
			return withTS(ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "client_unresolved"})
		}
		if cmd.BillableHours <= 0 {
			return withTS(ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "no_billable_hours"})
		}
		if cmd.UnitPriceCents <= 0 {
			return withTS(ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "zero_unit_price"})
		}
		if cmd.TaxRate <= 0 {
			cmd.TaxRate = defaultInvoiceTaxRate
		}
		if cmd.Currency == "" {
			cmd.Currency = "EUR"
		}
		if cmd.Description == "" {
			cmd.Description = defaultCRAInvoiceDescription(ts.ID, ts.Month, cmd.MissionLabel, cmd.UserLabel)
		}
	}

	invoiceID, err := s.invoices.PublishCRAValidationDraft(ctx, cmd)
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
	case errors.Is(err, invoicingdomain.ErrInvalidInvoiceLine):
		return ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "publish_failed"}
	default:
		return ports.InvoiceDraftOutcome{Status: ports.InvoiceDraftSkipped, Reason: "publish_failed"}
	}
}

func (s *Service) CreateInvoicesFromTimesheets(ctx context.Context, tenant kernel.TenantID, ids []uuid.UUID) ([]ports.InvoiceDraftOutcome, error) {
	items := make([]ports.CreateInvoiceFromTimesheetItem, 0, len(ids))
	for _, id := range ids {
		items = append(items, ports.CreateInvoiceFromTimesheetItem{TimesheetID: id})
	}
	return s.CreateInvoicesFromTimesheetItems(ctx, tenant, items)
}

func (s *Service) CreateInvoicesFromTimesheetItems(ctx context.Context, tenant kernel.TenantID, items []ports.CreateInvoiceFromTimesheetItem) ([]ports.InvoiceDraftOutcome, error) {
	out := make([]ports.InvoiceDraftOutcome, 0, len(items))
	for _, item := range items {
		idCopy := item.TimesheetID
		ts, err := s.repo.GetByID(ctx, tenant, item.TimesheetID)
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
		hasOverride := item.ClientID != nil ||
			item.BillableHours != nil ||
			item.UnitPriceCents != nil ||
			item.TaxRate != nil ||
			item.Currency != nil ||
			item.Description != nil ||
			item.MissionLabel != nil
		preview := s.buildInvoiceDraftPreview(ctx, ts)
		if hasOverride {
			itemCopy := item
			out = append(out, s.publishFromPreview(ctx, ts, preview, &itemCopy))
		} else {
			out = append(out, s.publishFromPreview(ctx, ts, preview, nil))
		}
	}
	return out, nil
}

func (s *Service) PreviewInvoicesFromTimesheets(ctx context.Context, tenant kernel.TenantID, ids []uuid.UUID) ([]ports.InvoiceDraftPreview, error) {
	out := make([]ports.InvoiceDraftPreview, 0, len(ids))
	for _, id := range ids {
		ts, err := s.repo.GetByID(ctx, tenant, id)
		if err != nil {
			out = append(out, ports.InvoiceDraftPreview{
				TimesheetID: id,
				OK:          false,
				Blockers:    []string{"timesheet_not_found"},
			})
			continue
		}
		if ts.Status != domain.StatusDefinitif {
			out = append(out, ports.InvoiceDraftPreview{
				TimesheetID: id,
				OK:          false,
				Blockers:    []string{"not_definitive"},
			})
			continue
		}
		out = append(out, s.buildInvoiceDraftPreview(ctx, ts))
	}
	return out, nil
}

// invoiceableBillableMinutes scopes hours to a single billing bucket (unique invoice per CRA):
// - mission resolved → only that mission's billable lines (other missions/apps excluded)
// - else → only temps_passe application lines for the dominant app, or other billable lines if none
// Application mode non/forfait is excluded; lookup errors are fail-closed.
func (s *Service) invoiceableBillableMinutes(ctx context.Context, ts domain.Timesheet) (int, string, error) {
	missionID := s.resolveMissionID(ctx, ts)
	var missionMinutes int
	var dominantAppMinutes int
	var otherMinutes int
	var sawNon, sawForfait bool
	var lookupFailed bool

	dominantApp := dominantApplicationFromLines(ts)
	dominantAppKey := ""
	if dominantApp != uuid.Nil {
		dominantAppKey = dominantApp.String()
	}

	for _, week := range ts.Weeks {
		for _, line := range week.Lines {
			if !line.Billable || line.Duration.Minutes <= 0 {
				continue
			}
			switch line.Source.Type {
			case "holiday":
				continue
			case "mission":
				if missionID == nil {
					continue
				}
				if line.Source.ID == missionID.String() {
					missionMinutes += line.Duration.Minutes
				}
			case "application":
				if missionID != nil {
					// One invoice per CRA: do not mix application hours with a mission bucket.
					continue
				}
				if dominantAppKey == "" || line.Source.ID != dominantAppKey {
					continue
				}
				if s.apps == nil {
					lookupFailed = true
					continue
				}
				appUUID, err := uuid.Parse(line.Source.ID)
				if err != nil {
					lookupFailed = true
					continue
				}
				info, err := s.apps.GetApplicationBilling(ctx, ts.TenantID, appUUID)
				if err != nil {
					lookupFailed = true
					continue
				}
				switch info.ModeFacturation {
				case orgdomain.ModeFacturationNon:
					sawNon = true
				case orgdomain.ModeFacturationForfait:
					sawForfait = true
				default:
					dominantAppMinutes += line.Duration.Minutes
				}
			default:
				if missionID != nil {
					continue
				}
				otherMinutes += line.Duration.Minutes
			}
		}
	}

	if missionID != nil {
		if missionMinutes > 0 {
			return missionMinutes, "", nil
		}
		if lookupFailed {
			return 0, "billing_mode_unresolved", nil
		}
		return 0, "no_billable_hours", nil
	}

	if dominantAppMinutes > 0 {
		return dominantAppMinutes, "", nil
	}
	if lookupFailed {
		return 0, "billing_mode_unresolved", nil
	}
	if sawNon {
		return 0, "billing_mode_disabled", nil
	}
	if sawForfait {
		return 0, "billing_mode_forfait", nil
	}
	if otherMinutes > 0 {
		return otherMinutes, "", nil
	}
	return 0, "no_billable_hours", nil
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
	if s.repo == nil {
		return nil
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
	if s.repo == nil {
		return nil
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

// resolveSellUnitPriceCents: mission rate > application default TJM > société default TJM.
// Mission hourly rates are used as-is; TJM rates are converted to €/h via day capacity.
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
			currency := rate.Currency
			if currency == "" {
				currency = "EUR"
			}
			if rate.RateUnit == "hourly" {
				return rate.TJMAmount, currency
			}
			return toHourly(rate.TJMAmount, currency)
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
