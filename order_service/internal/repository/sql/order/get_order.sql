SELECT
    o.id as order_id,
    o.total_price as total,
    o.status as status,
    o.created_at as created_at,
    s.id as store_id,
    s.name as store_name,
    si.id as store_item_id,
    i.name as name,
    i.card_img as card_img,
    oi.price as price,
    oi.quantity as quantity
FROM
    orders o
    JOIN order_item oi ON oi.order_id = o.id
    JOIN store_item si ON si.id = oi.store_item_id
    JOIN item i ON i.id = si.item_id
    JOIN store s ON s.id = si.store_id
WHERE
    o.id = $1::uuid
ORDER BY
    oi.created_at;