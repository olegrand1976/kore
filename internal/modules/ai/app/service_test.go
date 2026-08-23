package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/modules/ai/app"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memRepo struct {
	capOn    bool
	logs     []domain.RequestLog
	tenantOn bool
}

func (m *memRepo) IsCapabilityEnabled(_ context.Context, _ string) (bool, error) {
	return m.capOn, nil
}

func (m *memRepo) GetTenantSettings(_ context.Context, tenant kernel.TenantID) (domain.TenantSettings, error) {
	return domain.TenantSettings{TenantID: tenant, Enabled: m.tenantOn, LLMProvider: "stub"}, nil
}

func (m *memRepo) UpsertTenantSettings(_ context.Context, settings domain.TenantSettings) error {
	m.tenantOn = settings.Enabled
	return nil
}

func (m *memRepo) InsertRequestLog(_ context.Context, log domain.RequestLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *memRepo) GetRequestLog(_ context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.RequestLog, error) {
	for _, l := range m.logs {
		if l.ID == id && l.TenantID == tenant {
			return l, nil
		}
	}
	return domain.RequestLog{}, domain.ErrRequestNotFound
}

func TestSuggestAnalysisDraft(t *testing.T) {
	repo := &memRepo{capOn: true, tenantOn: true}
	svc := app.NewService(repo, stub.NewProvider(), nil, nil, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()

	result, err := svc.SuggestAnalysisDraft(context.Background(), ports.AnalysisDraftCommand{
		TenantID: tenant,
		UserID:   userID,
		Subject:  "Erreur export XML",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Draft.Functional)
	assert.NotEqual(t, uuid.Nil, result.RequestID)
	assert.Len(t, repo.logs, 1)
}

func TestClassifyDemand(t *testing.T) {
	category, confidence := stub.ClassifySubject("régression module login")
	assert.Equal(t, "regression", category)
	assert.Greater(t, confidence, 0.5)
}

func TestAIDisabled(t *testing.T) {
	repo := &memRepo{capOn: true, tenantOn: false}
	svc := app.NewService(repo, stub.NewProvider(), nil, nil, nil, nil)
	_, err := svc.SuggestAnalysisDraft(context.Background(), ports.AnalysisDraftCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		Subject:  "test",
	})
	assert.ErrorIs(t, err, domain.ErrAIDisabled)
}

type fakeCRAReader struct {
	timesheet cradomain.Timesheet
}

func (f *fakeCRAReader) GetTimesheetByID(_ context.Context, _ kernel.TenantID, _ uuid.UUID) (cradomain.Timesheet, error) {
	return f.timesheet, nil
}

func (f *fakeCRAReader) ListRecentTimesheets(_ context.Context, _ kernel.TenantID, _ uuid.UUID, _ int) ([]cradomain.Timesheet, error) {
	return nil, nil
}

func TestListCraAnomaliesWithoutTenantAI(t *testing.T) {
	repo := &memRepo{capOn: true, tenantOn: false}
	tsID := uuid.New()
	day := time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC)
	cra := &fakeCRAReader{timesheet: cradomain.Timesheet{
		ID:     tsID,
		Status: cradomain.StatusBrouillon,
		Weeks: []cradomain.WeekEntry{{
			WeekNumber: 5,
			Lines: []cradomain.TimeLine{{
				Day:      day,
				Duration: kernel.Duration{Minutes: 600},
			}},
		}},
	}}
	svc := app.NewService(repo, stub.NewProvider(), nil, cra, nil, nil)
	tenant := kernel.NewTenantID(uuid.New())

	anomalies, err := svc.ListCraAnomalies(context.Background(), ports.CraAnomaliesCommand{
		TenantID: tenant, UserID: uuid.New(), TimesheetID: tsID,
	})
	require.NoError(t, err)
	require.Len(t, anomalies, 1)
	assert.Equal(t, "DAY_CAPACITY", anomalies[0].Code)
}
