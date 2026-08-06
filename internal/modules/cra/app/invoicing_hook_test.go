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
	err  error
}

func (s stubAppBilling) GetApplicationBilling(context.Context, kernel.TenantID, uuid.UUID) (ports.ApplicationBillingInfo, error) {
	if s.err != nil {
		return ports.ApplicationBillingInfo{}, s.err
	}
	return ports.ApplicationBillingInfo{ModeFacturation: s.mode}, nil
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
