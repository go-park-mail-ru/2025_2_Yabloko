SELECT DISTINCT ON (o.id)
    o.id,
    o.status,
    o.total_price,
    o.created_at,
    o.is_fast,
    o.comment,
    s.id as store_id,
    s.name as store_name
FROM orders o
JOIN order_item oi ON oi.order_id = o.id
JOIN store_item si ON si.id = oi.store_item_id
JOIN store s ON s.id = si.store_id
WHERE o.user_id = $1::uuid
AND (
    $2::uuid IS NULL
    OR o.created_at < (
        SELECT created_at FROM orders WHERE id = $2::uuid
    )
)
ORDER BY o.id, o.created_at DESC, s.name
LIMIT $3;
