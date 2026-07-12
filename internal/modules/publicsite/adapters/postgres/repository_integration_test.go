//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/publicsite/adapters/postgres"
	"github.com/kore/kore/internal/modules/publicsite/domain"
	"github.com/kore/kore/internal/modules/publicsite/ports"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/stretchr/testify/require"
)

func buildFilter(commercialID uuid.UUID, around time.Time) ports.SlotFilter {
	return ports.SlotFilter{
		CommercialID: &commercialID,
		From:         around.Add(-time.Hour),
		To:           around.Add(time.Hour),
	}
}

func TestPublicsite_AntiDoubleBooking(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	commercialID := uuid.New()
	start := time.Now().UTC().Add(24 * time.Hour)
	end := start.Add(30 * time.Minute)
	require.NoError(t, repo.SeedSlot(ctx, commercialID, start, end))

	slots, err := repo.ListAvailableSlots(ctx, buildFilter(commercialID, start))
	require.NoError(t, err)
	require.Len(t, slots, 1)
	slotID := slots[0].ID

	// First reservation succeeds.
	require.NoError(t, repo.ReserveSlot(ctx, slotID))

	// Second reservation on the same slot must be rejected.
	err = repo.ReserveSlot(ctx, slotID)
	require.ErrorIs(t, err, domain.ErrSlotAlreadyBooked)
}
