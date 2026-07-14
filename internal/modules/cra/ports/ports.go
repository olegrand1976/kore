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
	ValidateFinal(ctx context.Context, cmd ManagerValidateCommand) error
	ValidateAll(ctx context.Context, cmd ValidateAllCommand) (ValidateAllResult, error)
	RejectTimesheet(ctx context.Context, cmd RejectTimesheetCommand) error
	PrefillPublicHolidays(ctx context.Context, tenant kernel.TenantID, userID UserID, month domain.Month, countryCode string) (int, error)
	ExportPrestationsXML(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]PrestationExportRow, error)
	BillableSummary(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]BillableUserSummary, error)
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
