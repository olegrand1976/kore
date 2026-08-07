package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/kernel"
)

type UserID = uuid.UUID
type TimesheetID = uuid.UUID
type ApplicationID = uuid.UUID

type SaveWeekCommand struct {
	TenantID    kernel.TenantID
	TimesheetID TimesheetID
	WeekNumber  domain.WeekNumber
	Lines       []domain.TimeLine
}

type SubmitWeekCommand struct {
	TenantID    kernel.TenantID
	TimesheetID TimesheetID
	WeekNumber  domain.WeekNumber
	UserID      UserID
}

type CommercialCommand struct {
	TenantID    kernel.TenantID
	TimesheetID TimesheetID
	Info        domain.CommercialInfo
}

type ManagerValidateCommand struct {
	TenantID    kernel.TenantID
	TimesheetID TimesheetID
	ManagerID   UserID
}

type InvoiceDraftStatus string

const (
	InvoiceDraftCreated     InvoiceDraftStatus = "created"
	InvoiceDraftSkipped     InvoiceDraftStatus = "skipped"
	InvoiceDraftUnavailable InvoiceDraftStatus = "unavailable"
)

type InvoiceDraftOutcome struct {
	Status      InvoiceDraftStatus `json:"status"`
	Reason      string             `json:"reason,omitempty"`
	InvoiceID   *uuid.UUID         `json:"invoiceId,omitempty"`
	TimesheetID *uuid.UUID         `json:"timesheetId,omitempty"`
}

type ValidateFinalResult struct {
	InvoiceDraft InvoiceDraftOutcome `json:"invoiceDraft"`
}

type RejectTimesheetCommand struct {
	TenantID    kernel.TenantID
	TimesheetID TimesheetID
	ManagerID   UserID
	Reason      string
}

type ValidateAllCommand struct {
	TenantID  kernel.TenantID
	ManagerID UserID
	Month     domain.Month
}

type ValidateAllResult struct {
	Validated int                  `json:"validated"`
	Failed    []ValidateAllFailure `json:"failed,omitempty"`
}

type ValidateAllFailure struct {
	TimesheetID TimesheetID `json:"timesheetId"`
	Reason      string      `json:"reason"`
}

type ProposedLine struct {
	TenantID   kernel.TenantID
	UserID     UserID
	Month      domain.Month
	WeekNumber domain.WeekNumber
	Source     domain.SourceRef
	Day        time.Time
	Duration   kernel.Duration
	Comment    string
}

type ValidationInvoiceCommand struct {
	TenantID        kernel.TenantID
	TimesheetID     TimesheetID
	TimesheetUserID uuid.UUID
	ClientID        uuid.UUID
	Month           domain.Month
	BillableHours   float64
	MissionLabel    string
	UserLabel       string
	Currency        string
	UnitPriceCents  int64
	TaxRate         float64
	Description     string // optional override; empty → auto description
}

// InvoiceDraftPreview is a dry-run CRA→invoice payload (no persistence).
type InvoiceDraftPreview struct {
	TimesheetID    TimesheetID `json:"timesheetId"`
	OK             bool        `json:"ok"`
	Blockers       []string    `json:"blockers,omitempty"`
	ClientID       *uuid.UUID  `json:"clientId,omitempty"`
	BillableHours  float64     `json:"billableHours,omitempty"`
	UnitPriceCents int64       `json:"unitPriceCents,omitempty"`
	Currency       string      `json:"currency,omitempty"`
	TaxRate        float64     `json:"taxRate,omitempty"`
	Description    string      `json:"description,omitempty"`
	MissionLabel   string      `json:"missionLabel,omitempty"`
	UserLabel      string      `json:"userLabel,omitempty"`
}

// CreateInvoiceFromTimesheetItem creates or overrides a CRA-linked invoice draft.
type CreateInvoiceFromTimesheetItem struct {
	TimesheetID    TimesheetID `json:"timesheetId"`
	ClientID       *uuid.UUID  `json:"clientId,omitempty"`
	BillableHours  *float64    `json:"billableHours,omitempty"`
	UnitPriceCents *int64      `json:"unitPriceCents,omitempty"`
	TaxRate        *float64    `json:"taxRate,omitempty"`
	Currency       *string     `json:"currency,omitempty"`
	Description    *string     `json:"description,omitempty"`
	MissionLabel   *string     `json:"missionLabel,omitempty"`
}

type InvoiceDraftPublisher interface {
	PublishCRAValidationDraft(ctx context.Context, cmd ValidationInvoiceCommand) (uuid.UUID, error)
	TimesheetAlreadyInvoiced(ctx context.Context, tenant kernel.TenantID, timesheetID uuid.UUID) (bool, error)
}

type MissionRate struct {
	TJMAmount int64  // cents — daily or hourly depending on RateUnit
	RateUnit  string // tjm | hourly
	Currency  string
}

type MissionRateReader interface {
	GetMissionRate(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID) (MissionRate, error)
}

type ETTDayHours struct {
	WorkDate time.Time
	Hours    float64
}

type ETTRecordReader interface {
	ListUserDayHours(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID, from, to time.Time) ([]ETTDayHours, error)
}

type DailyActivityRow struct {
	UserID       uuid.UUID
	UserPrenom   string
	UserNom      string
	Day          time.Time
	Minutes      int
	MissionID    string
	MissionLabel string
	ClientLabel  string
}

type CRAService interface {
	GetOrCreate(ctx context.Context, tenant kernel.TenantID, userID UserID, month domain.Month) (domain.Timesheet, error)
	GetByID(ctx context.Context, tenant kernel.TenantID, id TimesheetID) (domain.Timesheet, error)
	ListTimesheets(ctx context.Context, tenant kernel.TenantID, userID UserID, managerView bool, limit int) ([]domain.Timesheet, error)
	ListTimesheetSummaries(ctx context.Context, tenant kernel.TenantID, userID UserID, managerView bool, limit int) ([]domain.TimesheetSummary, error)
	ListPrestations(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]domain.TimesheetSummary, error)
	SaveWeek(ctx context.Context, cmd SaveWeekCommand) (domain.Timesheet, error)
	SubmitWeek(ctx context.Context, cmd SubmitWeekCommand) error
	CompleteCommercialInfo(ctx context.Context, cmd CommercialCommand) error
	GeneratePDF(ctx context.Context, tenant kernel.TenantID, id TimesheetID) (domain.Document, error)
	ValidateFinal(ctx context.Context, cmd ManagerValidateCommand) (ValidateFinalResult, error)
	ValidateAll(ctx context.Context, cmd ValidateAllCommand) (ValidateAllResult, error)
	CreateInvoicesFromTimesheets(ctx context.Context, tenant kernel.TenantID, ids []uuid.UUID) ([]InvoiceDraftOutcome, error)
	CreateInvoicesFromTimesheetItems(ctx context.Context, tenant kernel.TenantID, items []CreateInvoiceFromTimesheetItem) ([]InvoiceDraftOutcome, error)
	PreviewInvoicesFromTimesheets(ctx context.Context, tenant kernel.TenantID, ids []uuid.UUID) ([]InvoiceDraftPreview, error)
	RejectTimesheet(ctx context.Context, cmd RejectTimesheetCommand) error
	PrefillPublicHolidays(ctx context.Context, tenant kernel.TenantID, userID UserID, month domain.Month, countryCode string) (int, error)
	PrefillFromETT(ctx context.Context, tenant kernel.TenantID, userID UserID, month domain.Month) (int, error)
	ExportPrestationsXML(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]PrestationExportRow, error)
	BillableSummary(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]BillableUserSummary, error)
	ListDailyActivityInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]DailyActivityRow, error)
}

type PrestationExportRow struct {
	UserLogin     string  `json:"userLogin" xml:"userLogin"`
	UserName      string  `json:"userName" xml:"userName"`
	Month         string  `json:"month" xml:"month"`
	Status        string  `json:"status" xml:"status"`
	TotalHours    float64 `json:"totalHours" xml:"totalHours"`
	BillableHours float64 `json:"billableHours" xml:"billableHours"`
	WeeksRatio    string  `json:"weeksRatio" xml:"weeksRatio"`
}

type BillableUserSummary struct {
	UserID        uuid.UUID `json:"userId"`
	UserLogin     string    `json:"userLogin"`
	BillableHours float64   `json:"billableHours"`
}

type CRAFeeder interface {
	ProposeLines(ctx context.Context, lines []ProposedLine) error
}

type CRAFutureCleaner interface {
	RemoveFutureLines(ctx context.Context, source domain.SourceRef, from time.Time) error
}

type CRAReader interface {
	ConsumedByApplication(ctx context.Context, tenant kernel.TenantID, appID ApplicationID, period kernel.Period) ([]domain.Consumption, error)
	TimesheetOf(ctx context.Context, tenant kernel.TenantID, userID UserID, month domain.Month) (domain.Timesheet, error)
}

type SocieteCraSettings struct {
	WeekStartDay       int
	DayCapacityMinutes int
	WeekSubmitPolicy   string
	CraMailAuto        bool
	CraMailRecipients  []string
	TaskTypesEnabled   []string
	DefaultTJMCents    int64
}

type ApplicationBillingInfo struct {
	ModeFacturation string
	DefaultTJMCents int64
}

type ApplicationBillingReader interface {
	GetApplicationBilling(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) (ApplicationBillingInfo, error)
}

type UserEmailResolver interface {
	ResolveUserEmails(ctx context.Context, tenant kernel.TenantID, userIDs []uuid.UUID) ([]string, error)
}

type SocieteCalendarReader interface {
	SettingsForUser(ctx context.Context, tenant kernel.TenantID, userID UserID) (SocieteCraSettings, error)
}

type CRARepository interface {
	Save(ctx context.Context, ts domain.Timesheet) error
	Get(ctx context.Context, tenant kernel.TenantID, userID UserID, month domain.Month) (domain.Timesheet, error)
	GetByID(ctx context.Context, tenant kernel.TenantID, id TimesheetID) (domain.Timesheet, error)
	FindConsumption(ctx context.Context, tenant kernel.TenantID, appID ApplicationID, period kernel.Period) ([]domain.Consumption, error)
	ListByUser(ctx context.Context, tenant kernel.TenantID, userID UserID, limit int) ([]domain.Timesheet, error)
	ListByTenant(ctx context.Context, tenant kernel.TenantID, limit int) ([]domain.Timesheet, error)
	ListSummariesByUser(ctx context.Context, tenant kernel.TenantID, userID UserID, limit int) ([]domain.TimesheetSummary, error)
	ListSummariesByTenant(ctx context.Context, tenant kernel.TenantID, limit int) ([]domain.TimesheetSummary, error)
	ListSummariesByTenantMonth(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]domain.TimesheetSummary, error)
	ListReminderCandidatesByMonth(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]domain.ReminderCandidate, error)
	ListDailyActivityInPeriod(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]DailyActivityRow, error)
	DeleteFutureLines(ctx context.Context, tenant kernel.TenantID, source domain.SourceRef, from time.Time) error
}

type PDFRenderer interface {
	Render(ctx context.Context, ts domain.Timesheet) (domain.Document, error)
}

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
