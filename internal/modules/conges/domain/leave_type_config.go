package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrLeaveTypeNotFound   = errors.New("leave type config not found")
	ErrLeaveTypeInUse      = errors.New("leave type config in use")
	ErrLeaveTypeCodeExists = errors.New("leave type code already exists")
	ErrUnsupportedCountry  = errors.New("unsupported country for leave defaults")
)

type LeaveTypeConfig struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      kernel.TenantID `json:"tenantId"`
	SocieteID     uuid.UUID       `json:"societeId"`
	Code          string          `json:"code"`
	Label         string          `json:"label"`
	TracksBalance bool            `json:"tracksBalance"`
	Active        bool            `json:"active"`
	SortOrder     int             `json:"sortOrder"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type LeaveTypeTemplate struct {
	Code          string
	Label         string
	TracksBalance bool
	SortOrder     int
}
