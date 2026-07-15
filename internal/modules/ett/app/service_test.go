package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ett/domain"
	"github.com/kore/kore/internal/modules/ett/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ettRepoStub struct {
	records map[uuid.UUID]domain.WorkTimeRecord
	byDate  map[string]domain.WorkTimeRecord
	audits  []domain.AuditEntry
	rule    domain.CountryWorkRule
}

func (r *ettRepoStub) key(userID uuid.UUID, workDate time.Time) string {
	return userID.String() + workDate.Format("2006-01-02")
}

func (r *ettRepoStub) SaveRecord(_ context.Context, rec domain.WorkTimeRecord) error {
	r.records[rec.ID] = rec
	if r.byDate == nil {
		r.byDate = map[string]domain.WorkTimeRecord{}
	}
	r.byDate[r.key(rec.UserID, rec.WorkDate)] = rec
	return nil
}

func (r *ettRepoStub) SaveRecordAndAudit(_ context.Context, rec domain.WorkTimeRecord, entry domain.AuditEntry) error {
	if err := r.SaveRecord(context.Background(), rec); err != nil {
		return err
	}
	return r.AppendAuditEntry(context.Background(), entry)
}

func (r *ettRepoStub) GetRecord(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.WorkTimeRecord, error) {
	rec, ok := r.records[id]
	if !ok {
		return domain.WorkTimeRecord{}, domain.ErrRecordNotFound
	}
	return rec, nil
}

func (r *ettRepoStub) FindRecordByUserDate(_ context.Context, _ kernel.TenantID, userID uuid.UUID, workDate time.Time) (domain.WorkTimeRecord, error) {
	rec, ok := r.byDate[r.key(userID, workDate)]
	if !ok {
		return domain.WorkTimeRecord{}, domain.ErrRecordNotFound
	}
	return rec, nil
}

func (r *ettRepoStub) ListRecords(context.Context, ports.RecordsQuery) ([]domain.WorkTimeRecord, error) {
	return nil, nil
}

func (r *ettRepoStub) AppendAuditEntry(_ context.Context, entry domain.AuditEntry) error {
	r.audits = append(r.audits, entry)
	return nil
}

func (r *ettRepoStub) ListAuditEntries(context.Context, kernel.TenantID, uuid.UUID) ([]domain.AuditEntry, error) {
	return r.audits, nil
}

func (r *ettRepoStub) ListTenantAuditEntries(context.Context, kernel.TenantID) ([]domain.AuditEntry, error) {
	return r.audits, nil
}

func (r *ettRepoStub) GetCountryRule(context.Context, kernel.TenantID, string) (domain.CountryWorkRule, error) {
	return r.rule, nil
}

func TestClockOutAppliesCountryOvertimeRule(t *testing.T) {
	repo := &ettRepoStub{
		records: map[uuid.UUID]domain.WorkTimeRecord{},
		byDate:  map[string]domain.WorkTimeRecord{},
		rule:    domain.CountryWorkRule{MaxDailyHours: 8},
	}
	svc := NewService(repo, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()
	workDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	rec := domain.NewWorkTimeRecord(tenant, userID, workDate)
	in := workDate.Add(7 * time.Hour)
	out := workDate.Add(18 * time.Hour)
	rec.ClockInAt(in)
	repo.records[rec.ID] = rec
	repo.byDate[repo.key(userID, workDate)] = rec

	updated, err := svc.ClockOut(context.Background(), ports.ClockOutCommand{
		TenantID: tenant,
		UserID:   userID,
		At:       out,
	})
	if err != nil {
		t.Fatalf("ClockOut: %v", err)
	}
	if updated.OvertimeHours != 3 {
		t.Fatalf("expected 3 overtime hours, got %v", updated.OvertimeHours)
	}
}

func TestCorrectRecordAuditOnly(t *testing.T) {
	repo := &ettRepoStub{records: map[uuid.UUID]domain.WorkTimeRecord{}}
	svc := NewService(repo, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	recordID := uuid.New()
	actorID := uuid.New()
	in := time.Now().UTC()
	repo.records[recordID] = domain.WorkTimeRecord{
		ID:       recordID,
		TenantID: tenant,
		ClockIn:  &in,
	}

	correctedIn := in.Add(-time.Hour)
	_, err := svc.CorrectRecord(context.Background(), ports.CorrectRecordCommand{
		TenantID: tenant,
		RecordID: recordID,
		ActorID:  actorID,
		ClockIn:  &correctedIn,
	})
	if err != nil {
		t.Fatalf("CorrectRecord: %v", err)
	}
	stored := repo.records[recordID]
	if stored.ClockIn != nil && !stored.ClockIn.Equal(in) {
		t.Fatal("record should remain unchanged on correction")
	}
	if len(repo.audits) != 1 || repo.audits[0].Action != "correct" {
		t.Fatal("expected correction audit entry")
	}
}
