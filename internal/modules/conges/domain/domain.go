package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrLeaveAlreadyDecided = errors.New("leave request already decided")
	ErrLeaveNotEditable    = errors.New("leave request not editable after decision")
	ErrLeavePastDate       = errors.New("approval cannot affect past days")
	ErrInvalidDateRange    = errors.New("invalid date range")
)

type LeaveType string

const (
	LeaveTypeCongesPayes LeaveType = "conges_payes"
	LeaveTypeRTT         LeaveType = "rtt"
	LeaveTypeMaladie     LeaveType = "maladie"
)

type LeaveStatus string

const (
	LeaveStatusPending  LeaveStatus = "en_attente"
	LeaveStatusApproved LeaveStatus = "valide"
	LeaveStatusRejected LeaveStatus = "refuse"
)

type LeaveRequest struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	UserID    uuid.UUID
	Type      LeaveType
	Period    kernel.DateRange
	Motif     string
	Status    LeaveStatus
	DecidedBy *uuid.UUID
	DecidedAt *time.Time
}

type LeaveBalance struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	UserID    uuid.UUID
	Type      LeaveType
	Acquired  float64
	Taken     float64
	Remaining float64
}

func NewLeaveRequest(tenant kernel.TenantID, userID uuid.UUID, leaveType LeaveType, period kernel.DateRange, motif string) LeaveRequest {
	return LeaveRequest{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Type:     leaveType,
		Period:   period,
		Motif:    motif,
		Status:   LeaveStatusPending,
	}
}

func (r *LeaveRequest) Approve(now time.Time, decidedBy uuid.UUID) error {
	if r.Status != LeaveStatusPending {
		return ErrLeaveAlreadyDecided
	}
	r.Status = LeaveStatusApproved
	r.DecidedBy = &decidedBy
	t := now.UTC()
	r.DecidedAt = &t
	return nil
}

func (r *LeaveRequest) Reject(now time.Time, decidedBy uuid.UUID) error {
	if r.Status != LeaveStatusPending {
		return ErrLeaveAlreadyDecided
	}
	r.Status = LeaveStatusRejected
	r.DecidedBy = &decidedBy
	t := now.UTC()
	r.DecidedAt = &t
	return nil
}

// FutureDays returns calendar days strictly after today within the leave period.
func FutureDays(period kernel.DateRange, today time.Time) []time.Time {
	today = today.UTC().Truncate(24 * time.Hour)
	from := period.From.UTC().Truncate(24 * time.Hour)
	to := period.To.UTC().Truncate(24 * time.Hour)
	var days []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if d.After(today) {
			days = append(days, d)
		}
	}
	return days
}
