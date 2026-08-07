package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reminderRepo struct {
	fakeCRARepo
	candidates []domain.ReminderCandidate
}

func (r *reminderRepo) ListReminderCandidatesByMonth(context.Context, kernel.TenantID, domain.Month) ([]domain.ReminderCandidate, error) {
	return r.candidates, nil
}

func TestSendMonthlyReminders_IncludesUsersWithoutTimesheet(t *testing.T) {
	missing := uuid.New()
	incomplete := uuid.New()
	done := uuid.New()
	repo := &reminderRepo{candidates: []domain.ReminderCandidate{
		{UserID: missing, HasTimesheet: false},
		{UserID: incomplete, HasTimesheet: true, Status: domain.StatusBrouillon, WeeksSubmitted: 1, WeeksTotal: 4, TotalMinutes: 480},
		{UserID: done, HasTimesheet: true, Status: domain.StatusDefinitif, WeeksSubmitted: 4, WeeksTotal: 4, TotalMinutes: 9600},
	}}
	svc := &Service{repo: repo}
	pending, err := svc.SendMonthlyReminders(context.Background(), kernel.NewTenantID(uuid.New()), "2026-08")
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{missing, incomplete}, pending)
}
