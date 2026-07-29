package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/maintenance/app"
	"github.com/kore/kore/internal/modules/maintenance/domain"
	"github.com/kore/kore/internal/modules/maintenance/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type fakeMaintRepo struct {
	items map[uuid.UUID]domain.WorkRequest
}

func newFakeMaintRepo() *fakeMaintRepo {
	return &fakeMaintRepo{items: map[uuid.UUID]domain.WorkRequest{}}
}

func (f *fakeMaintRepo) ListWorkRequests(_ context.Context, _ kernel.TenantID) ([]domain.WorkRequest, error) {
	out := make([]domain.WorkRequest, 0, len(f.items))
	for _, wr := range f.items {
		out = append(out, wr)
	}
	return out, nil
}

func (f *fakeMaintRepo) GetWorkRequest(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.WorkRequest, error) {
	wr, ok := f.items[id]
	if !ok {
		return domain.WorkRequest{}, domain.ErrWorkRequestNotFound
	}
	return wr, nil
}

func (f *fakeMaintRepo) SaveWorkRequest(_ context.Context, wr domain.WorkRequest) error {
	f.items[wr.ID] = wr
	return nil
}

type noopFeeder struct{}

func (noopFeeder) ProposeLines(context.Context, []ports.ProposedLine) error { return nil }

func TestMaintenanceService_CreateAssignProgressComplete(t *testing.T) {
	repo := newFakeMaintRepo()
	svc := app.NewService(repo, noopFeeder{})
	tenant := kernel.NewTenantID(uuid.New())
	ctx := context.Background()

	wr, err := svc.Create(ctx, ports.CreateWorkRequestCommand{
		TenantID:      tenant,
		ApplicationID: uuid.New(),
		Subject:       "Patch",
		Description:   "desc",
		Priority:      "normal",
	})
	require.NoError(t, err)
	require.Equal(t, domain.WorkStateCreated, wr.State)

	assignee := uuid.New()
	wr, err = svc.Assign(ctx, ports.AssignCommand{TenantID: tenant, RequestID: wr.ID, AssigneeID: assignee})
	require.NoError(t, err)
	require.Equal(t, domain.WorkStateAssigned, wr.State)

	wr, err = svc.Progress(ctx, ports.ProgressCommand{TenantID: tenant, RequestID: wr.ID, ConsumptionDays: 2})
	require.NoError(t, err)
	require.Equal(t, domain.WorkStateInProgress, wr.State)

	wr, err = svc.Complete(ctx, tenant, wr.ID)
	require.NoError(t, err)
	require.Equal(t, domain.WorkStateCompleted, wr.State)
}
