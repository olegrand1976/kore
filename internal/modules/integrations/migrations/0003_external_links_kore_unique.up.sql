-- One Kore demand ↔ one Taiga link per tenant (keep newest row when deduping).
DELETE FROM integrations.external_links
WHERE id IN (
    SELECT id
    FROM (
        SELECT id,
            ROW_NUMBER() OVER (
                PARTITION BY tenant_id, kore_entity_type, kore_entity_id
                ORDER BY updated_at DESC, created_at DESC
            ) AS rn
        FROM integrations.external_links
    ) ranked
    WHERE rn > 1
);

ALTER TABLE integrations.external_links
    ADD CONSTRAINT external_links_kore_entity_unique
    UNIQUE (tenant_id, kore_entity_type, kore_entity_id);
