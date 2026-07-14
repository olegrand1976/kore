package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrCRAAlreadyValidated    = errors.New("cra already validated")
	ErrCommercialInfoRequired = errors.New("commercial info required")
	ErrDayCapacityExceeded    = errors.New("day capacity exceeded")
	ErrCRAConflictAbsence     = errors.New("cra conflict absence")
	ErrTimesheetNotFound      = errors.New("timesheet not found")
	ErrWeekNotFound           = errors.New("week not found")
	ErrWeekIncomplete         = errors.New("week has days without hours")
)

const (
	DefaultDayCapacityMinutes = 480
	DefaultWeekSubmitPolicy   = "warn"
)

type TimesheetStatus string

const (
	StatusBrouillon     TimesheetStatus = "Brouillon"
	StatusValideSemaine TimesheetStatus = "ValidéSemaine"
	StatusDefinitif     TimesheetStatus = "Définitif"
)

type LineOrigin string

const (
	OriginManual  LineOrigin = "manual"
	OriginPrefill LineOrigin = "prefill"
)

type Month string

func ParseMonth(raw string) (Month, error) {
	if _, err := time.Parse("2006-01", raw); err != nil {
		return "", fmt.Errorf("invalid month format, expected YYYY-MM")
	}
	return Month(raw), nil
}

type WeekNumber int

type CommercialInfo struct {
	ClientID          *uuid.UUID `json:"clientId,omitempty"`
	MissionID         *uuid.UUID `json:"missionId,omitempty"`
	Client            string     `json:"client"`
	Mission           string     `json:"mission"`
	Description       string     `json:"description,omitempty"`
	Technologies      []string   `json:"technologies,omitempty"`
	Lieu              string     `json:"lieu,omitempty"`
	ResponsableClient string     `json:"responsableClient,omitempty"`
}

func (c CommercialInfo) Complete() bool {
	return c.Client != "" && c.Mission != ""
}

type SourceRef struct {
	Type string
	ID   string
}

type TimeLine struct {
	ID          uuid.UUID
	TenantID    kernel.TenantID
	WeekEntryID uuid.UUID
	Source      SourceRef
	Day         time.Time
	Duration    kernel.Duration
	Comment     string
	Origin      LineOrigin
	Billable    bool
}

type WeekEntry struct {
	ID          uuid.UUID
	TenantID    kernel.TenantID
	TimesheetID uuid.UUID
	WeekNumber  WeekNumber
	SubmittedAt *time.Time
	Lines       []TimeLine
}

type Timesheet struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       kernel.TenantID `json:"tenantId"`
	UserID         uuid.UUID       `json:"userId"`
	Month          Month           `json:"month"`
	Status         TimesheetStatus `json:"status"`
	CommercialInfo CommercialInfo  `json:"commercialInfo"`
	Weeks          []WeekEntry     `json:"weeks"`
	ValidatedAt    *time.Time      `json:"validatedAt,omitempty"`
	ValidatedBy    *uuid.UUID      `json:"validatedBy,omitempty"`
	RejectedAt     *time.Time      `json:"rejectedAt,omitempty"`
	RejectedBy     *uuid.UUID      `json:"rejectedBy,omitempty"`
	RejectReason   string          `json:"rejectReason,omitempty"`
}

type TimesheetSummary struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"userId"`
	UserLogin      string          `json:"userLogin"`
	UserPrenom     string          `json:"userPrenom"`
	UserNom        string          `json:"userNom"`
	Month          Month           `json:"month"`
	Status         TimesheetStatus `json:"status"`
	CommercialInfo CommercialInfo  `json:"commercialInfo"`
	ClientID       *uuid.UUID      `json:"clientId,omitempty"`
	MissionID      *uuid.UUID      `json:"missionId,omitempty"`
	TotalMinutes   int             `json:"totalMinutes"`
	WeeksSubmitted int             `json:"weeksSubmitted"`
	WeeksTotal     int             `json:"weeksTotal"`
	RejectReason   string          `json:"rejectReason,omitempty"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

func (ts Timesheet) IsFinal() bool {
	return ts.Status == StatusDefinitif
}

func (ts Timesheet) CanEdit() bool {
	return ts.Status != StatusDefinitif
}

func (ts *Timesheet) Reject(now time.Time, managerID uuid.UUID, reason string) error {
	if ts.IsFinal() {
		return ErrCRAAlreadyValidated
	}
	ts.Status = StatusBrouillon
	t := now.UTC()
	ts.RejectedAt = &t
	ts.RejectedBy = &managerID
	ts.RejectReason = reason
	for i := range ts.Weeks {
		ts.Weeks[i].SubmittedAt = nil
	}
	return nil
}

func lineKey(source SourceRef, day time.Time) string {
	return fmt.Sprintf("%s:%s:%s", source.Type, source.ID, day.Format("2006-01-02"))
}

func FindLine(lines []TimeLine, source SourceRef, day time.Time) (*TimeLine, int) {
	key := lineKey(source, day)
	for i := range lines {
		if lineKey(lines[i].Source, lines[i].Day) == key {
			return &lines[i], i
		}
	}
	return nil, -1
}

// ApplyProposedLines merges proposed lines without overwriting manual entries (RG-CRA-01).
func ApplyProposedLines(week *WeekEntry, proposed []TimeLine, dayCapacityMinutes int) error {
	for _, p := range proposed {
		existing, _ := FindLine(week.Lines, p.Source, p.Day)
		if existing != nil {
			if existing.Origin == OriginManual {
				continue
			}
			existing.Duration = p.Duration
			existing.Comment = p.Comment
			existing.Origin = OriginPrefill
			continue
		}
		if p.ID == uuid.Nil {
			p.ID = uuid.New()
		}
		p.Billable = true
		p.Origin = OriginPrefill
		week.Lines = append(week.Lines, p)
	}
	return ValidateDayCapacity(week.Lines, dayCapacityMinutes)
}

func ValidateDayCapacity(lines []TimeLine, dayCapacityMinutes int) error {
	if dayCapacityMinutes <= 0 {
		dayCapacityMinutes = DefaultDayCapacityMinutes
	}
	totals := make(map[string]int)
	for _, line := range lines {
		key := line.Day.Format("2006-01-02")
		totals[key] += line.Duration.Minutes
		if totals[key] > dayCapacityMinutes {
			return ErrDayCapacityExceeded
		}
	}
	return DetectAbsenceConflict(lines)
}

// IncompleteDaysInWeek returns in-month days in the week tab with zero total minutes.
func IncompleteDaysInWeek(month Month, weekNumber WeekNumber, weekStartDay int, lines []TimeLine) ([]string, error) {
	days, err := WeekDaysInMonth(month, weekNumber, weekStartDay)
	if err != nil {
		return nil, err
	}
	totals := make(map[string]int)
	for _, line := range lines {
		if line.Duration.Minutes <= 0 {
			continue
		}
		totals[line.Day.Format("2006-01-02")] += line.Duration.Minutes
	}
	var missing []string
	for _, day := range days {
		if totals[day] <= 0 {
			missing = append(missing, day)
		}
	}
	return missing, nil
}

func DetectAbsenceConflict(lines []TimeLine) error {
	byDay := make(map[string][]SourceRef)
	for _, line := range lines {
		if line.Duration.Minutes == 0 {
			continue
		}
		key := line.Day.Format("2006-01-02")
		byDay[key] = append(byDay[key], line.Source)
	}
	for day, sources := range byDay {
		hasAbsence := false
		hasMission := false
		for _, s := range sources {
			if isAbsenceSource(s) {
				hasAbsence = true
			} else {
				hasMission = true
			}
		}
		if hasAbsence && hasMission {
			return fmt.Errorf("%w on %s", ErrCRAConflictAbsence, day)
		}
	}
	return nil
}

func isAbsenceSource(s SourceRef) bool {
	switch s.Type {
	case "absence", "conge", "leave":
		return true
	default:
		return false
	}
}

func (ts *Timesheet) Week(number WeekNumber) (*WeekEntry, int) {
	for i := range ts.Weeks {
		if ts.Weeks[i].WeekNumber == number {
			return &ts.Weeks[i], i
		}
	}
	return nil, -1
}

func (ts *Timesheet) EnsureWeek(number WeekNumber) *WeekEntry {
	if week, _ := ts.Week(number); week != nil {
		return week
	}
	entry := WeekEntry{
		ID:          uuid.New(),
		TenantID:    ts.TenantID,
		TimesheetID: ts.ID,
		WeekNumber:  number,
	}
	ts.Weeks = append(ts.Weeks, entry)
	return &ts.Weeks[len(ts.Weeks)-1]
}

type Consumption struct {
	UserID   uuid.UUID
	Source   SourceRef
	Day      time.Time
	Duration kernel.Duration
}

type Document struct {
	Filename string
	Content  []byte
	MimeType string
}
