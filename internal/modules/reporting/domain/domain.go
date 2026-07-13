package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrReportNotFound    = errors.New("report definition not found")
	ErrDashboardNotFound = errors.New("dashboard not found")
)

type ReportDefinition struct {
	ID        uuid.UUID
	TenantID  kernel.TenantID
	Code      string
	Name      string
	Config    map[string]any
	Active    bool
	CreatedAt time.Time
}

type Dashboard struct {
	Code       string
	Period     kernel.Period
	Payload    map[string]any
	ComputedAt time.Time
}

type GanttView struct {
	Period kernel.Period
	Items  []GanttItem
}

type GanttItem struct {
	ID        uuid.UUID
	Label     string
	StartDate time.Time
	EndDate   time.Time
	Progress  float64
}

type PlanningView struct {
	Period kernel.Period
	Rows   []PlanningRow
}

type PlanningRow struct {
	UserID   uuid.UUID
	UserName string
	Slots    []PlanningSlot
}

type PlanningSlot struct {
	Date  time.Time
	Label string
	Hours float64
}

type BillingStats struct {
	Period       kernel.Period
	TotalAmount  int64
	InvoiceCount int
	Currency     string
}

type ReportResult struct {
	Definition ReportDefinition
	Rows       []map[string]any
}
