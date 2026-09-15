package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/pkg/kernel"
)

func formatVector(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%g", f)
	}
	b.WriteByte(']')
	return b.String()
}

func (r *Repository) ReplaceSourceChunks(ctx context.Context, chunks []domain.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	first := chunks[0]
	if _, err := tx.Exec(ctx, `
		DELETE FROM ai.document_chunks
		WHERE tenant_id = $1 AND source_type = $2 AND source_id = $3`,
		first.TenantID.UUID(), first.SourceType, first.SourceID,
	); err != nil {
		return err
	}

	for _, c := range chunks {
		if len(c.Embedding) != domain.EmbeddingDimensions {
			return fmt.Errorf("embedding dim %d want %d", len(c.Embedding), domain.EmbeddingDimensions)
		}
		created := c.CreatedAt
		if created.IsZero() {
			created = time.Now().UTC()
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ai.document_chunks (
				id, tenant_id, source_type, source_id, demand_id, chunk_index,
				content, embedding, mime_type, file_name, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8::vector,$9,$10,$11)`,
			c.ID, c.TenantID.UUID(), c.SourceType, c.SourceID, c.DemandID, c.ChunkIndex,
			c.Content, formatVector(c.Embedding), c.MimeType, c.FileName, created,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) DeleteBySource(ctx context.Context, tenant kernel.TenantID, sourceType string, sourceID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM ai.document_chunks
		WHERE tenant_id = $1 AND source_type = $2 AND source_id = $3`,
		tenant.UUID(), sourceType, sourceID,
	)
	return err
}

func (r *Repository) SearchSimilar(
	ctx context.Context,
	tenant kernel.TenantID,
	demandID uuid.UUID,
	query []float32,
	limit int,
) ([]domain.DocumentChunk, error) {
	if limit <= 0 {
		limit = 6
	}
	if len(query) != domain.EmbeddingDimensions {
		return nil, fmt.Errorf("query embedding dim %d want %d", len(query), domain.EmbeddingDimensions)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, source_type, source_id, demand_id, chunk_index,
		       content, mime_type, file_name, created_at,
		       1 - (embedding <=> $3::vector) AS score
		FROM ai.document_chunks
		WHERE tenant_id = $1 AND demand_id = $2
		ORDER BY embedding <=> $3::vector
		LIMIT $4`,
		tenant.UUID(), demandID, formatVector(query), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.DocumentChunk
	for rows.Next() {
		var c domain.DocumentChunk
		var tenantUUID uuid.UUID
		var demand *uuid.UUID
		if err := rows.Scan(
			&c.ID, &tenantUUID, &c.SourceType, &c.SourceID, &demand, &c.ChunkIndex,
			&c.Content, &c.MimeType, &c.FileName, &c.CreatedAt, &c.Score,
		); err != nil {
			return nil, err
		}
		c.TenantID = kernel.NewTenantID(tenantUUID)
		c.DemandID = demand
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) CountByDemand(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ai.document_chunks
		WHERE tenant_id = $1 AND demand_id = $2`,
		tenant.UUID(), demandID,
	).Scan(&n)
	return n, err
}
