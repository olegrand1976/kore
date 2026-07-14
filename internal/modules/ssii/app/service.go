package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
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
	return ports.MissionDetail{
		ID:            m.ID,
		ClientID:      m.ClientID,
		ClientName:    clientName,
		Status:        string(m.Status),
		StartDate:     m.StartDate,
		EndDate:       m.EndDate,
		TJMAmount:     m.TJMAmount,
		Currency:      m.Currency,
		Technologies:  m.Technologies,
		ClientContact: m.ClientContact,
		CreatedAt:     m.CreatedAt,
		Collaborators: collaborators,
	}, nil
}

func (s *service) Create(ctx context.Context, cmd ports.CreateMissionCommand) (domain.Mission, error) {
	if len(cmd.CollaboratorIDs) == 0 {
		return domain.Mission{}, domain.ErrMissionWithoutCollaborator
	}
	m := domain.NewMission(cmd.TenantID, cmd.ClientID, cmd.StartDate, cmd.TJMAmount)
	m.EndDate = cmd.EndDate
	m.Currency = cmd.Currency
	m.Technologies = cmd.Technologies
	m.ClientContact = cmd.ClientContact
	if m.Currency == "" {
		m.Currency = "EUR"
	}
	if err := s.repo.SaveMission(ctx, m); err != nil {
		return domain.Mission{}, err
	}
	if err := s.repo.SaveMissionCollaborators(ctx, cmd.TenantID, m.ID, cmd.CollaboratorIDs); err != nil {
		return domain.Mission{}, err
	}
	country := cmd.CountryCode
	if country == "" {
		country = "FR"
	}
	if err := s.prefillMissionDays(ctx, m, cmd.CollaboratorIDs, country); err != nil {
		return domain.Mission{}, err
	}
	return m, nil
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
