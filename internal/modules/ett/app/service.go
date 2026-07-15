package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/modules/ett/domain"
	"github.com/kore/kore/internal/modules/ett/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type userProfileReader interface {
	FindUserDetailByID(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (orgports.UserDetail, error)
}

type service struct {
	repo  ports.ETTRepository
	recon *ReconciliationService
	users userProfileReader
}

func NewService(repo ports.ETTRepository, craReader craports.CRAReader, craSvc craports.CRAService, users userProfileReader) ports.ETTService {
	return &service{
		repo:  repo,
		recon: NewReconciliationService(repo, craReader, craSvc, users),
		users: users,
	}
}

func (s *service) ensureSalarieETT(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) error {
	if s.users == nil {
		return nil
	}
	detail, err := s.users.FindUserDetailByID(ctx, tenant, userID)
	if err != nil {
		return err
	}
	if !detail.SalarieETT {
		return domain.ErrNotSalarieETT
	}
	return nil
}

func (s *service) ClockIn(ctx context.Context, cmd ports.ClockInCommand) (domain.WorkTimeRecord, error) {
	if err := s.ensureSalarieETT(ctx, cmd.TenantID, cmd.UserID); err != nil {
		return domain.WorkTimeRecord{}, err
	}
	workDate := cmd.At.UTC().Truncate(24 * time.Hour)
	rec, err := s.repo.FindRecordByUserDate(ctx, cmd.TenantID, cmd.UserID, workDate)
	if err != nil {
		rec = domain.NewWorkTimeRecord(cmd.TenantID, cmd.UserID, workDate)
	}
	rec.ClockInAt(cmd.At.UTC())
	entry := domain.NewAuditEntry(cmd.TenantID, rec.ID, cmd.UserID, "clock_in", map[string]any{
		"at": cmd.At.UTC().Truncate(time.Microsecond),
	})
	return rec, s.repo.SaveRecordAndAudit(ctx, rec, entry)
}

func (s *service) ClockOut(ctx context.Context, cmd ports.ClockOutCommand) (domain.WorkTimeRecord, error) {
	if err := s.ensureSalarieETT(ctx, cmd.TenantID, cmd.UserID); err != nil {
		return domain.WorkTimeRecord{}, err
	}
	workDate := cmd.At.UTC().Truncate(24 * time.Hour)
	rec, err := s.repo.FindRecordByUserDate(ctx, cmd.TenantID, cmd.UserID, workDate)
	if err != nil {
		return domain.WorkTimeRecord{}, err
	}
	rec.ClockOutAt(cmd.At.UTC())
	s.applyOvertime(ctx, &rec, cmd.TenantID, "BE")
	entry := domain.NewAuditEntry(cmd.TenantID, rec.ID, cmd.UserID, "clock_out", map[string]any{
		"at": cmd.At.UTC().Truncate(time.Microsecond),
	})
	return rec, s.repo.SaveRecordAndAudit(ctx, rec, entry)
}

func (s *service) ListRecords(ctx context.Context, q ports.RecordsQuery) ([]domain.WorkTimeRecord, error) {
	return s.repo.ListRecords(ctx, q)
}

func (s *service) CorrectRecord(ctx context.Context, cmd ports.CorrectRecordCommand) (domain.WorkTimeRecord, error) {
	rec, err := s.repo.GetRecord(ctx, cmd.TenantID, cmd.RecordID)
	if err != nil {
		return domain.WorkTimeRecord{}, err
	}
	payload := map[string]any{
		"previousClockIn":  rec.ClockIn,
		"previousClockOut": rec.ClockOut,
	}
	if cmd.ClockIn != nil {
		payload["clockIn"] = cmd.ClockIn
	}
	if cmd.ClockOut != nil {
		payload["clockOut"] = cmd.ClockOut
	}
	entry := domain.NewAuditEntry(cmd.TenantID, rec.ID, cmd.ActorID, "correct", payload)
	return rec, s.repo.AppendAuditEntry(ctx, entry)
}

func (s *service) applyOvertime(ctx context.Context, rec *domain.WorkTimeRecord, tenant kernel.TenantID, countryCode string) {
	if countryCode == "" {
		countryCode = "BE"
	}
	maxDaily := 8.0
	if rule, err := s.repo.GetCountryRule(ctx, tenant, countryCode); err == nil && rule.MaxDailyHours > 0 {
		maxDaily = rule.MaxDailyHours
	}
	if rec.EffectiveHours > maxDaily {
		rec.OvertimeHours = rec.EffectiveHours - maxDaily
	} else {
		rec.OvertimeHours = 0
	}
}

func (s *service) GetAuditTrail(ctx context.Context, tenant kernel.TenantID, recordID uuid.UUID) ([]domain.AuditEntry, error) {
	return s.repo.ListAuditEntries(ctx, tenant, recordID)
}

func (s *service) VerifyAuditIntegrity(ctx context.Context, tenant kernel.TenantID) (ports.AuditIntegrityReport, error) {
	entries, err := s.repo.ListTenantAuditEntries(ctx, tenant)
	if err != nil {
		return ports.AuditIntegrityReport{}, err
	}
	brokenAt, valid := domain.VerifyChain(entries)
	report := ports.AuditIntegrityReport{Entries: len(entries), Valid: valid}
	if valid {
		report.Code = "INTEGRITY_OK"
	} else {
		seq := brokenAt
		report.BrokenAtSeq = &seq
		report.Code = "INTEGRITY_BROKEN"
	}
	return report, nil
}

func (s *service) CompareCRA(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month string) (ports.ReconciliationReport, error) {
	return s.recon.CompareMonth(ctx, tenant, userID, month)
}

func (s *service) CompareCRATeam(ctx context.Context, tenant kernel.TenantID, month string) ([]ports.ReconciliationReport, error) {
	return s.recon.CompareTenant(ctx, tenant, month)
}

var _ ports.ETTService = (*service)(nil)
