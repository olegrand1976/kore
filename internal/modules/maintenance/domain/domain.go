package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrWorkRequestNotFound = errors.New("work request not found")
	ErrInvalidWorkState    = errors.New("invalid work state transition")
)

type WorkState string

const (
	WorkStateCreated    WorkState = "created"
	WorkStateAssigned   WorkState = "assigned"
	WorkStateInProgress WorkState = "in_progress"
	WorkStateCompleted  WorkState = "completed"
)

type WorkRequest struct {
	ID              uuid.UUID
	TenantID        kernel.TenantID
	ApplicationID   uuid.UUID
	Subject         string
	State           WorkState
	AssigneeID      *uuid.UUID
	ConsumptionDays float64
	CreatedAt       time.Time
	CompletedAt     *time.Time
}

func NewWorkRequest(tenant kernel.TenantID, appID uuid.UUID, subject string) WorkRequest {
	return WorkRequest{
		ID:            uuid.New(),
		TenantID:      tenant,
		ApplicationID: appID,
		Subject:       subject,
		State:         WorkStateCreated,
		CreatedAt:     time.Now().UTC(),
	}
}

func (w *WorkRequest) Assign(assigneeID uuid.UUID) {
	w.AssigneeID = &assigneeID
	w.State = WorkStateAssigned
}

func (w *WorkRequest) Progress(consumptionDays float64) error {
	if w.State != WorkStateAssigned && w.State != WorkStateInProgress {
		return ErrInvalidWorkState
	}
	w.State = WorkStateInProgress
	w.ConsumptionDays = consumptionDays
	return nil
}

func (w *WorkRequest) Complete() error {
	if w.State != WorkStateInProgress {
		return ErrInvalidWorkState
	}
	now := time.Now().UTC()
	w.State = WorkStateCompleted
	w.CompletedAt = &now
	return nil
}
