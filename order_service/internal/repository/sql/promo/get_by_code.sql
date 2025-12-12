SELECT
    p.id,
    p.code,
    p.relative_discount,
    p.absolute_discount,
    p.start_at,
    p.end_at,
    p.created_at,
    p.updated_at
FROM promocode p
LEFT JOIN promocode_account pa
    ON pa.promocode_id = p.id
    AND pa.user_id = $1::uuid
WHERE p.code   = $2
  AND p.start_at <= $3
  AND p.end_at   >= $3
  AND pa.id IS NULL