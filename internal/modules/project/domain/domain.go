package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrEpicNotFound             = errors.New("epic not found")
	ErrSprintNotFound           = errors.New("sprint not found")
	ErrApplicationNotAgile      = errors.New("application is not in agile mode")
	ErrSprintNotPlanned         = errors.New("sprint is not in planned status")
	ErrSprintNotActive          = errors.New("sprint is not active")
	ErrActiveSprintExists       = errors.New("another sprint is already active")
	ErrInvalidStoryPoints       = errors.New("invalid story points")
	ErrMethodologyProfileLocked = errors.New("methodology profile locked after agile artifacts exist")
)

type EpicStatus string

const (
	EpicStatusDraft  EpicStatus = "draft"
	EpicStatusActive EpicStatus = "active"
	EpicStatusDone   EpicStatus = "done"
)

type SprintStatus string

const (
	SprintStatusPlanned SprintStatus = "planned"
	SprintStatusActive  SprintStatus = "active"
	SprintStatusClosed  SprintStatus = "closed"
)

type Epic struct {
	ID              uuid.UUID
	TenantID        kernel.TenantID
	ApplicationID   uuid.UUID
	Title           string
	Description     string
	Status          EpicStatus
	Priority        kernel.RequestPriority
	TargetSprintID  *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Sprint struct {
	ID              uuid.UUID
	TenantID        kernel.TenantID
	ApplicationID   uuid.UUID
	Name            string
	Goal            string
	StartDate       time.Time
	EndDate         time.Time
	Status          SprintStatus
	CapacityPoints  *int16
	CreatedAt       time.Time
	ClosedAt        *time.Time
}

type KanbanColumn struct {
	StateCode string `json:"stateCode"`
	WipLimit  *int   `json:"wipLimit,omitempty"`
	Label     string `json:"label,omitempty"`
}

type KanbanConfig struct {
	ApplicationID uuid.UUID
	TenantID      kernel.TenantID
	Columns       []KanbanColumn
	UpdatedAt     time.Time
}

type BacklogItem struct {
	DemandID     uuid.UUID
	Subject      string
	Status       string
	StoryPoints  *int16
	EpicID       *uuid.UUID
	SprintID     *uuid.UUID
	BacklogRank  *int
	AssigneeID   *uuid.UUID
}

type BurndownPoint struct {
	Date            time.Time `json:"date"`
	RemainingPoints int       `json:"remainingPoints"`
	IdealPoints     int       `json:"idealPoints"`
}

type BurndownSeries struct {
	SprintID        uuid.UUID       `json:"sprintId"`
	PlannedPoints   int             `json:"plannedPoints"`
	Points          []BurndownPoint `json:"points"`
}

type VelocitySprint struct {
	SprintID      uuid.UUID `json:"sprintId"`
	SprintName    string    `json:"sprintName"`
	ClosedPoints  int       `json:"closedPoints"`
}

type VelocityReport struct {
	ApplicationID   uuid.UUID        `json:"applicationId"`
	AverageVelocity float64          `json:"averageVelocity"`
	Sprints         []VelocitySprint `json:"sprints"`
}

func NewEpic(tenant kernel.TenantID, appID uuid.UUID, title, description string, priority kernel.RequestPriority) Epic {
	now := time.Now().UTC()
	return Epic{
		ID:            uuid.New(),
		TenantID:      tenant,
		ApplicationID: appID,
		Title:         title,
		Description:   description,
		Status:        EpicStatusDraft,
		Priority:      priority,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func NewSprint(tenant kernel.TenantID, appID uuid.UUID, name, goal string, start, end time.Time, capacity *int16) Sprint {
	return Sprint{
		ID:             uuid.New(),
		TenantID:       tenant,
		ApplicationID:  appID,
		Name:           name,
		Goal:           goal,
		StartDate:      start,
		EndDate:        end,
		Status:         SprintStatusPlanned,
		CapacityPoints: capacity,
		CreatedAt:      time.Now().UTC(),
	}
}

func (s *Sprint) Start() error {
	if s.Status != SprintStatusPlanned {
		return ErrSprintNotPlanned
	}
	s.Status = SprintStatusActive
	return nil
}

func (s *Sprint) Close(now time.Time) error {
	if s.Status != SprintStatusActive {
		return ErrSprintNotActive
	}
	s.Status = SprintStatusClosed
	s.ClosedAt = &now
	return nil
}

func ValidateStoryPoints(v *int16) error {
	if v == nil {
		return nil
	}
	if *v < 0 || *v > 999 {
		return ErrInvalidStoryPoints
	}
	return nil
}
