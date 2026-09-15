package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
	"github.com/kore/kore/pkg/kernel"
)

func (s *Service) IndexRequestAttachment(ctx context.Context, cmd ports.IndexAttachmentCommand) error {
	if s.embed == nil || s.chunks == nil {
		return nil
	}
	text, err := ExtractAttachmentText(cmd.StoragePath, cmd.FileName)
	if err != nil {
		return err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		_ = s.chunks.DeleteBySource(ctx, cmd.TenantID, domain.SourceTypeRequestAttachment, cmd.AttachmentID)
		return nil
	}
	parts := SplitTextChunks(text)
	if len(parts) == 0 {
		return nil
	}
	vectors, err := s.embed.Embed(ctx, parts)
	if err != nil {
		return err
	}
	if len(vectors) != len(parts) {
		return fmt.Errorf("embed count mismatch")
	}
	demandID := cmd.DemandID
	now := time.Now().UTC()
	chunks := make([]domain.DocumentChunk, 0, len(parts))
	for i, content := range parts {
		chunks = append(chunks, domain.DocumentChunk{
			ID:         uuid.New(),
			TenantID:   cmd.TenantID,
			SourceType: domain.SourceTypeRequestAttachment,
			SourceID:   cmd.AttachmentID,
			DemandID:   &demandID,
			ChunkIndex: i,
			Content:    content,
			Embedding:  vectors[i],
			MimeType:   cmd.MimeType,
			FileName:   cmd.FileName,
			CreatedAt:  now,
		})
	}
	return s.chunks.ReplaceSourceChunks(ctx, chunks)
}

func (s *Service) RemoveRequestAttachmentChunks(ctx context.Context, tenant kernel.TenantID, attachmentID uuid.UUID) error {
	if s.chunks == nil {
		return nil
	}
	return s.chunks.DeleteBySource(ctx, tenant, domain.SourceTypeRequestAttachment, attachmentID)
}

func (s *Service) retrieveRAG(
	ctx context.Context,
	tenant kernel.TenantID,
	demandID uuid.UUID,
	query string,
	limit int,
) ([]domain.DocumentChunk, error) {
	if s.embed == nil || s.chunks == nil {
		return nil, nil
	}
	vecs, err := s.embed.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, nil
	}
	return s.chunks.SearchSimilar(ctx, tenant, demandID, vecs[0], limit)
}

func formatRAGContext(chunks []domain.DocumentChunk) string {
	if len(chunks) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Extraits documentaires (RAG) :\n")
	for i, c := range chunks {
		fmt.Fprintf(&b, "[%d] %s#%d\n%s\n\n", i+1, c.FileName, c.ChunkIndex, c.Content)
	}
	return b.String()
}

func ragSources(chunks []domain.DocumentChunk) []domain.RAGSource {
	out := make([]domain.RAGSource, 0, len(chunks))
	for _, c := range chunks {
		out = append(out, domain.RAGSource{
			FileName:   c.FileName,
			ChunkIndex: c.ChunkIndex,
			Score:      c.Score,
		})
	}
	return out
}
