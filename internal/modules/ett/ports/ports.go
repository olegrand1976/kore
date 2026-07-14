package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ett/domain"
	"github.com/kore/kore/pkg/kernel"
)

type ClockInCommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	At       time.Time
}

type ClockOutCommand struct {
	TenantID kernel.TenantID
	UserID   uuid.UUID
	At       time.Time
}

type CorrectRecordCommand struct {
	TenantID kernel.TenantID
	RecordID uuid.UUID
	ActorID  uuid.UUID
	ClockIn  *time.Time
	ClockOut *time.Time
}

type RecordsQuery struct {
	TenantID kernel.TenantID
	UserID   *uuid.UUID
	From     *time.Time
	To       *time.Time
}

type ETTService interface {
	ClockIn(ctx context.Context, cmd ClockInCommand) (domain.WorkTimeRecord, error)
	ClockOut(ctx context.Context, cmd ClockOutCommand) (domain.WorkTimeRecord, error)
	ListRecords(ctx context.Context, q RecordsQuery) ([]domain.WorkTimeRecord, error)
	CorrectRecord(ctx context.Context, cmd CorrectRecordCommand) (domain.WorkTimeRecord, error)
	GetAuditTrail(ctx context.Context, tenant kernel.TenantID, recordID uuid.UUID) ([]domain.AuditEntry, error)
	CompareCRA(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, month string) (ReconciliationReport, error)
}

type ReconciliationReport struct {
	UserID       uuid.UUID `json:"userId"`
	Month        string    `json:"month"`
	CRAHours     float64   `json:"craHours"`
	ETTHours     float64   `json:"ettHours"`
	DeltaHours   float64   `json:"deltaHours"`
	Alert        bool      `json:"alert"`
	AlertMessage string    `json:"alertMessage,omitempty"`
}

type ETTRepository interface {
	SaveRecord(ctx context.Context, rec domain.WorkTimeRecord) error
	GetRecord(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.WorkTimeRecord, error)
	FindRecordByUserDate(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, workDate time.Time) (domain.WorkTimeRecord, error)
	ListRecords(ctx context.Context, q RecordsQuery) ([]domain.WorkTimeRecord, error)
	AppendAuditEntry(ctx context.Context, entry domain.AuditEntry) error
	ListAuditEntries(ctx context.Context, tenant kernel.TenantID, recordID uuid.UUID) ([]domain.AuditEntry, error)
	GetCountryRule(ctx context.Context, tenant kernel.TenantID, countryCode string) (domain.CountryWorkRule, error)
}
