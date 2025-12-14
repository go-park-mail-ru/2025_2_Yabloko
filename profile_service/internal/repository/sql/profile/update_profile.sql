UPDATE account
SET
    name    = $1,
    phone   = $2,
    city_id = $3,
    address = $4::text,
    addresses_history = CASE
        WHEN $5::text IS NULL THEN
            CASE
                WHEN $4::text IS NULL OR btrim($4::text) = '' THEN
                    COALESCE(addresses_history, '[]'::jsonb)
                ELSE
                    CASE
                        WHEN jsonb_array_length(COALESCE(addresses_history, '[]'::jsonb)) = 0 THEN
                            to_jsonb(ARRAY[btrim($4::text)]::text[])
                        WHEN (COALESCE(addresses_history, '[]'::jsonb) ->> -1)
                             IS DISTINCT FROM btrim($4::text) THEN
                            COALESCE(addresses_history, '[]'::jsonb)
                            || to_jsonb(ARRAY[btrim($4::text)]::text[])
                        ELSE
                            COALESCE(addresses_history, '[]'::jsonb)
                    END
            END

        WHEN $5::text = 'null' THEN
            '[]'::jsonb

        ELSE
            $5::jsonb
    END,
    avatar_url = $6::text
WHERE id = $7::uuid;
