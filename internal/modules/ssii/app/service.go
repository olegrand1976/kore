package app

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type service struct {
	repo     ports.SSIIRepository
	feeder   ports.CRAFeeder
	cleaner  ports.CRAFutureCleaner
	calendar ports.WorkCalendarGateway
}

func NewService(repo ports.SSIIRepository, feeder ports.CRAFeeder, cleaner ports.CRAFutureCleaner, calendar ports.WorkCalendarGateway) ports.SSIIService {
	return &service{repo: repo, feeder: feeder, cleaner: cleaner, calendar: calendar}
}

func (s *service) List(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error) {
	return s.repo.ListMissions(ctx, tenant)
}

func (s *service) ListSummaries(ctx context.Context, tenant kernel.TenantID) ([]ports.MissionSummary, error) {
	return s.repo.ListMissionSummaries(ctx, tenant)
}

func (s *service) Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error) {
	return s.repo.GetMission(ctx, tenant, id)
}

func (s *service) GetDetail(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (ports.MissionDetail, error) {
	m, err := s.repo.GetMission(ctx, tenant, id)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	clientName, err := s.repo.GetClientName(ctx, tenant, m.ClientID)
	if err != nil {
		clientName = ""
	}
	collaborators, err := s.repo.ListMissionCollaborators(ctx, tenant, m.ID)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	if collaborators == nil {
		collaborators = []ports.MissionCollaborator{}
	}
	applications, err := s.repo.ListMissionApplications(ctx, tenant, m.ID)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	if applications == nil {
		applications = []ports.MissionApplication{}
	}
	clientContacts, contactIDs, contactLabel, err := s.resolveClientContacts(ctx, tenant, m.ClientID, m.ClientContactIDs)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	// Legacy free-text only when no structured contact IDs were stored.
	if contactLabel == "" && len(m.ClientContactIDs) == 0 {
		contactLabel = m.ClientContact
	}
	rateUnit := string(m.RateUnit)
	if rateUnit == "" {
		rateUnit = string(domain.RateUnitTJM)
	}
	return ports.MissionDetail{
		ID:                 m.ID,
		ClientID:           m.ClientID,
		ClientName:         clientName,
		Status:             string(m.Status),
		StartDate:          m.StartDate,
		EndDate:            m.EndDate,
		Title:              m.Title,
		RateUnit:           rateUnit,
		TJMAmount:          m.TJMAmount,
		Currency:           m.Currency,
		Technologies:       m.Technologies,
		ClientContact:      contactLabel,
		ClientContactIDs:   contactIDs,
		ClientContacts:     clientContacts,
		PlannedWeekMinutes: m.PlannedWeekMinutes,
		CreatedAt:          m.CreatedAt,
		Collaborators:      collaborators,
		Applications:       applications,
	}, nil
}

func (s *service) Create(ctx context.Context, cmd ports.CreateMissionCommand) (domain.Mission, error) {
	if len(cmd.CollaboratorIDs) == 0 {
		return domain.Mission{}, domain.ErrMissionWithoutCollaborator
	}
	rateUnit, err := domain.NormalizeRateUnit(cmd.RateUnit)
	if err != nil {
		return domain.Mission{}, err
	}
	contactIDs, contactLabel, err := s.validateAndLabelContacts(ctx, cmd.TenantID, cmd.ClientID, cmd.ClientContactIDs)
	if err != nil {
		return domain.Mission{}, err
	}
	if contactLabel == "" {
		contactLabel = strings.TrimSpace(cmd.ClientContact)
	}
	appIDs, err := s.validateApplicationIDs(ctx, cmd.TenantID, cmd.ApplicationIDs, uuid.Nil)
	if err != nil {
		return domain.Mission{}, err
	}
	planned, err := domain.NormalizePlannedWeekMinutes(cmd.PlannedWeekMinutes)
	if err != nil {
		return domain.Mission{}, err
	}
	m := domain.NewMission(cmd.TenantID, cmd.ClientID, cmd.StartDate, cmd.TJMAmount)
	m.EndDate = cmd.EndDate
	m.Title = strings.TrimSpace(cmd.Title)
	m.RateUnit = rateUnit
	m.Currency = cmd.Currency
	m.Technologies = cmd.Technologies
	m.ClientContact = contactLabel
	m.ClientContactIDs = contactIDs
	m.PlannedWeekMinutes = planned
	if m.Currency == "" {
		m.Currency = "EUR"
	}
	if err := s.repo.CreateMissionWithRelations(ctx, m, cmd.CollaboratorIDs, appIDs); err != nil {
		return domain.Mission{}, err
	}
	country := s.resolveClientCountry(ctx, cmd.TenantID, cmd.ClientID)
	if err := s.prefillMissionDays(ctx, m, cmd.CollaboratorIDs, country); err != nil {
		return domain.Mission{}, err
	}
	return m, nil
}

func (s *service) Update(ctx context.Context, cmd ports.UpdateMissionCommand) (ports.MissionDetail, error) {
	m, err := s.repo.GetMission(ctx, cmd.TenantID, cmd.MissionID)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	rateUnit, err := domain.NormalizeRateUnit(cmd.RateUnit)
	if err != nil {
		return ports.MissionDetail{}, err
	}

	beforeTitle := m.Title
	beforeRateUnit := m.RateUnit
	beforeTJM := m.TJMAmount
	beforeContact := m.ClientContact
	beforeContactIDs := append([]uuid.UUID(nil), m.ClientContactIDs...)
	beforePlanned := cloneIntPtr(m.PlannedWeekMinutes)

	m.Title = strings.TrimSpace(cmd.Title)
	m.RateUnit = rateUnit
	m.TJMAmount = cmd.TJMAmount
	if cmd.ClientContactIDs != nil {
		contactIDs, contactLabel, err := s.validateAndLabelContacts(ctx, cmd.TenantID, m.ClientID, *cmd.ClientContactIDs)
		if err != nil {
			return ports.MissionDetail{}, err
		}
		m.ClientContactIDs = contactIDs
		if contactLabel != "" {
			m.ClientContact = contactLabel
		} else {
			m.ClientContact = strings.TrimSpace(cmd.ClientContact)
		}
	} else if strings.TrimSpace(cmd.ClientContact) != "" {
		m.ClientContact = strings.TrimSpace(cmd.ClientContact)
	}
	if cmd.PlannedWeekMinutesSet {
		planned, nErr := domain.NormalizePlannedWeekMinutes(cmd.PlannedWeekMinutes)
		if nErr != nil {
			return ports.MissionDetail{}, nErr
		}
		m.PlannedWeekMinutes = planned
	}
	if err := s.repo.SaveMission(ctx, m); err != nil {
		return ports.MissionDetail{}, err
	}

	rateChanged := beforeTitle != m.Title ||
		beforeRateUnit != m.RateUnit ||
		beforeTJM != m.TJMAmount ||
		beforeContact != m.ClientContact ||
		!slices.Equal(beforeContactIDs, m.ClientContactIDs)
	plannedChanged := cmd.PlannedWeekMinutesSet && !intPtrEqual(beforePlanned, m.PlannedWeekMinutes)

	if cmd.ActorUserID != uuid.Nil {
		if rateChanged {
			beforePayload := map[string]any{
				"title":            beforeTitle,
				"rateUnit":         string(beforeRateUnit),
				"tjmAmount":        beforeTJM,
				"clientContact":    beforeContact,
				"clientContactIds": uuidSliceStrings(beforeContactIDs),
			}
			afterPayload := map[string]any{
				"title":            m.Title,
				"rateUnit":         string(m.RateUnit),
				"tjmAmount":        m.TJMAmount,
				"clientContact":    m.ClientContact,
				"clientContactIds": uuidSliceStrings(m.ClientContactIDs),
			}
			if plannedChanged {
				beforePayload["plannedWeekMinutes"] = intPtrValue(beforePlanned)
				afterPayload["plannedWeekMinutes"] = intPtrValue(m.PlannedWeekMinutes)
			}
			_ = s.repo.InsertBillingEvent(ctx, cmd.TenantID, ports.MissionBillingEvent{
				ID:          uuid.New(),
				MissionID:   m.ID,
				ActorUserID: cmd.ActorUserID,
				EventType:   domain.BillingEventRateUpdated,
				Payload:     map[string]any{"before": beforePayload, "after": afterPayload},
				CreatedAt:   time.Now().UTC(),
			})
		}
		if plannedChanged {
			_ = s.repo.InsertBillingEvent(ctx, cmd.TenantID, ports.MissionBillingEvent{
				ID:          uuid.New(),
				MissionID:   m.ID,
				ActorUserID: cmd.ActorUserID,
				EventType:   domain.BillingEventPlannedHoursUpdated,
				Payload: map[string]any{
					"before": map[string]any{"plannedWeekMinutes": intPtrValue(beforePlanned)},
					"after":  map[string]any{"plannedWeekMinutes": intPtrValue(m.PlannedWeekMinutes)},
				},
				CreatedAt: time.Now().UTC(),
			})
		}
	}

	return s.GetDetail(ctx, cmd.TenantID, m.ID)
}

func (s *service) ListBillingEvents(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID) ([]ports.MissionBillingEvent, error) {
	if _, err := s.repo.GetMission(ctx, tenant, missionID); err != nil {
		return nil, err
	}
	events, err := s.repo.ListBillingEvents(ctx, tenant, missionID)
	if err != nil {
		return nil, err
	}
	if events == nil {
		return []ports.MissionBillingEvent{}, nil
	}
	return events, nil
}

func (s *service) AddBillingNote(ctx context.Context, cmd ports.AddBillingEventCommand) (ports.MissionBillingEvent, error) {
	if _, err := s.repo.GetMission(ctx, cmd.TenantID, cmd.MissionID); err != nil {
		return ports.MissionBillingEvent{}, err
	}
	if cmd.ActorUserID == uuid.Nil {
		return ports.MissionBillingEvent{}, domain.ErrInvalidBillingEvent
	}
	msg := strings.TrimSpace(cmd.Message)
	if msg == "" {
		return ports.MissionBillingEvent{}, domain.ErrInvalidBillingEvent
	}
	eventType := strings.TrimSpace(cmd.EventType)
	if eventType == "" {
		eventType = domain.BillingEventNote
	}
	if eventType != domain.BillingEventNote {
		return ports.MissionBillingEvent{}, domain.ErrInvalidBillingEvent
	}
	payload := cmd.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	event := ports.MissionBillingEvent{
		ID:          uuid.New(),
		MissionID:   cmd.MissionID,
		ActorUserID: cmd.ActorUserID,
		EventType:   eventType,
		Message:     msg,
		Payload:     payload,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.InsertBillingEvent(ctx, cmd.TenantID, event); err != nil {
		return ports.MissionBillingEvent{}, err
	}
	return event, nil
}

func cloneIntPtr(v *int) *int {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func intPtrValue(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func uuidSliceStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}

func (s *service) Stop(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error) {
	m, err := s.repo.GetMission(ctx, tenant, id)
	if err != nil {
		return domain.Mission{}, err
	}
	if err := m.Stop(); err != nil {
		return domain.Mission{}, err
	}
	if err := s.repo.SaveMission(ctx, m); err != nil {
		return domain.Mission{}, err
	}
	if err := s.purgeFutureMissionLines(ctx, m.ID); err != nil {
		return domain.Mission{}, err
	}
	return m, nil
}

func (s *service) UpdateEndDate(ctx context.Context, cmd ports.UpdateEndDateCommand) (domain.Mission, error) {
	m, err := s.repo.GetMission(ctx, cmd.TenantID, cmd.MissionID)
	if err != nil {
		return domain.Mission{}, err
	}
	m.SetEndDate(cmd.EndDate)
	if err := s.repo.SaveMission(ctx, m); err != nil {
		return domain.Mission{}, err
	}
	collaborators, err := s.repo.ListMissionCollaborators(ctx, cmd.TenantID, m.ID)
	if err != nil {
		return domain.Mission{}, err
	}
	ids := make([]uuid.UUID, len(collaborators))
	for i, c := range collaborators {
		ids[i] = c.UserID
	}
	if err := s.purgeFutureMissionLines(ctx, m.ID); err != nil {
		return domain.Mission{}, err
	}
	if err := s.prefillMissionDays(ctx, m, ids, "FR"); err != nil {
		return domain.Mission{}, err
	}
	return m, nil
}

func (s *service) UpdateCollaborators(ctx context.Context, cmd ports.UpdateCollaboratorsCommand) (ports.MissionDetail, error) {
	if len(cmd.CollaboratorIDs) == 0 {
		return ports.MissionDetail{}, domain.ErrMissionWithoutCollaborator
	}
	m, err := s.repo.GetMission(ctx, cmd.TenantID, cmd.MissionID)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	if err := s.repo.SaveMissionCollaborators(ctx, cmd.TenantID, m.ID, cmd.CollaboratorIDs); err != nil {
		return ports.MissionDetail{}, err
	}
	if err := s.purgeFutureMissionLines(ctx, m.ID); err != nil {
		return ports.MissionDetail{}, err
	}
	if err := s.prefillMissionDays(ctx, m, cmd.CollaboratorIDs, "FR"); err != nil {
		return ports.MissionDetail{}, err
	}
	return s.GetDetail(ctx, cmd.TenantID, m.ID)
}

func (s *service) UpdateApplications(ctx context.Context, cmd ports.UpdateApplicationsCommand) (ports.MissionDetail, error) {
	m, err := s.repo.GetMission(ctx, cmd.TenantID, cmd.MissionID)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	appIDs, err := s.validateApplicationIDs(ctx, cmd.TenantID, cmd.ApplicationIDs, m.ID)
	if err != nil {
		return ports.MissionDetail{}, err
	}
	if err := s.repo.SaveMissionApplications(ctx, cmd.TenantID, m.ID, appIDs); err != nil {
		return ports.MissionDetail{}, err
	}
	return s.GetDetail(ctx, cmd.TenantID, m.ID)
}

func (s *service) validateApplicationIDs(
	ctx context.Context,
	tenant kernel.TenantID,
	ids []uuid.UUID,
	missionID uuid.UUID,
) ([]uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, nil
	}
	requested := dedupeUUIDs(ids)
	out, err := s.repo.ValidateApplicationIDs(ctx, tenant, requested, missionID)
	if err != nil {
		return nil, err
	}
	if len(out) != len(requested) {
		return nil, domain.ErrInvalidApplication
	}
	return out, nil
}

func dedupeUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return []uuid.UUID{}
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *service) validateAndLabelContacts(
	ctx context.Context,
	tenant kernel.TenantID,
	clientID uuid.UUID,
	ids []uuid.UUID,
) ([]uuid.UUID, string, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, "", nil
	}
	all, err := s.repo.ListClientContacts(ctx, tenant, clientID)
	if err != nil {
		return nil, "", err
	}
	byID := make(map[uuid.UUID]ports.ClientContactSnapshot, len(all))
	for _, c := range all {
		byID[c.ID] = c
	}
	outIDs := make([]uuid.UUID, 0, len(ids))
	names := make([]string, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		c, ok := byID[id]
		if !ok {
			return nil, "", domain.ErrInvalidClientContact
		}
		seen[id] = struct{}{}
		outIDs = append(outIDs, id)
		names = append(names, contactDisplayName(c))
	}
	return outIDs, strings.Join(names, ", "), nil
}

func (s *service) resolveClientContacts(
	ctx context.Context,
	tenant kernel.TenantID,
	clientID uuid.UUID,
	ids []uuid.UUID,
) ([]ports.MissionClientContact, []uuid.UUID, string, error) {
	if len(ids) == 0 {
		return []ports.MissionClientContact{}, []uuid.UUID{}, "", nil
	}
	all, err := s.repo.ListClientContacts(ctx, tenant, clientID)
	if err != nil {
		return nil, nil, "", err
	}
	byID := make(map[uuid.UUID]ports.ClientContactSnapshot, len(all))
	for _, c := range all {
		byID[c.ID] = c
	}
	out := make([]ports.MissionClientContact, 0, len(ids))
	outIDs := make([]uuid.UUID, 0, len(ids))
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		c, ok := byID[id]
		if !ok {
			continue
		}
		out = append(out, ports.MissionClientContact(c))
		outIDs = append(outIDs, id)
		names = append(names, contactDisplayName(c))
	}
	return out, outIDs, strings.Join(names, ", "), nil
}

func contactDisplayName(c ports.ClientContactSnapshot) string {
	name := strings.TrimSpace(strings.TrimSpace(c.Prenom) + " " + strings.TrimSpace(c.Nom))
	if name != "" {
		return name
	}
	if email := strings.TrimSpace(c.Email); email != "" {
		return email
	}
	return c.ID.String()
}

// resolveClientCountry returns the client's billing country for holiday prefill.
// Falls back to FR when lookup fails or the value is outside the org whitelist.
func (s *service) resolveClientCountry(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID) string {
	pays, err := s.repo.GetClientPays(ctx, tenant, clientID)
	if err != nil {
		slog.Default().WarnContext(ctx, "ssii: resolve client pays failed, defaulting to FR",
			"clientId", clientID, "err", err)
		return "FR"
	}
	normalized, ok := orgdomain.NormalizeSocietePays(pays)
	if !ok {
		return "FR"
	}
	return normalized
}

func (s *service) prefillMissionDays(ctx context.Context, m domain.Mission, collaborators []uuid.UUID, countryCode string) error {
	if s.feeder == nil || s.calendar == nil || len(collaborators) == 0 {
		return nil
	}
	end := time.Now().UTC().AddDate(0, 3, 0)
	if m.EndDate != nil && m.EndDate.Before(end) {
		end = *m.EndDate
	}
	start := m.StartDate.UTC()
	if start.Before(time.Now().UTC()) {
		start = truncateDay(time.Now().UTC())
	}

	var lines []ports.ProposedMissionLine
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
			continue
		}
		month := cradomain.Month(day.Format("2006-01"))
		for _, userID := range collaborators {
			blocked, err := s.calendar.IsHolidayOrLeave(ctx, m.TenantID, userID, day, countryCode)
			if err != nil || blocked {
				continue
			}
			lines = append(lines, ports.ProposedMissionLine{
				TenantID:  m.TenantID,
				UserID:    userID,
				MissionID: m.ID,
				Month:     month,
				Day:       day,
				Duration:  kernel.Duration{Minutes: 480},
				Comment:   "Mission",
			})
		}
	}
	return s.feeder.ProposeLines(ctx, lines)
}

func (s *service) purgeFutureMissionLines(ctx context.Context, missionID uuid.UUID) error {
	if s.cleaner == nil {
		return nil
	}
	return s.cleaner.RemoveFutureLines(ctx, missionID.String(), truncateDay(time.Now().UTC()))
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

var _ ports.SSIIService = (*service)(nil)
