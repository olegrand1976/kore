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

func NewService(repo ports.ETTRepository, cra craports.CRAReader, users userProfileReader) ports.ETTService {
	return &service{
		repo:  repo,
		recon: NewReconciliationService(repo, cra),
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
	if err := s.repo.SaveRecord(ctx, rec); err != nil {
		return domain.WorkTimeRecord{}, err
	}
	entry := domain.NewAuditEntry(cmd.TenantID, rec.ID, cmd.UserID, "clock_in", map[string]any{"at": cmd.At})
	return rec, s.repo.AppendAuditEntry(ctx, entry)
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
	if err := s.repo.SaveRecord(ctx, rec); err != nil {
		return domain.WorkTimeRecord{}, err
	}
	entry := domain.NewAuditEntry(cmd.TenantID, rec.ID, cmd.UserID, "clock_out", map[string]any{"at": cmd.At})
	return rec, s.repo.AppendAuditEntry(ctx, entry)
}

func (s *service) ListRecords(ctx context.Context, q ports.RecordsQuery) ([]domain.WorkTimeRecord, error) {
	return s.repo.ListRecords(ctx, q)
}

func (s *service) CorrectRecord(ctx context.Context, cmd ports.CorrectRecordCommand) (domain.WorkTimeRecord, error) {
	rec, err := s.repo.GetRecord(ctx, cmd.TenantID, cmd.RecordID)
	if err != nil {
		return domain.WorkTimeRecord{}, err
	}
	payload := map[string]any{}
	if cmd.ClockIn != nil {
		rec.ClockIn = cmd.ClockIn
		payload["clockIn"] = cmd.ClockIn
	}
	if cmd.ClockOut != nil {
		rec.ClockOut = cmd.ClockOut
		payload["clockOut"] = cmd.ClockOut
		if rec.ClockIn != nil {
			rec.EffectiveHours = cmd.ClockOut.Sub(*rec.ClockIn).Hours()
		}
	}
	if err := s.repo.SaveRecord(ctx, rec); err != nil {
		return domain.WorkTimeRecord{}, err
	}
	entry := domain.NewAuditEntry(cmd.TenantID, rec.ID, cmd.ActorID, "correct", payload)
	return rec, s.repo.AppendAuditEntry(ctx, entry)
}

func (s *service) GetAuditTrail(ctx context.Context, tenant kernel.TenantID, recordID uuid.UUID) ([]domain.AuditEntry, error) {
	return s.repo.ListAuditEntries(ctx, tenant, recordID)
}

func (s *service) CompareCRA(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month string) (ports.ReconciliationReport, error) {
	return s.recon.CompareMonth(ctx, tenant, userID, month)
}

var _ ports.ETTService = (*service)(nil)
