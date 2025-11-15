SELECT
    id,
    order_id,
    yookassa_id,
    status,
    amount,
    currency,
    description,
    metadata,
    created_at,
    updated_at
FROM
    payment
WHERE
    order_id = $1
ORDER BY
    created_at DESC
LIMIT
    1