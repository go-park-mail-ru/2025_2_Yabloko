UPDATE
    orders
SET
    total_price = (
        SELECT
            COALESCE(SUM(oi.price * oi.quantity), 0)
        FROM
            order_item oi
        WHERE
            oi.order_id = $1
    )
WHERE
    id = $1;