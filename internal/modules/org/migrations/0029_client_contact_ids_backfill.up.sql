-- Backfill stable UUIDs on org.clients.contacts entries missing a usable id.
UPDATE org.clients AS c
SET contacts = COALESCE((
    SELECT jsonb_agg(
        CASE
            WHEN (elem ? 'id')
                AND NULLIF(TRIM(elem->>'id'), '') IS NOT NULL
                AND TRIM(elem->>'id') <> '00000000-0000-0000-0000-000000000000'
            THEN elem
            ELSE elem || jsonb_build_object('id', gen_random_uuid()::text)
        END
        ORDER BY ordinality
    )
    FROM jsonb_array_elements(COALESCE(c.contacts, '[]'::jsonb)) WITH ORDINALITY AS t(elem, ordinality)
), '[]'::jsonb)
WHERE jsonb_typeof(COALESCE(c.contacts, '[]'::jsonb)) = 'array';
