DELETE FROM ai.ai_capabilities WHERE code = 'tma.analysis_section';

DROP INDEX IF EXISTS ai.document_chunks_embedding_hnsw_idx;
DROP INDEX IF EXISTS ai.document_chunks_source_idx;
DROP INDEX IF EXISTS ai.document_chunks_tenant_demand_idx;
DROP TABLE IF EXISTS ai.document_chunks;
