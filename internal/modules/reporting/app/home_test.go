package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/reporting/domain"
	"github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubHomeUser struct{ requis bool }

func (s stubHomeUser) GetCraRequis(context.Context, kernel.TenantID, uuid.UUID) (bool, error) {
	return s.requis, nil
}

type stubHomeCRA struct{ items []ports.HomeCRATimesheet }

func (s stubHomeCRA) ListRecentSummaries(context.Context, kernel.TenantID, uuid.UUID, int) ([]ports.HomeCRATimesheet, error) {
	return s.items, nil
}

type stubHomeLeave struct{ statuses []string }

func (s stubHomeLeave) ListStatuses(context.Context, kernel.TenantID, *uuid.UUID) ([]string, error) {
	return s.statuses, nil
}

type stubTMAOK struct{}

func (stubTMAOK) SummaryStats(context.Context, kernel.TenantID, time.Time) (ports.TMASummaryStats, error) {
	return ports.TMASummaryStats{}, nil
}

func (stubTMAOK) HomeStats(context.Context, kernel.TenantID) (ports.HomeTMAStats, error) {
	return ports.HomeTMAStats{
		Open:  2,
		Total: 5,
		StatusCounts: []domain.HomeStatusCount{
			{Key: "ouverte", Value: 2},
			{Key: "resolue", Value: 3},
		},
	}, nil
}

func TestGetHomeDashboard_CRAAlertWithoutTimesheet(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, stubTMAOK{},
		WithHomeReaders(
			stubHomeUser{requis: true},
			stubHomeCRA{},
			stubHomeLeave{statuses: []string{"en_attente", "valide"}},
			nil,
		),
	)

	home, err := svc.GetHomeDashboard(context.Background(), ports.HomeDashboardQuery{
		TenantID:         kernel.NewTenantID(uuid.New()),
		UserID:           uuid.New(),
		IncludeCRA:       true,
		IncludeLeave:     true,
		IncludeTMA:       true,
		CanValidateLeave: true,
	})
	require.NoError(t, err)
	require.NotNil(t, home.CRA)
	assert.True(t, home.CRA.Required)
	assert.True(t, home.CRA.Alert)
	require.NotNil(t, home.Leave)
	assert.Equal(t, 1, home.Leave.Pending)
	assert.Equal(t, 1, home.Leave.PendingValidations)
	require.NotNil(t, home.TMA)
	assert.Equal(t, 2, home.TMA.Open)
	assert.Equal(t, 5, home.TMA.Total)
}
