package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrMissionNotFound      = errors.New("mission not found")
	ErrInvalidMissionStatus = errors.New("invalid mission status transition")
)

type MissionStatus string

const (
	MissionStatusActive   MissionStatus = "active"
	MissionStatusArretee  MissionStatus = "arretee"
	MissionStatusTerminee MissionStatus = "terminee"
)

type Mission struct {
	ID            uuid.UUID
	TenantID      kernel.TenantID
	ClientID      uuid.UUID
	Status        MissionStatus
	StartDate     time.Time
	EndDate       *time.Time
	TJMAmount     int64
	Currency      string
	Technologies  []string
	ClientContact string
	CreatedAt     time.Time
}

func NewMission(tenant kernel.TenantID, clientID uuid.UUID, startDate time.Time, tjm int64) Mission {
	return Mission{
		ID:        uuid.New(),
		TenantID:  tenant,
		ClientID:  clientID,
		Status:    MissionStatusActive,
		StartDate: startDate,
		TJMAmount: tjm,
		Currency:  "EUR",
		CreatedAt: time.Now().UTC(),
	}
}

func (m *Mission) Stop() error {
	if m.Status != MissionStatusActive {
		return ErrInvalidMissionStatus
	}
	m.Status = MissionStatusArretee
	return nil
}

func (m *Mission) SetEndDate(endDate time.Time) {
	m.EndDate = &endDate
	if m.Status == MissionStatusActive {
		m.Status = MissionStatusTerminee
	}
}
