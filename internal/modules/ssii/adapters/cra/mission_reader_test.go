package cra

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/modules/ssii/domain"
	ssiiports "github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type missionRepoStub struct {
	mission domain.Mission
}

func (r *missionRepoStub) SaveMission(context.Context, domain.Mission) error { return nil }
func (r *missionRepoStub) CreateMissionWithRelations(context.Context, domain.Mission, []uuid.UUID, []uuid.UUID) error {
	return nil
}
func (r *missionRepoStub) GetMission(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.Mission, error) {
	if id != r.mission.ID {
		return domain.Mission{}, domain.ErrMissionNotFound
	}
	return r.mission, nil
}
func (r *missionRepoStub) ListMissions(context.Context, kernel.TenantID) ([]domain.Mission, error) {
	return nil, nil
}
func (r *missionRepoStub) ListMissionSummaries(context.Context, kernel.TenantID) ([]ssiiports.MissionSummary, error) {
	return nil, nil
}
func (r *missionRepoStub) ListMissionCollaborators(context.Context, kernel.TenantID, uuid.UUID) ([]ssiiports.MissionCollaborator, error) {
	return nil, nil
}
func (r *missionRepoStub) SaveMissionCollaborators(context.Context, kernel.TenantID, uuid.UUID, []uuid.UUID) error {
	return nil
}
func (r *missionRepoStub) ListMissionApplications(context.Context, kernel.TenantID, uuid.UUID) ([]ssiiports.MissionApplication, error) {
	return nil, nil
}
func (r *missionRepoStub) SaveMissionApplications(context.Context, kernel.TenantID, uuid.UUID, []uuid.UUID) error {
	return nil
}
func (r *missionRepoStub) ValidateApplicationIDs(_ context.Context, _ kernel.TenantID, ids []uuid.UUID, _ uuid.UUID) ([]uuid.UUID, error) {
	return ids, nil
}
func (r *missionRepoStub) GetClientName(context.Context, kernel.TenantID, uuid.UUID) (string, error) {
	return "", nil
}
func (r *missionRepoStub) GetClientPays(context.Context, kernel.TenantID, uuid.UUID) (string, error) {
	return "FR", nil
}
func (r *missionRepoStub) ListClientContacts(context.Context, kernel.TenantID, uuid.UUID) ([]ssiiports.ClientContactSnapshot, error) {
	return nil, nil
}
func (r *missionRepoStub) PurgeClientContactsFromMissions(context.Context, kernel.TenantID, uuid.UUID, []uuid.UUID) error {
	return nil
}

type craActivityStub struct {
	rows []craports.DailyActivityRow
}

func (s *craActivityStub) ListDailyActivityInPeriod(context.Context, kernel.TenantID, kernel.Period) ([]craports.DailyActivityRow, error) {
	return s.rows, nil
}

func TestMissionReaderActiveMissionDays_TJM(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	missionID := uuid.New()
	clientID := uuid.New()
	day := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewPeriod(day, day)
	if err != nil {
		t.Fatalf("period: %v", err)
	}
	reader := NewMissionReader(&missionRepoStub{
		mission: domain.Mission{
			ID:        missionID,
			TenantID:  tenant,
			ClientID:  clientID,
			Title:     "Portail",
			RateUnit:  domain.RateUnitTJM,
			TJMAmount: 50000,
			Currency:  "EUR",
		},
	}, &craActivityStub{rows: []craports.DailyActivityRow{{
		Day:       day,
		Minutes:   480,
		MissionID: missionID.String(),
	}}})

	billing, err := reader.ActiveMissionDays(context.Background(), tenant, missionID, period)
	if err != nil {
		t.Fatalf("ActiveMissionDays: %v", err)
	}
	if billing.Days != 1 || billing.Quantity != 1 {
		t.Fatalf("expected 1 day, got days=%v qty=%v", billing.Days, billing.Quantity)
	}
	if billing.TotalAmount != 50000 {
		t.Fatalf("expected total 50000, got %d", billing.TotalAmount)
	}
	if billing.Title != "Portail" || billing.RateUnit != "tjm" {
		t.Fatalf("unexpected billing meta: %+v", billing)
	}
}

func TestMissionReaderActiveMissionDays_Hourly(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	missionID := uuid.New()
	day := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewPeriod(day, day)
	if err != nil {
		t.Fatalf("period: %v", err)
	}
	reader := NewMissionReader(&missionRepoStub{
		mission: domain.Mission{
			ID:        missionID,
			ClientID:  uuid.New(),
			RateUnit:  domain.RateUnitHourly,
			TJMAmount: 10000, // 100€/h
			Currency:  "EUR",
		},
	}, &craActivityStub{rows: []craports.DailyActivityRow{{
		Day:       day,
		Minutes:   120, // 2h
		MissionID: missionID.String(),
	}}})

	billing, err := reader.ActiveMissionDays(context.Background(), tenant, missionID, period)
	if err != nil {
		t.Fatalf("ActiveMissionDays: %v", err)
	}
	if billing.Hours != 2 || billing.Quantity != 2 {
		t.Fatalf("expected 2h, got hours=%v qty=%v", billing.Hours, billing.Quantity)
	}
	if billing.TotalAmount != 20000 {
		t.Fatalf("expected total 20000, got %d", billing.TotalAmount)
	}
	if billing.RateUnit != "hourly" {
		t.Fatalf("rate unit = %q", billing.RateUnit)
	}
}
