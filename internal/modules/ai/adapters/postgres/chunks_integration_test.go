//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/adapters/postgres"
	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/platform/db/dbtest"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestDocumentChunks_replaceAndSearch(t *testing.T) {
	pool := dbtest.NewPostgres(t)
	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()
	sourceID := uuid.New()
	vec := stub.DeterministicEmbedding("contenu chunk test")

	chunk := domain.DocumentChunk{
		ID:         uuid.New(),
		TenantID:   tenant,
		SourceType: domain.SourceTypeRequestAttachment,
		SourceID:   sourceID,
		DemandID:   &demandID,
		ChunkIndex: 0,
		Content:    "contenu chunk test",
		Embedding:  vec,
		FileName:   "notes.txt",
		CreatedAt:  time.Now().UTC(),
	}
	require.NoError(t, repo.ReplaceSourceChunks(ctx, []domain.DocumentChunk{chunk}))

	n, err := repo.CountByDemand(ctx, tenant, demandID)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	found, err := repo.SearchSimilar(ctx, tenant, demandID, vec, 3)
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "notes.txt", found[0].FileName)

	require.NoError(t, repo.DeleteBySource(ctx, tenant, domain.SourceTypeRequestAttachment, sourceID))
	n, err = repo.CountByDemand(ctx, tenant, demandID)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
