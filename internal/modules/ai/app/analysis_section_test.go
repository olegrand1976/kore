package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/adapters/stub"
	"github.com/kore/kore/internal/modules/ai/app"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memChunkRepo struct {
	byDemand map[uuid.UUID][]domain.DocumentChunk
}

func (m *memChunkRepo) ReplaceSourceChunks(_ context.Context, chunks []domain.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	demandID := *chunks[0].DemandID
	sourceID := chunks[0].SourceID
	if m.byDemand == nil {
		m.byDemand = make(map[uuid.UUID][]domain.DocumentChunk)
	}
	kept := make([]domain.DocumentChunk, 0, len(m.byDemand[demandID]))
	for _, c := range m.byDemand[demandID] {
		if c.SourceID != sourceID {
			kept = append(kept, c)
		}
	}
	m.byDemand[demandID] = append(kept, chunks...)
	return nil
}

func (m *memChunkRepo) DeleteBySource(_ context.Context, _ kernel.TenantID, _ string, _ uuid.UUID) error {
	return nil
}

func (m *memChunkRepo) SearchSimilar(
	_ context.Context,
	_ kernel.TenantID,
	demandID uuid.UUID,
	_ []float32,
	limit int,
) ([]domain.DocumentChunk, error) {
	chunks := m.byDemand[demandID]
	if limit > 0 && len(chunks) > limit {
		chunks = chunks[:limit]
	}
	return chunks, nil
}

func (m *memChunkRepo) CountByDemand(_ context.Context, _ kernel.TenantID, demandID uuid.UUID) (int, error) {
	return len(m.byDemand[demandID]), nil
}

type fakeTMAReader struct {
	demand tmadomain.Demand
	err    error
}

func (f *fakeTMAReader) GetDemand(_ context.Context, _ kernel.TenantID, id uuid.UUID) (tmadomain.Demand, error) {
	if f.err != nil {
		return tmadomain.Demand{}, f.err
	}
	if f.demand.ID != uuid.Nil && f.demand.ID != id {
		return tmadomain.Demand{}, tmadomain.ErrDemandNotFound
	}
	d := f.demand
	d.ID = id
	return d, nil
}

func (f *fakeTMAReader) ListDemands(context.Context, kernel.TenantID, bool) ([]tmadomain.Demand, error) {
	return nil, nil
}

func (f *fakeTMAReader) GetAnalysis(context.Context, kernel.TenantID, uuid.UUID) (tmadomain.AnalysisDossier, error) {
	return tmadomain.AnalysisDossier{}, nil
}

func TestSuggestAnalysisSection_requiresPrompt(t *testing.T) {
	repo := &memRepo{capOn: true, tenantOn: true}
	provider := stub.NewProvider()
	tma := &fakeTMAReader{demand: tmadomain.Demand{Subject: "Export XML"}}
	svc := app.NewService(repo, provider, tma, nil, nil, nil, app.WithRAG(provider, &memChunkRepo{}))
	tenant := kernel.NewTenantID(uuid.New())

	_, err := svc.SuggestAnalysisSection(context.Background(), ports.AnalysisSectionCommand{
		TenantID: tenant,
		UserID:   uuid.New(),
		DemandID: uuid.New(),
		Section:  domain.AnalysisSectionFunctional,
		Prompt:   "  ",
	})
	assert.ErrorIs(t, err, domain.ErrEmptyAnalysisPrompt)
}

func TestSuggestAnalysisSection_demandNotFound(t *testing.T) {
	repo := &memRepo{capOn: true, tenantOn: true}
	provider := stub.NewProvider()
	tma := &fakeTMAReader{err: tmadomain.ErrDemandNotFound}
	svc := app.NewService(repo, provider, tma, nil, nil, nil, app.WithRAG(provider, &memChunkRepo{}))

	_, err := svc.SuggestAnalysisSection(context.Background(), ports.AnalysisSectionCommand{
		TenantID: kernel.NewTenantID(uuid.New()),
		UserID:   uuid.New(),
		DemandID: uuid.New(),
		Section:  domain.AnalysisSectionFunctional,
		Prompt:   "Impact",
	})
	assert.ErrorIs(t, err, domain.ErrAnalysisDemandNotFound)
}

func TestSuggestAnalysisSection_withRAG(t *testing.T) {
	repo := &memRepo{capOn: true, tenantOn: true}
	provider := stub.NewProvider()
	demandID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	chunks := &memChunkRepo{
		byDemand: map[uuid.UUID][]domain.DocumentChunk{
			demandID: {{
				ID:         uuid.New(),
				TenantID:   tenant,
				FileName:   "spec.pdf",
				ChunkIndex: 0,
				Content:    "Le module export XML échoue sur les gros fichiers.",
			}},
		},
	}
	tma := &fakeTMAReader{demand: tmadomain.Demand{ID: demandID, Subject: "Export XML"}}
	svc := app.NewService(repo, provider, tma, nil, nil, nil, app.WithRAG(provider, chunks))

	result, err := svc.SuggestAnalysisSection(context.Background(), ports.AnalysisSectionCommand{
		TenantID: tenant,
		UserID:   uuid.New(),
		DemandID: demandID,
		Section:  domain.AnalysisSectionFunctional,
		Prompt:   "Impact utilisateur",
		UseRAG:   true,
		Subject:  "Export XML",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Text)
	assert.Contains(t, result.Text, "Analyse fonctionnelle")
	assert.Contains(t, result.Text, "Impact utilisateur")
	assert.NotContains(t, result.Text, "===UNTRUSTED")
	assert.NotEqual(t, uuid.Nil, result.RequestID)
	assert.True(t, result.UsedDocuments)
	require.Len(t, result.Sources, 1)
	assert.Equal(t, "spec.pdf", result.Sources[0].FileName)
	assert.Len(t, repo.logs, 1)
	assert.Equal(t, stub.ModelName, repo.logs[0].Model)
}

func TestDemandDocumentContext_countsChunks(t *testing.T) {
	repo := &memRepo{capOn: true, tenantOn: true}
	provider := stub.NewProvider()
	demandID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	chunks := &memChunkRepo{
		byDemand: map[uuid.UUID][]domain.DocumentChunk{
			demandID: {{ID: uuid.New(), Content: "x"}},
		},
	}
	tma := &fakeTMAReader{demand: tmadomain.Demand{ID: demandID, Subject: "Export"}}
	svc := app.NewService(repo, provider, tma, nil, nil, nil, app.WithRAG(provider, chunks))

	res, err := svc.DemandDocumentContext(context.Background(), tenant, demandID)
	require.NoError(t, err)
	assert.Equal(t, 1, res.IndexedChunkCount)
	assert.True(t, res.HasIndexedDocuments)

	_, err = svc.DemandDocumentContext(context.Background(), tenant, uuid.New())
	require.ErrorIs(t, err, domain.ErrAnalysisDemandNotFound)
}
