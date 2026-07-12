package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/adapters/pdf"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
)

const consumptionCacheTTL = 5 * time.Minute

type Service struct {
	repo     ports.CRARepository
	cache    cache.Cache
	keys     cache.KeyBuilder
	pdf      ports.PDFRenderer
	clock    ports.Clock
}

func NewService(repo ports.CRARepository, appCache cache.Cache, keys cache.KeyBuilder) *Service {
	return &Service{
		repo:  repo,
		cache: appCache,
		keys:  keys,
		pdf:   pdf.NewStubRenderer(),
		clock: ports.RealClock{},
	}
}

func (s *Service) WithPDFRenderer(renderer ports.PDFRenderer) *Service {
	if renderer != nil {
		s.pdf = renderer
	}
	return s
}

func (s *Service) WithClock(clock ports.Clock) *Service {
	if clock != nil {
		s.clock = clock
	}
	return s
}

func (s *Service) GetOrCreate(ctx context.Context, tenant kernel.TenantID, userID ports.UserID, month domain.Month) (domain.Timesheet, error) {
	ts, err := s.repo.Get(ctx, tenant, userID, month)
	if err == nil {
		return ts, nil
	}
	if !errors.Is(err, domain.ErrTimesheetNotFound) {
		return domain.Timesheet{}, err
	}
	ts = domain.Timesheet{
		ID:       uuid.New(),
		TenantID: tenant,
		UserID:   userID,
		Month:    month,
		Status:   domain.StatusBrouillon,
	}
	if err := s.repo.Save(ctx, ts); err != nil {
		return domain.Timesheet{}, err
	}
	return ts, nil
}

func (s *Service) GetByID(ctx context.Context, tenant kernel.TenantID, id ports.TimesheetID) (domain.Timesheet, error) {
	return s.repo.GetByID(ctx, tenant, id)
}

func (s *Service) ListTimesheets(ctx context.Context, tenant kernel.TenantID, userID ports.UserID, managerView bool, limit int) ([]domain.Timesheet, error) {
	if limit <= 0 {
		limit = 12
	}
	if managerView {
		return s.repo.ListByTenant(ctx, tenant, limit)
	}
	return s.repo.ListByUser(ctx, tenant, userID, limit)
}

func (s *Service) SaveWeek(ctx context.Context, cmd ports.SaveWeekCommand) (domain.Timesheet, error) {
	ts, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return domain.Timesheet{}, err
	}
	if !ts.CanEdit() {
		return domain.Timesheet{}, domain.ErrCRAAlreadyValidated
	}
	week := ts.EnsureWeek(cmd.WeekNumber)
	for _, line := range cmd.Lines {
		line.TenantID = cmd.TenantID
		line.WeekEntryID = week.ID
		line.Origin = domain.OriginManual
		if line.ID == uuid.Nil {
			line.ID = uuid.New()
		}
		existing, idx := domain.FindLine(week.Lines, line.Source, line.Day)
		if existing != nil {
			week.Lines[idx] = line
		} else {
			week.Lines = append(week.Lines, line)
		}
	}
	if err := domain.ValidateDayCapacity(week.Lines); err != nil {
		return domain.Timesheet{}, err
	}
	if err := s.repo.Save(ctx, ts); err != nil {
		return domain.Timesheet{}, err
	}
	s.invalidateConsumptionCache(ctx, cmd.TenantID)
	return ts, nil
}

func (s *Service) SubmitWeek(ctx context.Context, cmd ports.SubmitWeekCommand) error {
	ts, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return err
	}
	if !ts.CanEdit() {
		return domain.ErrCRAAlreadyValidated
	}
	week, _ := ts.Week(cmd.WeekNumber)
	if week == nil {
		return domain.ErrWeekNotFound
	}
	now := s.clock.Now().UTC()
	week.SubmittedAt = &now
	if ts.Status == domain.StatusBrouillon {
		ts.Status = domain.StatusValideSemaine
	}
	if err := s.repo.Save(ctx, ts); err != nil {
		return err
	}
	s.invalidateConsumptionCache(ctx, cmd.TenantID)
	return nil
}

func (s *Service) CompleteCommercialInfo(ctx context.Context, cmd ports.CommercialCommand) error {
	ts, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return err
	}
	if !ts.CanEdit() {
		return domain.ErrCRAAlreadyValidated
	}
	ts.CommercialInfo = cmd.Info
	return s.repo.Save(ctx, ts)
}

func (s *Service) GeneratePDF(ctx context.Context, tenant kernel.TenantID, id ports.TimesheetID) (domain.Document, error) {
	ts, err := s.repo.GetByID(ctx, tenant, id)
	if err != nil {
		return domain.Document{}, err
	}
	if !ts.CommercialInfo.Complete() {
		return domain.Document{}, domain.ErrCommercialInfoRequired
	}
	return s.pdf.Render(ctx, ts)
}

func (s *Service) ValidateFinal(ctx context.Context, cmd ports.ManagerValidateCommand) error {
	ts, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return err
	}
	if ts.IsFinal() {
		return domain.ErrCRAAlreadyValidated
	}
	now := s.clock.Now().UTC()
	ts.Status = domain.StatusDefinitif
	ts.ValidatedAt = &now
	ts.ValidatedBy = &cmd.ManagerID
	if err := s.repo.Save(ctx, ts); err != nil {
		return err
	}
	s.invalidateConsumptionCache(ctx, cmd.TenantID)
	return nil
}

func (s *Service) ProposeLines(ctx context.Context, lines []ports.ProposedLine) error {
	if len(lines) == 0 {
		return nil
	}
	type sheetGroup struct {
		tenant kernel.TenantID
		userID uuid.UUID
		month  domain.Month
	}
	grouped := make(map[sheetGroup][]ports.ProposedLine)
	for _, line := range lines {
		month := line.Month
		if month == "" {
			month = domain.Month(line.Day.Format("2006-01"))
		}
		key := sheetGroup{tenant: line.TenantID, userID: line.UserID, month: month}
		grouped[key] = append(grouped[key], line)
	}
	for key, batch := range grouped {
		ts, err := s.GetOrCreate(ctx, key.tenant, key.userID, key.month)
		if err != nil {
			return err
		}
		if !ts.CanEdit() {
			return domain.ErrCRAAlreadyValidated
		}
		byWeek := make(map[domain.WeekNumber][]ports.ProposedLine)
		for _, line := range batch {
			weekNum := line.WeekNumber
			if weekNum == 0 {
				_, isoWeek := line.Day.ISOWeek()
				weekNum = domain.WeekNumber(isoWeek)
			}
			byWeek[weekNum] = append(byWeek[weekNum], line)
		}
		for weekNum, weekLines := range byWeek {
			week := ts.EnsureWeek(weekNum)
			proposed := make([]domain.TimeLine, 0, len(weekLines))
			for _, pl := range weekLines {
				proposed = append(proposed, domain.TimeLine{
					ID:          uuid.New(),
					TenantID:    pl.TenantID,
					WeekEntryID: week.ID,
					Source:      pl.Source,
					Day:         pl.Day,
					Duration:    pl.Duration,
					Comment:     pl.Comment,
					Origin:      domain.OriginPrefill,
				})
			}
			if err := domain.ApplyProposedLines(week, proposed); err != nil {
				return err
			}
		}
		if err := s.repo.Save(ctx, ts); err != nil {
			return err
		}
		s.invalidateConsumptionCache(ctx, key.tenant)
	}
	return nil
}

func (s *Service) RemoveFutureLines(ctx context.Context, source domain.SourceRef, from time.Time) error {
	identity, ok := authx.FromContext(ctx)
	if !ok {
		return errors.New("tenant context required")
	}
	if err := s.repo.DeleteFutureLines(ctx, identity.TenantID, source, from); err != nil {
		return err
	}
	s.invalidateConsumptionCache(ctx, identity.TenantID)
	return nil
}

func (s *Service) ConsumedByApplication(ctx context.Context, tenant kernel.TenantID, appID ports.ApplicationID, period kernel.Period) ([]domain.Consumption, error) {
	key := s.keys.Key(tenant, "cra", "consumption", appID.String(), period.Start.Format("2006-01-02"), period.End.Format("2006-01-02"))
	var out []domain.Consumption
	err := s.cache.GetOrLoad(ctx, key, consumptionCacheTTL, func(ctx context.Context) (any, error) {
		return s.repo.FindConsumption(ctx, tenant, appID, period)
	}, &out)
	return out, err
}

func (s *Service) TimesheetOf(ctx context.Context, tenant kernel.TenantID, userID ports.UserID, month domain.Month) (domain.Timesheet, error) {
	return s.repo.Get(ctx, tenant, userID, month)
}

func (s *Service) invalidateConsumptionCache(ctx context.Context, tenant kernel.TenantID) {
	_ = ctx
	_ = tenant
}

var (
	_ ports.CRAService       = (*Service)(nil)
	_ ports.CRAFeeder        = (*Service)(nil)
	_ ports.CRAReader        = (*Service)(nil)
	_ ports.CRAFutureCleaner = (*Service)(nil)
)
