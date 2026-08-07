package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrMissionNotFound            = errors.New("mission not found")
	ErrInvalidMissionStatus       = errors.New("invalid mission status transition")
	ErrMissionWithoutCollaborator = errors.New("mission without collaborator")
	ErrInvalidRateUnit            = errors.New("invalid rate unit")
	ErrInvalidClientContact       = errors.New("invalid client contact")
)

type MissionStatus string

const (
	MissionStatusActive   MissionStatus = "active"
	MissionStatusArretee  MissionStatus = "arretee"
	MissionStatusTerminee MissionStatus = "terminee"
)

// RateUnit selects how TJMAmount is interpreted for billing.
type RateUnit string

const (
	RateUnitTJM    RateUnit = "tjm"
	RateUnitHourly RateUnit = "hourly"
)

type Mission struct {
	ID               uuid.UUID
	TenantID         kernel.TenantID
	ClientID         uuid.UUID
	Status           MissionStatus
	StartDate        time.Time
	EndDate          *time.Time
	Title            string
	RateUnit         RateUnit
	TJMAmount        int64 // cents — daily (tjm) or hourly (hourly)
	Currency         string
	Technologies     []string
	ClientContact    string // legacy display label
	ClientContactIDs []uuid.UUID
	CreatedAt        time.Time
}

func NewMission(tenant kernel.TenantID, clientID uuid.UUID, startDate time.Time, tjm int64) Mission {
	return Mission{
		ID:        uuid.New(),
		TenantID:  tenant,
		ClientID:  clientID,
		Status:    MissionStatusActive,
		StartDate: startDate,
		RateUnit:  RateUnitTJM,
		TJMAmount: tjm,
		Currency:  "EUR",
		CreatedAt: time.Now().UTC(),
	}
}

// NormalizeRateUnit returns a canonical rate unit. Empty defaults to tjm.
func NormalizeRateUnit(raw string) (RateUnit, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(RateUnitTJM):
		return RateUnitTJM, nil
	case string(RateUnitHourly):
		return RateUnitHourly, nil
	default:
		return "", ErrInvalidRateUnit
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
