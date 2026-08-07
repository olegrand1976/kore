package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type stubAppBilling struct {
	mode string
	tjm  int64
	err  error
}

func (s stubAppBilling) GetApplicationBilling(context.Context, kernel.TenantID, uuid.UUID) (ports.ApplicationBillingInfo, error) {
	if s.err != nil {
		return ports.ApplicationBillingInfo{}, s.err
	}
	return ports.ApplicationBillingInfo{ModeFacturation: s.mode, DefaultTJMCents: s.tjm}, nil
}

type stubMissionRate struct {
	tjm      int64
	currency string
}

func (s stubMissionRate) GetMissionRate(context.Context, kernel.TenantID, uuid.UUID) (ports.MissionRate, error) {
	return ports.MissionRate{TJMAmount: s.tjm, Currency: s.currency}, nil
}

type recordingPublisher struct {
	lastCmd   ports.ValidationInvoiceCommand
	invoiceID uuid.UUID
	exists    bool
}

func (p *recordingPublisher) PublishCRAValidationDraft(_ context.Context, cmd ports.ValidationInvoiceCommand) (uuid.UUID, error) {
	p.lastCmd = cmd
	if p.invoiceID == uuid.Nil {
		p.invoiceID = uuid.New()
	}
	return p.invoiceID, nil
}

func (p *recordingPublisher) TimesheetAlreadyInvoiced(context.Context, kernel.TenantID, uuid.UUID) (bool, error) {
	return p.exists, nil
}

func TestInvoiceableMinutes_ScopesToDominantMission(t *testing.T) {
	missionA := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	missionB := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	ts := domain.Timesheet{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Month:  domain.Month("2026-08"),
		CommercialInfo: domain.CommercialInfo{
			MissionID: &missionA,
		},
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "mission", ID: missionA.String()}, Duration: kernel.Duration{Minutes: 120}, Billable: true},
				{Source: domain.SourceRef{Type: "mission", ID: missionB.String()}, Duration: kernel.Duration{Minutes: 480}, Billable: true},
				{Source: domain.SourceRef{Type: "manual", ID: "x"}, Duration: kernel.Duration{Minutes: 60}, Billable: true},
			},
		}},
	}
	svc := (&Service{}).WithApplicationBillingReader(stubAppBilling{mode: orgdomain.ModeFacturationTempsPasse})
	minutes, reason, err := svc.invoiceableBillableMinutes(context.Background(), ts)
	require.NoError(t, err)
	require.Empty(t, reason)
	require.Equal(t, 120, minutes)
}

func TestInvoiceableMinutes_AppNonSkips(t *testing.T) {
	appID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	ts := domain.Timesheet{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Month:  domain.Month("2026-08"),
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "application", ID: appID.String()}, Duration: kernel.Duration{Minutes: 240}, Billable: true},
			},
		}},
	}
	svc := (&Service{}).WithApplicationBillingReader(stubAppBilling{mode: orgdomain.ModeFacturationNon})
	minutes, reason, err := svc.invoiceableBillableMinutes(context.Background(), ts)
	require.NoError(t, err)
	require.Equal(t, 0, minutes)
	require.Equal(t, "billing_mode_disabled", reason)
}

func TestInvoiceableMinutes_AppLookupFailClosed(t *testing.T) {
	appID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	ts := domain.Timesheet{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Month:  domain.Month("2026-08"),
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "application", ID: appID.String()}, Duration: kernel.Duration{Minutes: 240}, Billable: true},
			},
		}},
	}
	svc := (&Service{}).WithApplicationBillingReader(stubAppBilling{err: errors.New("db down")})
	minutes, reason, err := svc.invoiceableBillableMinutes(context.Background(), ts)
	require.NoError(t, err)
	require.Equal(t, 0, minutes)
	require.Equal(t, "billing_mode_unresolved", reason)
}

func TestPreviewInvoices_OK(t *testing.T) {
	clientID := uuid.New()
	missionID := uuid.New()
	ts := domain.Timesheet{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Month:    domain.Month("2026-08"),
		Status:   domain.StatusDefinitif,
		CommercialInfo: domain.CommercialInfo{
			ClientID:  &clientID,
			MissionID: &missionID,
			Mission:   "Mission Alpha",
		},
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "mission", ID: missionID.String()}, Duration: kernel.Duration{Minutes: 180}, Billable: true},
			},
		}},
	}
	repo := &fakeCRARepo{ts: ts}
	pub := &recordingPublisher{}
	svc := NewService(repo, nil, nil).
		WithInvoicePublisher(pub).
		WithMissionRateReader(stubMissionRate{tjm: 80000, currency: "EUR"})

	previews, err := svc.PreviewInvoicesFromTimesheets(context.Background(), ts.TenantID, []uuid.UUID{ts.ID})
	require.NoError(t, err)
	require.Len(t, previews, 1)
	require.True(t, previews[0].OK)
	require.Empty(t, previews[0].Blockers)
	require.Equal(t, clientID, *previews[0].ClientID)
	require.Equal(t, 3.0, previews[0].BillableHours)
	require.Greater(t, previews[0].UnitPriceCents, int64(0))
	require.Contains(t, previews[0].Description, "Mission Alpha")
}

func TestPreviewInvoices_Blockers(t *testing.T) {
	ts := domain.Timesheet{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Month:    domain.Month("2026-08"),
		Status:   domain.StatusDefinitif,
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "manual", ID: "x"}, Duration: kernel.Duration{Minutes: 60}, Billable: true},
			},
		}},
	}
	repo := &fakeCRARepo{ts: ts}
	svc := NewService(repo, nil, nil).WithInvoicePublisher(&recordingPublisher{})

	previews, err := svc.PreviewInvoicesFromTimesheets(context.Background(), ts.TenantID, []uuid.UUID{ts.ID})
	require.NoError(t, err)
	require.Len(t, previews, 1)
	require.False(t, previews[0].OK)
	require.Contains(t, previews[0].Blockers, "client_unresolved")
	require.Contains(t, previews[0].Blockers, "zero_unit_price")
}

func TestPreviewInvoices_AlreadyExists(t *testing.T) {
	clientID := uuid.New()
	ts := domain.Timesheet{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Month:    domain.Month("2026-08"),
		Status:   domain.StatusDefinitif,
		CommercialInfo: domain.CommercialInfo{
			ClientID: &clientID,
		},
	}
	repo := &fakeCRARepo{ts: ts}
	svc := NewService(repo, nil, nil).WithInvoicePublisher(&recordingPublisher{exists: true})

	previews, err := svc.PreviewInvoicesFromTimesheets(context.Background(), ts.TenantID, []uuid.UUID{ts.ID})
	require.NoError(t, err)
	require.False(t, previews[0].OK)
	require.Equal(t, []string{"already_exists"}, previews[0].Blockers)
}

func TestCreateInvoicesFromTimesheetItems_Overrides(t *testing.T) {
	clientID := uuid.New()
	overrideClient := uuid.New()
	missionID := uuid.New()
	ts := domain.Timesheet{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Month:    domain.Month("2026-08"),
		Status:   domain.StatusDefinitif,
		CommercialInfo: domain.CommercialInfo{
			ClientID:  &clientID,
			MissionID: &missionID,
			Mission:   "Mission Alpha",
		},
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "mission", ID: missionID.String()}, Duration: kernel.Duration{Minutes: 180}, Billable: true},
			},
		}},
	}
	repo := &fakeCRARepo{ts: ts}
	pub := &recordingPublisher{}
	svc := NewService(repo, nil, nil).
		WithInvoicePublisher(pub).
		WithMissionRateReader(stubMissionRate{tjm: 80000, currency: "EUR"})

	hours := 5.5
	price := int64(12345)
	tax := 10.0
	desc := "Override description"
	outcomes, err := svc.CreateInvoicesFromTimesheetItems(context.Background(), ts.TenantID, []ports.CreateInvoiceFromTimesheetItem{{
		TimesheetID:    ts.ID,
		ClientID:       &overrideClient,
		BillableHours:  &hours,
		UnitPriceCents: &price,
		TaxRate:        &tax,
		Description:    &desc,
	}})
	require.NoError(t, err)
	require.Len(t, outcomes, 1)
	require.Equal(t, ports.InvoiceDraftCreated, outcomes[0].Status)
	require.Equal(t, overrideClient, pub.lastCmd.ClientID)
	require.Equal(t, hours, pub.lastCmd.BillableHours)
	require.Equal(t, price, pub.lastCmd.UnitPriceCents)
	require.Equal(t, tax, pub.lastCmd.TaxRate)
	require.Equal(t, desc, pub.lastCmd.Description)
}

func TestCreateInvoicesFromTimesheetItems_HardBlockerNotBypassed(t *testing.T) {
	appID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	clientID := uuid.New()
	ts := domain.Timesheet{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Month:    domain.Month("2026-08"),
		Status:   domain.StatusDefinitif,
		CommercialInfo: domain.CommercialInfo{
			ClientID: &clientID,
		},
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "application", ID: appID.String()}, Duration: kernel.Duration{Minutes: 240}, Billable: true},
			},
		}},
	}
	repo := &fakeCRARepo{ts: ts}
	pub := &recordingPublisher{}
	svc := NewService(repo, nil, nil).
		WithInvoicePublisher(pub).
		WithApplicationBillingReader(stubAppBilling{mode: orgdomain.ModeFacturationNon})

	hours := 4.0
	price := int64(10000)
	outcomes, err := svc.CreateInvoicesFromTimesheetItems(context.Background(), ts.TenantID, []ports.CreateInvoiceFromTimesheetItem{{
		TimesheetID:    ts.ID,
		ClientID:       &clientID,
		BillableHours:  &hours,
		UnitPriceCents: &price,
	}})
	require.NoError(t, err)
	require.Len(t, outcomes, 1)
	require.Equal(t, ports.InvoiceDraftSkipped, outcomes[0].Status)
	require.Equal(t, "billing_mode_disabled", outcomes[0].Reason)
}

func TestCreateInvoicesFromTimesheetItems_SoftBlockerFixedByOverride(t *testing.T) {
	ts := domain.Timesheet{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		TenantID: kernel.NewTenantID(uuid.New()),
		Month:    domain.Month("2026-08"),
		Status:   domain.StatusDefinitif,
		Weeks: []domain.WeekEntry{{
			Lines: []domain.TimeLine{
				{Source: domain.SourceRef{Type: "manual", ID: "x"}, Duration: kernel.Duration{Minutes: 120}, Billable: true},
			},
		}},
	}
	repo := &fakeCRARepo{ts: ts}
	pub := &recordingPublisher{}
	svc := NewService(repo, nil, nil).WithInvoicePublisher(pub)

	clientID := uuid.New()
	hours := 2.0
	price := int64(5000)
	outcomes, err := svc.CreateInvoicesFromTimesheetItems(context.Background(), ts.TenantID, []ports.CreateInvoiceFromTimesheetItem{{
		TimesheetID:    ts.ID,
		ClientID:       &clientID,
		BillableHours:  &hours,
		UnitPriceCents: &price,
	}})
	require.NoError(t, err)
	require.Len(t, outcomes, 1)
	require.Equal(t, ports.InvoiceDraftCreated, outcomes[0].Status)
	require.Equal(t, clientID, pub.lastCmd.ClientID)
	require.Equal(t, hours, pub.lastCmd.BillableHours)
	require.Equal(t, price, pub.lastCmd.UnitPriceCents)
}
