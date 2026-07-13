package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrRecordNotFound = errors.New("work time record not found")
)

type WorkTimeRecord struct {
	ID             uuid.UUID
	TenantID       kernel.TenantID
	UserID         uuid.UUID
	WorkDate       time.Time
	ClockIn        *time.Time
	ClockOut       *time.Time
	EffectiveHours float64
	OvertimeHours  float64
	Status         string
	Origin         string
	CreatedAt      time.Time
}

type AuditEntry struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	RecordID  uuid.UUID
	Action    string
	ActorID   uuid.UUID
	Payload   map[string]any
	CreatedAt time.Time
}

type CountryWorkRule struct {
	ID            uuid.UUID
	TenantID      kernel.TenantID
	CountryCode   string
	MaxDailyHours float64
	MinRestHours  float64
	RetentionDays int
}

func NewWorkTimeRecord(tenant kernel.TenantID, userID uuid.UUID, workDate time.Time) WorkTimeRecord {
	return WorkTimeRecord{
		ID:        uuid.New(),
		TenantID:  tenant,
		UserID:    userID,
		WorkDate:  workDate,
		Status:    "pointed",
		Origin:    "web",
		CreatedAt: time.Now().UTC(),
	}
}

func NewAuditEntry(tenant kernel.TenantID, recordID, actorID uuid.UUID, action string, payload map[string]any) AuditEntry {
	if payload == nil {
		payload = map[string]any{}
	}
	return AuditEntry{
		ID:        uuid.New(),
		TenantID:  tenant,
		RecordID:  recordID,
		Action:    action,
		ActorID:   actorID,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}
}

func (r *WorkTimeRecord) ClockInAt(t time.Time) {
	r.ClockIn = &t
}

func (r *WorkTimeRecord) ClockOutAt(t time.Time) {
	r.ClockOut = &t
	if r.ClockIn != nil {
		hours := t.Sub(*r.ClockIn).Hours()
		r.EffectiveHours = hours
	}
}
