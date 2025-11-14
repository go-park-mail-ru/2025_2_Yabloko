SELECT o.id          as order_id,
       o.total_price as total,
       o.status      as status,
       o.created_at  as created_at,
       si.id         as store_item_id,
       i.name        as name,
       i.card_img    as card_img,
       oi.price      as price,
       oi.quantity   as quantity
FROM orders o
JOIN order_item oi on oi.order_id = o.id
JOIN store_item si on si.id = oi.store_item_id
JOIN item i on i.id = si.item_id
WHERE o.id = $1
ORDER BY oi.created_at;