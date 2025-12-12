SELECT
    o.id as order_id,
    o.total_price as total,
    o.status as status,
    o.is_fast as is_fast,
    o.comment as comment,
    o.created_at as created_at,
    s.id as store_id,
    s.name as store_name,
    s.card_img as store_card_img,
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
    s.id, oi.created_at;
