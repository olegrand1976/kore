CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE ai.document_chunks (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    source_type TEXT NOT NULL,
    source_id UUID NOT NULL,
    demand_id UUID,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768) NOT NULL,
    mime_type TEXT NOT NULL DEFAULT '',
    file_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT document_chunks_source_chunk_uq UNIQUE (source_type, source_id, chunk_index)
);

CREATE INDEX document_chunks_tenant_demand_idx
    ON ai.document_chunks (tenant_id, demand_id);

CREATE INDEX document_chunks_source_idx
    ON ai.document_chunks (source_type, source_id);

CREATE INDEX document_chunks_embedding_hnsw_idx
    ON ai.document_chunks
    USING hnsw (embedding vector_cosine_ops);

INSERT INTO ai.ai_capabilities (code, risk_class, annex_iii, wave) VALUES
    ('tma.analysis_section', 'limited', FALSE, 1)
ON CONFLICT (code) DO NOTHING;
