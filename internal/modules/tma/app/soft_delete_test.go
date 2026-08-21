package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

func TestSoftDelete_MarksDemand(t *testing.T) {
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	demand := domain.NewDemand(tenant, uuid.New(), uuid.New(), "to delete", "", kernel.PriorityNormal, nil, false)
	at := time.Date(2026, 8, 21, 15, 0, 0, 0, time.UTC)

	repo := &fakeDemandRepo{demand: demand}
	svc := NewService(repo, nil, nil, WithClock(fixedClock{at: at}))

	require.NoError(t, svc.SoftDelete(ctx, tenant, demand.ID))
	require.NotNil(t, repo.lastDeletedAt)
	require.True(t, repo.lastDeletedAt.Equal(at))
}

func TestSoftDelete_NotFound(t *testing.T) {
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	repo := &fakeDemandRepo{getErr: domain.ErrDemandNotFound}
	svc := NewService(repo, nil, nil)

	err := svc.SoftDelete(ctx, tenant, uuid.New())
	require.ErrorIs(t, err, domain.ErrDemandNotFound)
}
