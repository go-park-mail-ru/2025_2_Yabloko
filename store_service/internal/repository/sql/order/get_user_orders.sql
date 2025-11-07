SELECT o.id          as id,
       o.status      as status,
       o.total_price as total,
       o.created_at  as created_at
FROM orders o
WHERE o.user_id = $1
ORDER BY o.created_at DESC;