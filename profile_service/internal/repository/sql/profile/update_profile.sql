UPDATE account
SET
    name    = $1,
    phone   = $2,
    city_id = $3,
    address = $4,
    addresses_history = CASE
        WHEN $5::jsonb IS NOT NULL THEN
            $5::jsonb
        ELSE
            CASE
                WHEN $4 IS NULL OR btrim($4) = '' THEN
                    COALESCE(addresses_history, '[]'::jsonb)
                ELSE
                    CASE
                        WHEN jsonb_array_length(COALESCE(addresses_history, '[]'::jsonb)) = 0 THEN
                            to_jsonb(ARRAY[btrim($4)]::text[])
                        ELSE
                            CASE
                                WHEN (COALESCE(addresses_history, '[]'::jsonb) ->> -1)
                                      IS DISTINCT FROM btrim($4) THEN
                                    COALESCE(addresses_history, '[]'::jsonb)
                                    || to_jsonb(ARRAY[btrim($4)]::text[])
                                ELSE
                                    COALESCE(addresses_history, '[]'::jsonb)
                            END
                    END
            END
    END,
    avatar_url = $6
WHERE id = $7;
