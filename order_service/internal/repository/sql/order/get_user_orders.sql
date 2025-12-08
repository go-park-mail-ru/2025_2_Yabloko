SELECT 
    o.id,
    o.status,
    o.total_price,
    o.created_at,
    s.id,
    s.name
FROM orders o
JOIN order_item oi ON oi.order_id = o.id
JOIN store s ON s.id = oi.store_id
WHERE o.user_id = $1::uuid
AND (
    $2::uuid IS NULL
    OR o.created_at < (
        SELECT created_at FROM orders WHERE id = $2::uuid
    )
)
GROUP BY o.id, s.id, s.name
ORDER BY o.created_at DESC
LIMIT $3;
