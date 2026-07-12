package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeaveRequest_ApproveReject(t *testing.T) {
	today := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewDateRange(today.AddDate(0, 0, -2), today.AddDate(0, 0, 5))
	require.NoError(t, err)

	req := domain.NewLeaveRequest(kernel.NewTenantID(uuid.New()), uuid.New(), domain.LeaveTypeCongesPayes, period, "vacances")
	decider := uuid.New()

	require.NoError(t, req.Approve(today, decider))
	assert.Equal(t, domain.LeaveStatusApproved, req.Status)
	assert.Equal(t, &decider, req.DecidedBy)

	err = req.Approve(today, decider)
	assert.ErrorIs(t, err, domain.ErrLeaveAlreadyDecided)
}

func TestFutureDays_OnlyStrictlyFuture(t *testing.T) {
	today := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewDateRange(today.AddDate(0, 0, -2), today.AddDate(0, 0, 2))
	require.NoError(t, err)

	days := domain.FutureDays(period, today)
	assert.Equal(t, []time.Time{
		today.AddDate(0, 0, 1),
		today.AddDate(0, 0, 2),
	}, days)
}
